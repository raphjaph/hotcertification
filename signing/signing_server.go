package signing

import (
	"context"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/niclabs/tcrsa"
	"github.com/relab/gorums"
	"google.golang.org/grpc"

	"github.com/raphasch/hotcertification/crypto"
)

type signingServer struct {
	Key           *crypto.ThresholdKey
	backendSrv    *gorums.Server // handles the transport/serialization/tls....
	signingMgr    *Manager       // calls the RPC on the other servers to get a partial signature
	caCert        *x509.Certificate
	Peers         []string
	pendingCerts  chan *x509.Certificate
	finishedCerts chan *x509.Certificate
}

func NewSigningServer(key *crypto.ThresholdKey, peers []string, pendingCerts chan *x509.Certificate, finishedCerts chan *x509.Certificate) *signingServer {

	// add options here
	gorumsSrv := gorums.NewServer()
	mgr := NewManager(
		gorums.WithDialTimeout(500*time.Millisecond),
		gorums.WithGrpcDialOptions(
			grpc.WithBlock(),
			grpc.WithInsecure(),
		),
	)

	sigSrv := &signingServer{
		Key:           key,
		backendSrv:    gorumsSrv,
		signingMgr:    mgr,
		Peers:         peers,
		pendingCerts:  pendingCerts,
		finishedCerts: finishedCerts,
	}

	// glue together backend (transport/serialization) with frontend implementation (computing the partial signature)
	RegisterSigningServer(gorumsSrv, sigSrv)

	return sigSrv
}

func (srv *signingServer) GetPartialSig(_ context.Context, cert *Certificate, out func(*SigShare, error)) {
	/*
		1. Parse certificate
		2. Partially Sign certificate
		3. return signature share
	*/

	log.Println("Receiving request for partial signature on certificate.")
	certificate, err := x509.ParseCertificate(cert.Certificate)
	if err != nil {
		log.Println("Failed to parse certificate from bytes: ", err)
	}

	partialSig, err := crypto.ComputePartialSignature(certificate, srv.Key)
	if err != nil {
		log.Println("Failed to compute a partial signature: ", err)
		out(nil, fmt.Errorf("Failed to compute a partial signature"))
		return
	}

	log.Printf("Successfully partially signed certificate for client %v.\n", certificate.Subject.CommonName)

	out(&SigShare{
		Xi: partialSig.Xi,
		C:  partialSig.C,
		Z:  partialSig.Z,
		Id: uint32(partialSig.Id)},
		nil)
	return
}

func (srv *signingServer) CallGetPartialSig(cert *x509.Certificate) (partialSigs tcrsa.SigShareList, err error) {

	log.Printf("Initializing treshold signing session for certificate from %v.\n", cert.Subject.CommonName)
	// where are the other signers available
	// Create a configuration including all signers/nodes
	signersConfig, err := srv.signingMgr.NewConfiguration(gorums.WithNodeList(srv.Peers))
	if err != nil {
		log.Println("failed to initialize signing session: ", err)
		return nil, err
	}

	// in tcrsa K is threshold and L is total number of participants
	partialSigs = make(tcrsa.SigShareList, len(signersConfig.Nodes()))
	for i, node := range signersConfig.Nodes() {

		// TODO: rename Certificate to certificate bytes or raw or ...
		log.Printf("Sending partial signature request to peer %v.\n", node.ID())
		sigShare, err := node.GetPartialSig(context.Background(), &Certificate{Certificate: cert.Raw})
		if err != nil {
			log.Printf("failed to get partial signature from peer %v.\n", node.ID())
			return nil, err
		}

		partialSigs[i] = &tcrsa.SigShare{
			Xi: sigShare.GetXi(),
			C:  sigShare.GetC(),
			Z:  sigShare.GetZ(),
			Id: uint16(sigShare.GetId()),
		}
	}

	return partialSigs, nil
}

func (srv *signingServer) Start(port int) {
	// open port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println(err)
	}

	log.Printf("Signing server listening on port %v.\n", port)

	srv.backendSrv.Serve(lis)
}
