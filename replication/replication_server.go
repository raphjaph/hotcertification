/*
	REPLICATION LOGIC
		1. Processes the requests(CSRs) from request buffer
		2. Request buffer validates with Accept() function
		3. Validated request is replicated (and also validated) by other nodes
		4. OnExec() function (when replication done) writes request to a database
		5. finishedReqs channel signals signing go routine to start threshold signature process


*/
package replication

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/relab/hotstuff"
	hotstuffbackend "github.com/relab/hotstuff/backend/gorums"
	hsconfig "github.com/relab/hotstuff/config"
	"github.com/relab/hotstuff/consensus/chainedhotstuff"
	"github.com/relab/hotstuff/crypto"
	"github.com/relab/hotstuff/crypto/ecdsa"
	"github.com/relab/hotstuff/crypto/keygen"
	"github.com/relab/hotstuff/leaderrotation"

	hc "github.com/raphasch/hotcertification"
	"github.com/raphasch/hotcertification/logging"
	"github.com/raphasch/hotcertification/options"
)

// reqID is a unique identifier for a command
// TODO: use fingerprint of
type reqID struct {
	clientID    uint64
	sequenceNum uint64
}

// This implements the Certification interface from the certification_gorums.pb.go (protocol.Certification)
// This struct holds all the data/variables a certificationServer needs
type replicationServer struct {
	//conf   *options
	hs             *hotstuff.HotStuff       // the byzantine fault tolerant replication (consensus) algorithm
	hsSrv          *hotstuffbackend.Server  // the transport backend for the consensus algorithm
	mgr            *hotstuffbackend.Manager // manages the connections to the other peers/replicas in the network
	ReqBuffer      *hc.ReqBuffer            // the request buffer (CSRs); passed in from client server
	mut            sync.Mutex
	replicatedReqs chan struct{} // TODO: change from struct{} to *client.CSR or *x509.CertificateRequest. Can I put anything into a chan struct{} and then transform at the other end through reflection
	//finishedReqs map[reqID]chan struct{} // signalling channel

	lastExecTime int64
}

func NewReplicationServer(opts *options.Options, reqBuffer *hc.ReqBuffer, replicatedRequests chan struct{}) *replicationServer {

	replicaConfig := createReplicaConfig(opts)

	srv := &replicationServer{
		ReqBuffer:      reqBuffer,
		replicatedReqs: replicatedRequests,
	}

	// building the hotstuff consensus algorithm
	builder := chainedhotstuff.DefaultModules(
		*replicaConfig,
		hotstuff.ExponentialTimeout{Base: time.Duration(opts.ViewTimeout) * time.Millisecond, ExponentBase: 2, MaxExponent: 8},
	)
	srv.mgr = hotstuffbackend.NewManager(*replicaConfig)
	srv.hsSrv = hotstuffbackend.NewServer(*replicaConfig)
	builder.Register(srv.mgr, srv.hsSrv)

	var leaderRotation hotstuff.LeaderRotation
	switch opts.PmType {
	case "fixed":
		leaderRotation = leaderrotation.NewFixed(opts.LeaderID)
	case "round-robin":
		// assumes IDs start at 1
		leaderRotation = leaderrotation.NewRoundRobin()
	default:
		fmt.Fprintf(os.Stderr, "Invalid pacemaker type: '%s'\n", opts.PmType)
		os.Exit(1)
	}
	var consensus hotstuff.Consensus
	consensus = chainedhotstuff.New()

	var cryptoImpl hotstuff.CryptoImpl
	cryptoImpl = ecdsa.New()

	builder.Register(
		consensus,
		crypto.NewCache(cryptoImpl, 2*srv.mgr.Len()),
		leaderRotation,
		srv,           // executor
		srv.ReqBuffer, // acceptor and command queue
		logging.New(fmt.Sprintf("HOTSTUFF LOG%d:", opts.ID)),
	)
	srv.hs = builder.Build()

	return srv
}

// TODO: parse peer info into hotstuff/config.ReplicaConfig and pass that struct into NewReplicationServer()
func createReplicaConfig(opts *options.Options) *hsconfig.ReplicaConfig {
	// Read the HotStuff ecdsa private key
	privKey, err := keygen.ReadPrivateKeyFile(opts.PrivKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read private key file: %v\n", err)
		os.Exit(1)
	}

	// Ignoring TLS for now
	replicaConfig := hsconfig.NewConfig(opts.ID, privKey, nil)
	for _, p := range opts.Peers {
		key, err := keygen.ReadPublicKeyFile(p.PubKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read public key file '%s': %v\n", p.PubKey, err)
			os.Exit(1)
		}

		info := &hsconfig.ReplicaInfo{
			ID:      p.ID,
			Address: p.ReplicationPeerAddr,
			PubKey:  key,
		}

		replicaConfig.Replicas[p.ID] = info
	}

	return replicaConfig
}

func (srv *replicationServer) Start(ctx context.Context, addr string) (err error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	srv.hsSrv.StartOnListener(lis)

	err = srv.mgr.Connect(10 * time.Second)
	if err != nil {
		return err
	}

	// sleep so that all replicas can be ready before we start
	time.Sleep(time.Second)

	c := make(chan struct{})
	go func() {
		srv.hs.EventLoop().Run(ctx)
		close(c)
	}()

	log.Printf("Replication server listening on %v.\n", addr)

	// wait for the event loop to exit
	<-c
	return nil
}

func (srv *replicationServer) Stop() {
	srv.hs.ViewSynchronizer().Stop()
	srv.mgr.Close()
	srv.hsSrv.Stop()
}

func (srv *replicationServer) Exec(cmd hotstuff.Command) {
	if cmd != "" {
		log.Println(cmd)
		log.Println("Replication finished. Writing to database")
		srv.replicatedReqs <- struct{}{}
	}
}

/*
func (certSrv *replicationServer) GetCertificate(_ context.Context, req *protocol.Request, out func(*protocol.Reply, error)) {
	// channel for signalling purpose (check if command done with consensus phase) and synchronization
	finished := make(chan struct{})
	id := reqID{89, req.SequenceNumber}
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
func (certSrv *replicationServer) Exec(cmd hotstuff.Command) {
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

*/