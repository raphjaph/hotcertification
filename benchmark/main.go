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
	"io/ioutil"
	"log"
	"os"
	"strings"
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
	File        string `mapstructure:"file"`
	Scenario    string `mapstructure:"scenario"`
	Num         int    `mapstructure:"num"`
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
	flag.String("file", "", "The file to attach to the CSR as the Validation Info.")
	flag.String("scenario", "4,100,none,0", "What scenario the measurements are for.")
	num := flag.IntP("number", "n", 100, "number of requests to send.")

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
	opts.Num = *num

	return opts
}

func getClient(opts options) (client pb.CertificationClient, err error) {
	// grpc setup
	var gRPC_opts []grpc.DialOption
	if opts.TLS {
		return nil, fmt.Errorf("tls not implemented yet")
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
		return nil, fmt.Errorf("failed to dial: %v", err)
	}
	//defer conn.Close()
	client = pb.NewCertificationClient(conn)

	return client, nil
}

func generateTestCSR(opts options) (csr *pb.CSR, err error) {

	// generate private and public key for certificate
	clientKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		return nil, fmt.Errorf("failed to generate client keys: %v", err)
	}

	csrTmpl := &x509.CertificateRequest{
		SignatureAlgorithm: x509.SHA256WithRSA,
		PublicKeyAlgorithm: x509.RSA,
		Subject: pkix.Name{
			CommonName: "Raphael Schleithoff",
		},
		EmailAddresses: []string{"raphael.schleithoff@tum.de"},
	}

	// the client's public key is inserted into the CSR
	bytes, err := x509.CreateCertificateRequest(rand.Reader, csrTmpl, clientKey)
	if err != nil {
		return nil, err
	}

	var valInfo []byte
	if opts.File == "" {
		opts.File = "0.info"
		_, err := os.Create(opts.File)
		checkError("failed to create dummy file: ", err)
	}

	valInfo, err = ioutil.ReadFile(opts.File)
	checkError("failed to open validation info file: ", err)

	csr = &pb.CSR{
		ClientID:           8,
		CertificateRequest: bytes,
		ValidationInfo:     valInfo,
	}

	return csr, nil
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

func main() {
	opts := parseOptions()

	client, err := getClient(opts)
	checkError("couldn't instantiate client: ", err)

	csr, err := generateTestCSR(opts)
	checkError("failed to generate CSR:", err)

	measurements := make([]time.Duration, opts.Num)
	for i := 0; i < opts.Num; i++ {

		start := time.Now()
		// putting CSR into protocol buffers format and calling remote function
		_, err := client.GetCertificate(context.Background(), csr)
		checkError("failed to call RPC:", err)

		elapsed := time.Since(start)
		measurements[i] = elapsed
	}

	file, err := os.OpenFile(opts.Destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	checkError("Cannot create file", err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header/Column names
	writer.Write([]string{"time-to-certificate", "num-nodes", "csr-size", "adversary-type", "adversary-fraction"})
	row := strings.Split(opts.Scenario, ",")
	for _, duration := range measurements {
		t2c := []string{fmt.Sprintf("%d", duration.Milliseconds())}
		err := writer.Write(append(t2c, row...))
		checkError("Cannot write to file", err)
	}
}
