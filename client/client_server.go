/*
	CLIENT SERVER LOGIC:
		1. Parses the ASN.1 encoded byte array into a x509 certificate request
		2. Passes it on to signing server through a channel
		3. Waits for full signed certificate on other channel
		4. Serializes fully signed certificate and returns it back to client
*/
package client

import (
	"context"
	"crypto/x509"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

type clientServer struct {
	pendingCSRs   chan *x509.CertificateRequest
	finishedCerts chan *x509.Certificate
	backendSrv    *grpc.Server
}

func NewClientServer(pendingCSRs chan *x509.CertificateRequest, finishedCerts chan *x509.Certificate) *clientServer {

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	clientSrv := &clientServer{
		pendingCSRs:   pendingCSRs,
		finishedCerts: finishedCerts,
		backendSrv:    grpcServer,
	}

	RegisterCertificationServer(grpcServer, clientSrv)

	return clientSrv
}

// need this because [see here](https://stackoverflow.com/questions/65079032/grpc-with-mustembedunimplemented-method)
func (srv *clientServer) mustEmbedUnimplementedCertificationServer() {}

func (srv *clientServer) GetCertificate(_ context.Context, csr *CSR) (*Certificate, error) {

	log.Println("Received CSR")

	certRequest, err := x509.ParseCertificateRequest(csr.CSR)
	if err != nil {
		return nil, err
	}
	// send to signing server
	srv.pendingCSRs <- certRequest

	// wait for fully signed certificate
	certificate := <-srv.finishedCerts

	log.Println("Returning fully signed certificate to client.")

	return &Certificate{Certificate: certificate.Raw}, nil
}

func (srv *clientServer) Start(port int) {

	// open port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println(err)
	}

	log.Printf("Client server listening on port %v.\n", port)

	srv.backendSrv.Serve(lis)
}
