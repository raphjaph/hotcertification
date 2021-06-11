package signing

import (
	"context"
	"crypto/x509"
	"fmt"
	"net"
	"time"

	"github.com/niclabs/tcrsa"
	"github.com/relab/gorums"
	"google.golang.org/grpc"

	hc "github.com/raphasch/hotcertification"
	"github.com/raphasch/hotcertification/crypto"
	"github.com/raphasch/hotcertification/protocol"
)

type signingServer struct {
	key         *crypto.ThresholdKey
	rootCA      *x509.Certificate
	coordinator *hc.Coordinator
	nodes       []string
	signingMgr  *Manager       // calls the RPC on the other servers to get a partial signature
	backendSrv  *gorums.Server // handles the transport/serialization/tls....
}

func NewSigningServer(coordinator *hc.Coordinator, key *crypto.ThresholdKey, nodes []string, rootCAFile string) *signingServer {

	// add options here
	gorumsSrv := gorums.NewServer()
	mgr := NewManager(
		gorums.WithDialTimeout(500*time.Millisecond),
		gorums.WithGrpcDialOptions(
			grpc.WithBlock(),
			grpc.WithInsecure(),
		),
	)

	rootCA, err := crypto.ReadCertFile(rootCAFile)
	if err != nil {
		coordinator.Log.Error(err)
	}

	sigSrv := &signingServer{
		key:         key,
		backendSrv:  gorumsSrv,
		signingMgr:  mgr,
		rootCA:      rootCA,
		nodes:       nodes,
		coordinator: coordinator,
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
		4. TODO: check database if already signed and then update to signed = true
	*/

	srv.coordinator.Log.Info("Receiving request for partial signature on certificate.")
	certificate, err := x509.ParseCertificate(cert.Certificate)
	if err != nil {
		srv.coordinator.Log.Error("Failed to parse certificate from bytes: ", err)
	}

	partialSig, err := crypto.ComputePartialSignature(certificate, srv.key)
	if err != nil {
		srv.coordinator.Log.Error("failed to compute a partial signature: ", err)
		out(nil, fmt.Errorf("failed to compute a partial signature"))
		return
	}

	srv.coordinator.Log.Infof("Successfully partially signed certificate for client %v.", certificate.Subject.CommonName)

	out(&SigShare{
		Xi: partialSig.Xi,
		C:  partialSig.C,
		Z:  partialSig.Z,
		Id: uint32(partialSig.Id)},
		nil)
}

func (srv *signingServer) GetFullSignature(csr *protocol.CSR) (*x509.Certificate, error) {
	// where are the other signers available
	// Create a configuration including all signers/nodes

	// TODO: do not create this everytime but once in constructor
	signersConfig, err := srv.signingMgr.NewConfiguration(gorums.WithNodeList(srv.nodes))
	if err != nil {
		srv.coordinator.Log.Error("failed to initialize signing session: ", err)
		return nil, err
	}

	x509csr, err := x509.ParseCertificateRequest(csr.CertificateRequest)
	if err != nil {
		srv.coordinator.Log.Error(err)
	}

	cert, err := crypto.GenerateCert(x509csr, srv.rootCA, srv.key)
	if err != nil {
		srv.coordinator.Log.Error(err)
	}

	srv.coordinator.Log.Infof("Initializing treshold signing session for certificate from %v.", cert.Subject.CommonName)

	// in tcrsa K is threshold and L is total number of participants
	partialSigs := make(tcrsa.SigShareList, len(signersConfig.Nodes()))
	for i, node := range signersConfig.Nodes() {

		// TODO: rename Certificate to certificate bytes or raw or ...
		srv.coordinator.Log.Infof("Sending partial signature request to Node %v.", node.ID())
		sigShare, err := node.GetPartialSig(context.Background(), &Certificate{Certificate: cert.Raw})
		if err != nil {
			srv.coordinator.Log.Errorf("failed to get partial signature from Node %v.", node.ID())
			return nil, err
		}

		partialSigs[i] = &tcrsa.SigShare{
			Xi: sigShare.GetXi(),
			C:  sigShare.GetC(),
			Z:  sigShare.GetZ(),
			Id: uint16(sigShare.GetId()),
		}
	}

	srv.coordinator.Log.Info("Computing full signature for certificate.")
	fullCert, err := crypto.ComputeFullySignedCert(cert, srv.key, partialSigs...)
	if err != nil {
		return nil, fmt.Errorf("failed to compute full signature: %v", err)
	}

	return fullCert, nil
}

func (srv *signingServer) Start(addr string) {
	// open port
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println(err)
	}

	srv.coordinator.Log.Infof("Signing server listening on %v.", addr)

	go srv.backendSrv.Serve(lis)

	// TODO: add cancel function through context
	for {
		csr := <-srv.coordinator.SigningQueue
		cert, err := srv.GetFullSignature(csr)
		if err != nil {
			srv.coordinator.Log.Errorf("Couldn't generate full signature: %v", err)
		}

		// wrap x509 cert and convert to ASN.1 DER encoded byte array
		srv.coordinator.FinishedCerts <- &protocol.Certificate{Certificate: cert.Raw}
	}
}
