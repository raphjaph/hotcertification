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
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	pb "github.com/raphasch/hotcertification/protocol"
)

type options struct {
	TLS         bool   `mapstructure:"tls"`
	RootCA      string `mapstructure:"root-ca"`
	ServerAddr  string `mapstructure:"server-addr"`
	Destination string `mapstructure:"destination"`
}

func usage() {
	fmt.Printf("Usage: %s [options] [destination]\n", os.Args[0])
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

	flag.Parse()

	if *help {
		usage()
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		fmt.Println("Please specify where the certificate file should be written")
		usage()
		os.Exit(1)
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

	opts.Destination = flag.Arg(0)

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
	hotcertification := pb.NewCertificationClient(conn)

	// generate private and public key for certificate
	keySize := 512
	clientKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		log.Fatalf("failed to generate client keys: %v", err)
		return
	}

	// creating CSR with client public key
	clientCSR, err := generateCSR(clientKey)
	if err != nil {
		log.Fatalf("failed to generate CSR: %v", err)
		return
	}

	// Adding some context
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	csr := &pb.CSR{
		ClientID:           8,
		CertificateRequest: clientCSR.Raw,
		ValidationInfo:     make([]byte, 100),
	}

	measurements := make([]time.Duration, 100)
	for i := 0; i < 100; i++ {

		start := time.Now()
		// putting CSR into protocol buffers format and calling remote function
		_, err := hotcertification.GetCertificate(ctx, csr)
		if err != nil {
			log.Fatalf("failed to call RPC: %v", err)
			return
		}
		elapsed := time.Since(start)
		measurements[i] = elapsed
	}

	file, err := os.Create("measurements.csv")
	checkError("Cannot create file", err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, duration := range measurements {
		err := writer.Write([]string{fmt.Sprintf("%d", duration.Milliseconds())})
		checkError("Cannot write to file", err)
	}

	//certificate, err := x509.ParseCertificate(response.Certificate)
	//if err != nil {
	//log.Fatalf("failed to parse certificate: %v", err)
	//return
	//}

	// TODO: verify signature with root certificate

	// Write certificate to file

	//crypto.WriteCertFile(certificate, opts.Destination)

	//fmt.Println("Wrote certificate to file")
}

func generateCSR(clientPrivKey *rsa.PrivateKey) (csr *x509.CertificateRequest, err error) {
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

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}
