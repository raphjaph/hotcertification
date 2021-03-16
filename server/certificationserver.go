package main

// copied from server.go in hoststuff/cmd/hoststuffserver
import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/relab/gorums"
	"github.com/relab/hotstuff"
	hotstuffbackend "github.com/relab/hotstuff/backend/gorums"
	"github.com/relab/hotstuff/config"
	"github.com/relab/hotstuff/consensus/chainedhotstuff"
	"github.com/relab/hotstuff/leaderrotation"
	"github.com/relab/hotstuff/synchronizer"
	"google.golang.org/protobuf/proto"

	"github.com/raphasch/hotcertification/protocol"
	"github.com/raphasch/hotvertification/internal/crypto"
)

// cmdID is a unique identifier for a command
// TODO: clientID -> long-term public key of client or make unique identifier a hash of the data in a command salted with date time?
type cmdID struct {
	clientID    uint32
	sequenceNum uint64
}

// This implements the Certification interface from the certification_gorums.pb.go (protocol.Certification)
// This struct holds all the data/variables a certificationServer needs
type certificationServer struct {
	ctx       context.Context
	cancel    context.CancelFunc
	conf      *options
	gorumsSrv *gorums.Server
	hsSrv     *hotstuffbackend.Server
	cfg       *hotstuffbackend.Config
	hs        hotstuff.Consensus
	pm        hotstuff.ViewSynchronizer
	reqCache  *reqCache

	mut          sync.Mutex
	finishedCmds map[cmdID]chan struct{}

	lastExecTime int64
}

// TODO: add TLS
//func newCertificationServer(conf *options, replicaConfig *config.ReplicaConfig, tlsCert *tls.Certificate) *certificationServer
func newCertificationServer(conf *options, replicaConfig *config.ReplicaConfig) *certificationServer {
	ctx, cancel := context.WithCancel(context.Background())

	/* serverOpts := []protocol.ServerOption{}
	grpcServerOpts := []grpc.ServerOption{}
	// server options if needed; if TLS see other implementation
	serverOpts = append(serverOpts, protocol.WithGRPCServerOptions(grpcServerOpts...)) */

	// the Gorums Server (Connection Manager); de- and serialization, tls, etc.
	// connect Certification Server with the gorums backend object
	// TODO: call this backend server instead?
	gorumsSrv := gorums.NewServer()
	certSrv := &certificationServer{
		ctx:          ctx,
		cancel:       cancel,
		conf:         conf,
		gorumsSrv:    gorumsSrv,
		reqCache:     newReqCache(conf.BatchSize),
		finishedCmds: make(map[cmdID]chan struct{}),
		lastExecTime: time.Now().UnixNano(),
	}

	var err error
	certSrv.cfg = hotstuffbackend.NewConfig(*replicaConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init gorums backend: %s\n", err)
		os.Exit(1)
	}

	certSrv.hsSrv = hotstuffbackend.NewServer(*replicaConfig)

	// just fixed Leader for now
	leaderRotation := leaderrotation.NewFixed(conf.LeaderID)
	certSrv.pm = synchronizer.New(leaderRotation, time.Duration(conf.ViewTimeout)*time.Millisecond)
	certSrv.hs = chainedhotstuff.Builder{
		Config:       certSrv.cfg,
		Acceptor:     certSrv.reqCache,
		Executor:     certSrv,
		Synchronizer: certSrv.pm,
		CommandQueue: certSrv.reqCache,
	}.Build()

	protocol.RegisterCertificationServer(gorumsSrv, certSrv)
	return certSrv
}

func (certSrv *certificationServer) GetCertificate(_ context.Context, req *protocol.Request, out func(*protocol.Reply, error)) {
	// channel for signalling purpose (check if command done with consensus phase) and synchronization
	finished := make(chan struct{})
	id := cmdID{req.ClientID, req.SequenceNumber}
	certSrv.mut.Lock()
	certSrv.finishedCmds[id] = finished
	certSrv.mut.Unlock()

	// reqCache processes the commands before passing them on to the Hotstuff Core
	certSrv.reqCache.addRequest(req)

	// start a go routine to make asynchronous; that's what the option (gorums.async) = true is for!
	go func(id cmdID, finished chan struct{}) {
		// executes following code when command has been processed by hotstuff consensus
		// blocks until other process signals that consensus for that command is finished
		<-finished

		certSrv.mut.Lock()
		// delete signalling channel for specific command
		delete(certSrv.finishedCmds, id)
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
		if finishedRequestChannel, ok := certSrv.finishedCmds[cmdID{req.ClientID, req.SequenceNumber}]; ok {
			finishedRequestChannel <- struct{}{}
		}
		certSrv.mut.Unlock()
	}
}
