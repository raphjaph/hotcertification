package main

// copied from server.go in hoststuff/cmd/hoststuffserver
import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/relab/gorums"
	"github.com/relab/hotstuff"
	hotstuffbackend "github.com/relab/hotstuff/backend/gorums"
	"github.com/relab/hotstuff/config"
	"github.com/relab/hotstuff/consensus/chainedhotstuff"
	"github.com/relab/hotstuff/leaderrotation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/proto"

	"github.com/raphasch/hotcertification/logging"
	"github.com/raphasch/hotcertification/protocol"
)

// reqID is a unique identifier for a command
// TODO: use fingerprint of
type reqID struct {
	clientID    uint64
	sequenceNum uint64
}

// This implements the Certification interface from the certification_gorums.pb.go (protocol.Certification)
// This struct holds all the data/variables a certificationServer needs
type certificationServer struct {
	ctx    context.Context
	cancel context.CancelFunc
	conf   *options
	cfg    *hotstuffbackend.Config // the group of peers/nodes in this distributed certification authority

	gorumsSrv *gorums.Server // transport backend for the client facing interaction

	hs    *hotstuff.HotStuff      // the byzantine fault tolerant replication (consensus) algorithm
	hsSrv *hotstuffbackend.Server // the transport backend for the consensus algorithm

	pm        hotstuff.ViewSynchronizer // the heart beat of the system (synchronizer/clock)
	reqBuffer *reqBuffer                // the request buffer (CSRs)

	mut          sync.Mutex
	finishedReqs map[reqID]chan struct{} // signalling channel

	lastExecTime int64
}

func newCertificationServer(conf *options, replicaConfig *config.ReplicaConfig, tlsCert *tls.Certificate) *certificationServer {
	ctx, cancel := context.WithCancel(context.Background())

	serverOpts := []gorums.ServerOption{}
	grpcServerOpts := []grpc.ServerOption{}

	if conf.TLS {
		grpcServerOpts = append(grpcServerOpts, grpc.Creds(credentials.NewServerTLSFromCert(tlsCert)))
	}

	serverOpts = append(serverOpts, gorums.WithGRPCServerOptions(grpcServerOpts...))

	// the Gorums Server (Connection Manager); de- and serialization, tls, etc.
	// connect Certification Server with the gorums backend object
	certSrv := &certificationServer{
		ctx:          ctx,
		cancel:       cancel,
		conf:         conf,
		gorumsSrv:    gorums.NewServer(serverOpts...),
		reqBuffer:    newReqBuffer(conf.BatchSize),
		finishedReqs: make(map[reqID]chan struct{}),
		lastExecTime: time.Now().UnixNano(),
	}

	builder := chainedhotstuff.DefaultModules(
		*replicaConfig,
		hotstuff.ExponentialTimeout{Base: time.Duration(conf.ViewTimeout) * time.Millisecond, ExponentBase: 2, MaxExponent: 8},
	)
	certSrv.cfg = hotstuffbackend.NewConfig(*replicaConfig)
	certSrv.hsSrv = hotstuffbackend.NewServer(*replicaConfig)
	builder.Register(certSrv.cfg, certSrv.hsSrv)

	var leaderRotation hotstuff.LeaderRotation
	switch conf.PmType {
	case "fixed":
		leaderRotation = leaderrotation.NewFixed(conf.LeaderID)
	case "round-robin":
		leaderRotation = leaderrotation.NewRoundRobin()
	default:
		fmt.Fprintf(os.Stderr, "Invalid pacemaker type: '%s'\n", conf.PmType)
		os.Exit(1)
	}
	builder.Register(
		leaderRotation,
		certSrv,           // executor
		certSrv.reqBuffer, // acceptor and request buffer
		logging.New(fmt.Sprintf("hs%d", conf.SelfID)),
	)
	certSrv.hs = builder.Build()

	// Use a custom server instead of the gorums one
	protocol.RegisterCertificationServer(certSrv.gorumsSrv, certSrv)
	return certSrv
}

func (srv *certificationServer) Start(address string) (err error) {

	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	fmt.Println(address)
	err = srv.hsSrv.Start()
	if err != nil {
		return err
	}

	err = srv.cfg.Connect(10 * time.Second)
	if err != nil {
		return err
	}

	// sleep so that all replicas can be ready before we start
	time.Sleep(time.Second)

	go srv.hs.EventLoop().Run(srv.ctx)

	go func() {
		err := srv.gorumsSrv.Serve(lis)
		if err != nil {
			log.Println(err)
		}
	}()

	return nil
}

func (srv *certificationServer) Stop() {
	srv.hs.ViewSynchronizer().Stop()
	srv.cfg.Close()
	srv.hsSrv.Stop()
	srv.gorumsSrv.Stop()
	srv.cancel()
}

func (certSrv *certificationServer) GetCertificate(_ context.Context, req *protocol.Request, out func(*protocol.Reply, error)) {
	// channel for signalling purpose (check if command done with consensus phase) and synchronization
	finished := make(chan struct{})
	id := reqID{req.ClientID, req.SequenceNumber}
	certSrv.mut.Lock()
	certSrv.finishedReqs[id] = finished
	certSrv.mut.Unlock()

	// reqBuffer processes the commands before passing them on to the Hotstuff Core
	certSrv.reqBuffer.addRequest(req)

	// start a go routine to make asynchronous; that's what the option (gorums.async) = true is for!
	go func(id reqID, finished chan struct{}) {
		// executes following code when command has been processed by hotstuff consensus
		// blocks until other process signals that consensus for that command is finished
		<-finished

		certSrv.mut.Lock()
		// delete signalling channel for specific command
		delete(certSrv.finishedReqs, id)
		certSrv.mut.Unlock()

		// send response
		out(&protocol.Reply{Data: "Partially Signed Certificate from server " + fmt.Sprint(certSrv.cfg.ID())}, nil)
	}(id, finished)
}

// implementing the Executor interface
// HS works on strings so Command can be a batch of seperate commands
func (certSrv *certificationServer) Exec(cmd hotstuff.Command) {
	batch := new(protocol.Batch)
	// unmarshalling infuses the string with meaning by seperating different commands
	err := proto.UnmarshalOptions{AllowPartial: true}.Unmarshal([]byte(cmd), batch)
	if err != nil {
		return
	}

	// batch of commands are processes and executed
	for _, req := range batch.GetRequests() {
		// ?
		if err != nil {
			log.Printf("Failed to unmarshal command: %v\n", err)
		}

		// IMPORTANT: This is where certificates are partially signed and stored in a database
		fmt.Printf("Server %v: partially signing certificate for client %v\n", certSrv.conf.SelfID, req.ClientID)

		// signal to GetCertificate() that response can be sent out
		certSrv.mut.Lock()
		if finishedRequestChannel, ok := certSrv.finishedReqs[reqID{req.ClientID, req.SequenceNumber}]; ok {
			finishedRequestChannel <- struct{}{}
		}
		certSrv.mut.Unlock()
	}
}
