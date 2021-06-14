package signing

import (
	"context"
	"crypto/x509"
	"fmt"
	"net"
	"os"
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
	mgr         *Manager // calls the RPC on the other servers to get a partial signature
	cfg         *Configuration
	backendSrv  *gorums.Server // handles the transport/serialization/tls....
}

func NewSigningServer(coordinator *hc.Coordinator, key *crypto.ThresholdKey, nodes []string, rootCAFile string) *signingServer {
	// add options here
	gorumsSrv := gorums.NewServer()
	mgr := NewManager(
		gorums.WithDialTimeout(10*time.Second),
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
		mgr:         mgr,
		rootCA:      rootCA,
		nodes:       nodes,
		coordinator: coordinator,
	}

	// glue together backend (transport/serialization) with frontend implementation (computing the partial signature)
	RegisterSigningServer(gorumsSrv, sigSrv)

	return sigSrv
}

func (srv *signingServer) GetPartialSig(_ context.Context, tbs *TBS, out func(*SigShare, error)) {
	/*
		1. Parse certificate
		2. Partially Sign certificate
		3. return signature share
		4. TODO: check database if already signed and then update to signed = true
	*/

	srv.coordinator.Log.Info("Received request for partial signature. Checking authorization.")
	info := srv.coordinator.Database[tbs.CSRHash]
	if info == nil || !info.Validated {
		srv.coordinator.Log.Error("CSR has not been validated.")
		out(nil, fmt.Errorf("CSR has not been validated"))
		return
	}

	cert, err := x509.ParseCertificate(tbs.Certificate)
	if err != nil {
		srv.coordinator.Log.Error("error parsing certificate")
	}

	partialSig, err := crypto.ComputePartialSignature(cert, srv.key)
	if err != nil {
		srv.coordinator.Log.Error("failed to compute a partial signature: ", err)
		out(nil, fmt.Errorf("failed to compute a partial signature"))
		return
	}

	srv.coordinator.Log.Info("Successfully partially signed certificate for CSR ", tbs.CSRHash[:6])
	info.Signed = true

	out(&SigShare{
		Xi: partialSig.Xi,
		C:  partialSig.C,
		Z:  partialSig.Z,
		Id: uint32(partialSig.Id)},
		nil)
}

func (srv *signingServer) GetFullSignature(csr *protocol.CSR) (*x509.Certificate, error) {

	hash := hc.HashCSR(csr)
	srv.coordinator.Log.Info("Initializing treshold signing session CSR ", hash[:6])

	x509csr, err := x509.ParseCertificateRequest(csr.CertificateRequest)
	if err != nil {
		srv.coordinator.Log.Error(err)
	}

	cert, err := crypto.GenerateCert(x509csr, srv.rootCA, srv.key)
	if err != nil {
		srv.coordinator.Log.Error(err)
	}

	// TODO: rename to quorumAnswer? quorumOfReplies?
	thresholdOf, err := srv.cfg.GetPartialSig(context.Background(), &TBS{CSRHash: hash, Certificate: cert.Raw})
	if err != nil {
		srv.coordinator.Log.Errorf("failed to get enough partial signatures.")
		return nil, err
	}

	// in tcrsa K is threshold and L is total number of participants
	partialSigs := make(tcrsa.SigShareList, hc.QuorumSize(len(srv.cfg.Nodes())))
	for i, share := range thresholdOf.SigShares {
		partialSigs[i] = &tcrsa.SigShare{
			Xi: share.GetXi(),
			C:  share.GetC(),
			Z:  share.GetZ(),
			Id: uint16(share.GetId()),
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
	// starting listener
	go srv.backendSrv.Serve(lis)

	// creating configuration (group of signers)
	signersConfig, err := srv.mgr.NewConfiguration(
		&QSpec{hc.QuorumSize(len(srv.nodes))},
		gorums.WithNodeList(srv.nodes))
	if err != nil {
		srv.coordinator.Log.Error("failed to initialize signing configuration: ", err)
		os.Exit(1)
	}
	srv.cfg = signersConfig

	srv.coordinator.Log.Infof("Signing server listening on %v.", addr)

	// TODO: add cancel function through context
	for {
		csr := <-srv.coordinator.SigningQueue
		cert, err := srv.GetFullSignature(csr)
		if err != nil {
			srv.coordinator.Log.Errorf("Couldn't generate full signature: %v", err)
			srv.coordinator.FinishedCerts <- &protocol.Certificate{}
		} else {
			// wrap x509 cert and convert to ASN.1 DER encoded byte array
			srv.coordinator.FinishedCerts <- &protocol.Certificate{Certificate: cert.Raw}
		}

	}
}

type QSpec struct {
	quorumSize int
}

func (qs *QSpec) GetPartialSigQF(_ *TBS, sigShares map[uint32]*SigShare) (*ThresholdOf, bool) {
	if len(sigShares) < qs.quorumSize {
		return nil, false
	}
	// TODO: is there a more efficient way?
	shares := []*SigShare{}
	for _, share := range sigShares {
		shares = append(shares, share)
	}
	return &ThresholdOf{SigShares: shares}, true
}
