/*
	CLIENT SERVER LOGIC:
		1. Parses the ASN.1 encoded byte array into a x509 certificate request
		2. Passes it on to signing server through a channel
		3. Waits for full signed certificate on other channel
		4. Serializes fully signed certificate and returns it back to client
*/
package main

import (
	"context"
	"fmt"
	"net"

	hc "github.com/raphasch/hotcertification"
	"github.com/raphasch/hotcertification/protocol"
	"google.golang.org/grpc"
)

type clientServer struct {
	backendSrv  *grpc.Server
	coordinator *hc.Coordinator

	// gRPC stuff for backward compatability
	protocol.UnimplementedCertificationServer
}

func NewClientServer(coordinator *hc.Coordinator) *clientServer {

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	clientSrv := &clientServer{
		backendSrv:  grpcServer,
		coordinator: coordinator,
	}

	protocol.RegisterCertificationServer(grpcServer, clientSrv)

	return clientSrv
}

// need this because [see here](https://stackoverflow.com/questions/65079032/grpc-with-mustembedunimplemented-method)
func (srv *clientServer) mustEmbedUnimplementedCertificationServer() {}

func (srv *clientServer) GetCertificate(_ context.Context, csr *protocol.CSR) (*protocol.Certificate, error) {

	srv.coordinator.Log.Info("Received CSR ", hc.HashCSR(csr)[:6])

	// First step; replication
	srv.coordinator.AddRequest(csr)

	// wait for fully signed certificate
	certificate := <-srv.coordinator.FinishedCerts
	if certificate.Certificate != nil {
		srv.coordinator.Log.Info("Returning fully signed certificate to client.")
		return certificate, nil
	} else {
		return nil, fmt.Errorf("couldn't compute full signature on certificate")
	}
}

func (srv *clientServer) Start(addr string) {

	// open port
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println(err)
	}

	srv.coordinator.Log.Infof("Client server listening on %v.", addr)

	srv.backendSrv.Serve(lis)
}
