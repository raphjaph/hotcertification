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
	"sync"

	"github.com/relab/hotstuff"
	hotstuffbackend "github.com/relab/hotstuff/backend/gorums"
)

type options struct {
	// The ID of this server
	ID int `mapstructure:"id"`

	// TLS configs
	RootCA  string `mapstructure:"root-ca"`
	TLS     bool   `mapstructure:"tls"`
	PrivKey string `mapstructure:"privkey"` // privkey has to belong the to the pubkey and should be ecdsa because thresholdkey can't do TLS

	// HotStuff configs
	PmType   string      `mapstructure:"pacemaker"`
	LeaderID hotstuff.ID `mapstructure:"leader-id"`

	// HotCertification and miscellaneous configs
	ThresholdKey string `mapstructure:"thresholdkey"`
	KeySize      int    `mapstructure:"key-size"`
	ConfigFile   string `mapstructure:"config"`
	//Peers        []peer
}

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
	reqBuffer      *reqBuffer               // the request buffer (CSRs); passed in from client server
	mut            sync.Mutex
	replicatedReqs chan struct{} // TODO: change from struct{} to *client.CSR or *x509.CertificateRequest. Can I put anything into a chan struct{} and then transform at the other end through reflection
	//finishedReqs map[reqID]chan struct{} // signalling channel

	lastExecTime int64
}

func NewReplicationServer(opts *options, replicatedRequests chan struct{}) *replicationServer {

	srv := &replicationServer{
		reqBuffer:      NewRequestBuffer(100),
		replicatedReqs: replicatedRequests,
	}

	// building the hotstuff consensus algorithm

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
