package hotcertification

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"sync"

	"github.com/relab/hotstuff"
	"google.golang.org/protobuf/proto"

	"github.com/raphasch/hotcertification/logging"
	"github.com/raphasch/hotcertification/protocol"
)

// Information other replicas in network have to know about each other (public knowledge)
// Private knowledge (threshold key, ecdsa private key) have to given through command line and not in global config file
type Node struct {
	ID                 hotstuff.ID
	PubKey             string `mapstructure:"pubkey"`
	TLSCert            string `mapstructure:"tls-cert"`
	ClientSrvAddr      string `mapstructure:"client-srv-address"`
	ReplicationSrvAddr string `mapstructure:"replication-srv-address"`
	SigningSrvAddr     string `mapstructure:"signing-srv-address"`
}

type Options struct {
	// The ID of this server
	ID hotstuff.ID `mapstructure:"id"`

	// TLS configs
	RootCA  string `mapstructure:"root-ca"`
	TLS     bool   `mapstructure:"tls"`
	PrivKey string `mapstructure:"privkey"` // privkey has to belong the to the pubkey and should be ecdsa because thresholdkey can't do TLS

	// HotStuff configs
	PmType      string      `mapstructure:"pacemaker"`
	LeaderID    hotstuff.ID `mapstructure:"leader-id"`
	ViewTimeout int         `mapstructure:"view-timeout"`

	// HotCertification and miscellaneous configs
	ThresholdKey string `mapstructure:"thresholdkey"`
	KeySize      int    `mapstructure:"key-size"`
	ConfigFile   string `mapstructure:"config"`
	Nodes        []Node
}

type RequestInfo struct {
	CSR         *protocol.CSR
	Certificate *x509.Certificate
	Received    bool
	Validated   bool
	Proposed    bool
	Replicated  bool
	Signed      bool
	Returned    bool
}

type Coordinator struct {
	Mut              sync.Mutex
	ReplicationQueue chan *protocol.CSR
	SigningQueue     chan *protocol.CSR
	FinishedCerts    chan *protocol.Certificate
	Database         map[string]*RequestInfo // simulating a basic database; the key the hash of the CSR
	Marshaler        proto.MarshalOptions    // for translating into hotstuff.Command
	Unmarshaler      proto.UnmarshalOptions  // for checking semantics of a request
	HS               *hotstuff.HotStuff
	Log              logging.Logger
	c                chan struct{}
}

func NewCoordinator() *Coordinator {
	// TODO: How to pick good channel size?
	return &Coordinator{
		ReplicationQueue: make(chan *protocol.CSR, 100000),
		SigningQueue:     make(chan *protocol.CSR, 100000),
		FinishedCerts:    make(chan *protocol.Certificate, 100000),
		Database:         make(map[string]*RequestInfo),
		Marshaler:        proto.MarshalOptions{Deterministic: true},
		Unmarshaler:      proto.UnmarshalOptions{DiscardUnknown: true},
		Log:              logging.New("HOT LOG"),
		c:                make(chan struct{}),
	}
}

// InitModule gives the module a reference to the HotStuff object.
func (c *Coordinator) InitModule(hs *hotstuff.HotStuff, _ *hotstuff.OptionsBuilder) {
	c.HS = hs
}

func (c *Coordinator) AddRequest(csr *protocol.CSR) {
	/*
		0. ?Validate Request or do this in protocolServer struct?
		1. Wrap protocol.CSR into RequestInfo struct
		2. Add to database identified to ReqID or Hash of CSR
		3. Add to Queue
	*/

	c.ReplicationQueue <- csr

	c.Log.Info("Added to CSR to Replication Queue")

	hash := HashCSR(csr)
	info := &RequestInfo{
		CSR:        csr,
		Received:   true,
		Validated:  false,
		Proposed:   false,
		Replicated: false,
		Signed:     false,
		Returned:   false,
	}

	c.Mut.Lock()
	c.Database[hash] = info
	c.Mut.Unlock()
}

// Implements the CommandQueue for HotStuff to get next requests to replicate
func (c *Coordinator) Get(ctx context.Context) (cmd hotstuff.Command, ok bool) {
	/*
		1. Checks wether there are pending requests
		1a. TODO: Has consume/be waiting for the ReplicationQueue continuosly
		2. Marshal them into the hotstuff.Command format
		3. Return command, true
		4. Run continuosly until ctx cancelled ? or return "", false ? or received through channel another signal
	*/
	// for now it only processes on request at a time, in future also batching possible

	// This will probably lock?

	var csr *protocol.CSR

	select {
	case csr = <-c.ReplicationQueue:
	case <-ctx.Done():
		return "", false
	default: // default makes this non-blocking (https://gobyexample.com/non-blocking-channel-operations)
	}

	bytes, err := c.Marshaler.Marshal(csr)
	if err != nil {
		c.Log.Errorf("Failed to marshal batch: %v", err)
		return "", false
	}

	cmd = hotstuff.Command(bytes)

	return cmd, true

}

// Implements the Acceptor for HotStuff and the Validator for the certification process
func (c *Coordinator) Accept(cmd hotstuff.Command) bool {
	/*
		1. Unmarshal hotstuff.Command to CSR + ValidationInfo
		2. Validate CSR with ValidationInfo
		3. Update database (async?)
		4. return true or false
	*/

	// TODO: fix this stupid workaround in Get() function
	if cmd == "" {
		return true
	}

	csr := new(protocol.CSR)
	err := c.Unmarshaler.Unmarshal([]byte(cmd), csr)
	if err != nil {
		c.Log.Errorf("Failed to unmarshal command: %v", err)
		return false
	}

	hash := HashCSR(csr)

	// get certificate so that it can be validated
	_, err = x509.ParseCertificateRequest(csr.CertificateRequest)
	if err != nil {
		c.Log.Error(err)
	}

	c.Log.Infof("Validating CSR %v", hash[:6])
	validated := true

	c.Mut.Lock()
	defer c.Mut.Unlock()

	if c.Database[hash] == nil {
		c.Log.Info("Adding to database")
		info := &RequestInfo{
			CSR:        csr,
			Received:   false,
			Validated:  validated,
			Proposed:   false,
			Replicated: false,
			Signed:     false,
			Returned:   false,
		}
		c.Database[hash] = info

	}

	c.Database[hash].Validated = validated

	return !c.Database[hash].Proposed
}

// Tells the coordinator that the request/batch of requests have succesfully been proposed to other nodes
func (c *Coordinator) Proposed(cmd hotstuff.Command) {
	/*
		1. Unmarshal hotstuff.Command to CSR format
		2. Get database key
		3. Update state to proposed because Propose Phase of HotStuff was successful
		4. Accept() shouldn't accept any old CSRs
	*/

	// TODO:

	/*
		csr := new(protocol.CSR)
		err := c.Unmarshaler.Unmarshal([]byte(cmd), csr)
		if err != nil {
			c.Log.Error("Failed to unmarshal command: %v\n", err)
		}

		id := csr.ClientID
		log.Println(id)
		if id != 0 {
			c.Mut.Lock()
			c.Database[id].Proposed = true
			c.Mut.Unlock()
		}
	*/
}

// Implements the Executor for HotStuff and starts the threshold signing process for a given (batch of) request(s)
func (c *Coordinator) Exec(cmd hotstuff.Command) {
	/*
		1. Unmarshal hotstuff.Command to CSR format
		2. Update state in database to replicated
		3. Signal with channel to partial signing routine that (batch of) CSRs ready for threshold signing process
	*/

	if cmd == "" {
		return
	}

	csr := new(protocol.CSR)
	err := c.Unmarshaler.Unmarshal([]byte(cmd), csr)
	if err != nil {
		c.Log.Errorf("Failed to unmarshal command: %v", err)
		return
	}

	hash := HashCSR(csr)

	c.Mut.Lock()
	defer c.Mut.Unlock()

	reqInfo := c.Database[hash]
	if reqInfo == nil {
		c.Log.Infof("Couldn't find CSR %v in database.", hash[:6])
		return
	}

	// if this is server handling client request then initiates signing sesshion
	if reqInfo.Received {
		c.Log.Info("Replication finished.")

		// TODO: make this channel send non-blocking with select+default
		c.SigningQueue <- csr
	}
	c.Database[hash].Replicated = true
}

// helper functions

func HashCSR(csr *protocol.CSR) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", csr.ClientID)))
	h.Write([]byte(fmt.Sprintf("%v", csr.CertificateRequest)))
	h.Write([]byte(fmt.Sprintf("%v", csr.ValidationInfo)))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func NumFaulty(n int) int {
	return (n - 1) / 3
}

func QuorumSize(n int) int {
	return n - NumFaulty(n)
}
