/*
	CLIENT EXECUTABLE:
		1. Client generates her rsa or ecdsa key (doesn't matter)
		2. Client creates a CSR (with the "x509" package)
		3. Client uses gRPC to send a CSR
*/
package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"log"
	"os"
	"time"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	pb "example/threshold/client"
	"example/threshold/crypto"
)

type options struct {
	TLS        bool   `mapstructure:"tls"`
	RootCA     string `mapstructure:"root-ca"`
	ServerAddr string `mapstructure:"server-addr"`
	Output     string `mapstructure:"output"`
}

func usage() {
	fmt.Printf("Usage: %s [options]\n", os.Args[0])
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
}

func parseOptions() options {
	flag.Usage = usage
	help := flag.BoolP("help", "h", false, "Prints this text.")
	flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	flag.String("root-ca", "", "The file containing the root CA  cert file")
	flag.String("server-addr", "localhost:8081", "The server address in the format of host:port")
	flag.String("output", "", "The file to which the certificate is written")
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	err := viper.BindPFlags(flag.CommandLine)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	var opts options
	err = viper.Unmarshal(&opts)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	return opts
}

func main() {
	// from command line
	opts := parseOptions()

	// grpc setup
	var gRPC_opts []grpc.DialOption
	if opts.TLS {
		log.Fatalf("tls not implemented yet")
		return
		//creds, err := credentials.NewClientTLSFromFile(*caFile, *serverHostOverride)
		//if err != nil {
		//	log.Fatalf("Failed to create TLS credentials %v", err)
		//}
		//opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		gRPC_opts = append(gRPC_opts, grpc.WithInsecure())
	}
	gRPC_opts = append(gRPC_opts, grpc.WithBlock())

	// gprc start connection
	conn, err := grpc.Dial(opts.ServerAddr, gRPC_opts...)
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
		return
	}
	defer conn.Close()
	client := pb.NewCertificationClient(conn)

	// generate private and public key for certificate
	keySize := 512
	clientKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		log.Fatalf("failed to generate client keys: %v", err)
		return
	}

	// creating CSR with client public key
	clientCSR, err := GenerateCSR(clientKey)
	if err != nil {
		log.Fatalf("failed to generate CSR: %v", err)
		return
	}

	// Adding some context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// putting CSR into protocol buffers format and calling remote function
	response, err := client.GetCertificate(ctx, &pb.CSR{CSR: clientCSR.Raw})
	if err != nil {
		log.Fatalf("failed to call RPC: %v", err)
		return
	}

	certificate, err := x509.ParseCertificate(response.Certificate)
	if err != nil {
		log.Fatalf("failed to parse certificate: %v", err)
		return
	}

	// TODO: verify signature with root certificate

	// Write certificate to file
	if opts.Output != "" {
		crypto.WriteCertFile(certificate, opts.Output)
	}

	fmt.Println("Wrote certificate to file")

	return
}

func GenerateCSR(clientPrivKey *rsa.PrivateKey) (csr *x509.CertificateRequest, err error) {
	csrTmpl := &x509.CertificateRequest{
		SignatureAlgorithm: x509.SHA256WithRSA,
		PublicKeyAlgorithm: x509.RSA,
		Subject: pkix.Name{
			CommonName: "Raphael Schleithoff",
		},
		EmailAddresses: []string{"raphael.schleithoff@tum.de"},
	}

	// the client's public key is inserted into the CSR
	bytes, err := x509.CreateCertificateRequest(rand.Reader, csrTmpl, clientPrivKey)
	if err != nil {
		fmt.Println(err)
	}

	return x509.ParseCertificateRequest(bytes)
}
