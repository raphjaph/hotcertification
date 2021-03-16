package main

import (
	"context"
	"fmt"
	"time"

	"github.com/raphasch/hotcertification/protocol"
	"github.com/relab/gorums"
	"google.golang.org/grpc"
)

func main() {
	client, err := newCertificationClient()
	if err != nil {
		fmt.Printf("couldn't create client: %v\n", err)
	}

	//TODO: figure out how to properly use context
	//TODO: use logger instead of fmt.Printf()!

	//Testing RegisterCSR
	// synchronous RPC on one node in the Configuration (Coordinator Node for this client)
	/* csr := client.generateCSR("Batman")
	request := &protocol.Request{ClientID: 8, SequenceNumber: 1, CSR: csr}
	coordinatorNode := client.nodesConfig.Nodes()[0]
	response, err := coordinatorNode.RegisterCSR(context.Background(), request)
	if err != nil {
		fmt.Printf("sending Registration Request request went wrong: %v\n", err)
	}
	fmt.Println(response) */

	//Testing GetCertificate(); handled synchronously
	// TODO: add sync.WaitGroup
	certRequest := &protocol.Request{ClientID: 8, SequenceNumber: 2, Data: "CSR"}
	ctx, cancel := context.WithCancel(context.Background())
	fmt.Println("Calling GetCertificate()")
	certPromise := client.allNodes.GetCertificate(ctx, certRequest)
	certificate, err := certPromise.Get()
	if err != nil {
		qcError, ok := err.(gorums.QuorumCallError)
		if !ok || qcError.Reason != context.Canceled.Error() {
			fmt.Printf("Did not get enough signatures for certificate: %v\n", err)
		}
	}
	fmt.Println(certificate.Data)
	cancel()

	// close connections
	client.connectionMgr.Close()
}

type certificationClient struct {
	//reader			io.ReadCloser // reads the commands; just for testing?
	connectionMgr *protocol.Manager       // the gorums/gRPC server handles all the connections
	allNodes      *protocol.Configuration // A configuration is a set of nodes on which RPC calls can be invoked. Maybe on config for HotStuff and another for the Signing and Validation
	//replicaConf		*config.ReplicaConfig // stores the node information of the network of nodes that represent the Certification Authority
}

func newCertificationClient() (*certificationClient, error) {
	// TODO: grpc.WithBlock and WithInsecure definitions
	// TODO: understand DialTimeout and other options
	mgr := protocol.NewManager(
		gorums.WithDialTimeout(500*time.Millisecond),
		gorums.WithGrpcDialOptions(
			grpc.WithBlock(),
			grpc.WithInsecure(),
		),
	)

	// if this would be pure gRPC I would use a connection channel but since gorums -> quorum call using a manager that manages the connections with all the servers
	// normally list of all servers
	addrs := []string{
		"localhost:23371",
		"localhost:23372",
		"localhost:23373",
		"localhost:23374",
	}

	// A configuration is a set of nodes on which RPC calls can be invoked. The manager assigns every node and configuration a unique identifier.
	allNodes, err := mgr.NewConfiguration(
		&QuorumSpec{3},
		gorums.WithNodeList(addrs),
	)
	if err != nil {
		fmt.Println("error creating read config:", err)
	}

	return &certificationClient{
		connectionMgr: mgr,
		allNodes:      allNodes,
	}, nil
}

/*
func (client *certificationClient) generateCSR(identity string) *protocol.CSR {
	certificate := &protocol.Certificate{Identity: identity, PublicKey: "batman's public key"}
	validationInfo := client.getValidationInfo()
	return &protocol.CSR{
		Certificate:    certificate,
		ValidationInfo: validationInfo,
	}
}

func (client *certificationClient) getValidationInfo() *protocol.ValidationInfo {
	return &protocol.ValidationInfo{
		LongTermPublicKey: "This is a long term public key to for validation of a CSR",
		SignedMessage:     "This signed message proves that the sender has the corresponding private key",
	}
}
*/

// TODO: implement this fully and make own file?
// TODO: find different name like ThresholdSpecification?
// qorum specifications defines how many nodes in a quorum or how many of the quorum allowed to be
// and also implements stores the RPC stubs of quorum function (QF). -> if gorums.quorumcall set to true in certification.protocol
type QuorumSpec struct {
	quorumSize int
}

// waits for the replies of all the servers and combines individual signatures into one fully signed certificate
func (qs *QuorumSpec) GetCertificateQF(_ *protocol.Request, replies map[uint32]*protocol.Reply) (*protocol.Reply, bool) {
	// not enough to generate certificate
	// check certificate correct
	if len(replies) < qs.quorumSize {
		return nil, false
	}

	// IMPORTANT: This is where the full certificate is computed from the replies of the individual servers
	fullCertificate := &protocol.Reply{}
	for _, reply := range replies {
		fullCertificate.Data += reply.Data
		fullCertificate.Data += " + "
	}
	fullCertificate.Data += " = Fully Signed Certificate"
	return fullCertificate, true
}
