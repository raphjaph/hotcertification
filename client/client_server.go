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
	pendingCSRs   chan *CSR
	finishedCerts chan *x509.Certificate
	backendSrv    *grpc.Server
}

func NewClientServer(pendingCSRs chan *CSR, finishedCerts chan *x509.Certificate) *clientServer {

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

	// send to signing server
	srv.pendingCSRs <- csr

	// wait for fully signed certificate
	certificate := <-srv.finishedCerts

	log.Println("Returning fully signed certificate to client.")

	return &Certificate{Certificate: certificate.Raw}, nil
}

func (srv *clientServer) Start(addr string) {

	// open port
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println(err)
	}

	log.Printf("Client server listening on %v.\n", addr)

	srv.backendSrv.Serve(lis)
}
