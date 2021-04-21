/*
	HOTCERTIFICATION LOGIC:
		1. Verify Proof of Identity of CSR
		2. Replicate through HotStuff
		3. Generate x509 Certificate
		4. Compute Partial Signature and store in database
			4a. If Gateway Node then ask other replicas for partial signatures
				4a1. Verify partial signatures
				4a2. Compute Full signature and return certificate to client
			4b. If Request from other node then pass on partial signature

*/

package main

import (
	"context"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"os/signal"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	hc "github.com/raphasch/hotcertification"
	"github.com/raphasch/hotcertification/client"
	"github.com/raphasch/hotcertification/crypto"
	"github.com/raphasch/hotcertification/options"
	"github.com/raphasch/hotcertification/replication"
	"github.com/raphasch/hotcertification/signing"
)

func usage() {
	fmt.Printf("Usage: %s [options]\n", os.Args[0])
	fmt.Println()
	fmt.Println("Loads configuration from ./hotcertification.yml and file specified by --config")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
}

func parseOptionsAndConfig() *options.Options {
	flag.Usage = usage

	help := flag.BoolP("help", "h", false, "Prints this text.")
	config := flag.String("config", "", "The path to the config file in case it isn't in working directory.")
	thresholdkey := flag.String("thresholdkey", "", "The path to the threshold key file")
	id := flag.Int("id", 0, "The ID of this server.")
	flag.String("privkey", "", "The path to the ecdsa private key file used for TLS and HotStuff")

	//tls := flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *id == 0 {
		fmt.Printf("Please specify the id of this server.\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if *thresholdkey == "" {
		fmt.Printf("Please pass in a path to the threshold key.\n\n")
		flag.Usage()
		os.Exit(1)
	}
	/*
		if *tls {}
		if *privkey == "" {
			fmt.Printf("Please pass in a path to the private.\n\n")
			flag.Usage()
			os.Exit(1)
		}
	*/

	err := viper.BindPFlags(flag.CommandLine)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	viper.SetConfigName("hotcertification")
	//viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err = viper.ReadInConfig()
	if err != nil {
		log.Printf("Fatal error config file: %s \n", err)
		os.Exit(1)
	}

	// read config if not in working directory
	if *config != "" {
		viper.SetConfigFile(*config)
		err = viper.MergeInConfig()
		if err != nil {
			log.Printf("Fatal error config file: %s \n", err)
			os.Exit(1)
		}
	}

	var opts options.Options
	err = viper.Unmarshal(&opts)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	return &opts
}

func main() {
	/*
		1. Read keys from command line
		2. Store in server struct
		3. Start threshold signing backend
		4. Start client server
	*/
	opts := parseOptionsAndConfig()

	// so program can be stopped with CTRL+C
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	thresholdKey, err := crypto.ReadThresholdKeyFile(opts.ThresholdKey)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// TODO: What should channel capacity be?
	pendingCSRs := make(chan *client.CSR, 10)
	pendingCerts := make(chan *x509.Certificate, 10)
	finishedCerts := make(chan *x509.Certificate, 10)
	// signalling channel
	replicatedReqs := make(chan struct{}, 10)

	log.Println("Setting up servers.")

	// Parsing signing peer information
	signingPeers := make([]string, len(opts.Peers))
	for i, peer := range opts.Peers {
		signingPeers[i] = peer.SigningPeerAddr
	}

	// instantiating request buffer
	reqBuffer := hc.NewReqBuffer()

	replicationServer := replication.NewReplicationServer(opts, reqBuffer, replicatedReqs)
	signingServer := signing.NewSigningServer(thresholdKey, signingPeers, pendingCerts, finishedCerts)
	clientServer := client.NewClientServer(pendingCSRs, finishedCerts)

	go replicationServer.Start(ctx, opts.Peers[opts.ID-1].ReplicationPeerAddr)
	go signingServer.Start(opts.Peers[opts.ID-1].SigningPeerAddr)
	go clientServer.Start(opts.Peers[opts.ID-1].ClientAddr)

	// The logic for validating a csr and transforming into certificate
	// need root cert for this
	// This logic is only executed by the server that directly handles the client's request

	rootCA, err := crypto.ReadCertFile(opts.RootCA)
	if err != nil {
		log.Println(err)
	}

	csr := <-pendingCSRs

	log.Println("Replicating CSR")
	// replicate ?and validate?
	replicationServer.ReqBuffer.AddRequest(csr)
	// wait for replication finished
	<-replicatedReqs

	// starting signing process
	log.Println("Processing CSR.")
	x509_csr, err := x509.ParseCertificateRequest(csr.CSR)
	if err != nil {
		log.Println(err)
	}

	cert, err := crypto.GenerateCert(x509_csr, rootCA, thresholdKey)
	if err != nil {
		log.Println(err)
	}

	partialSigs, err := signingServer.CallGetPartialSig(cert)
	if err != nil {
		log.Println("Failed to collect enough partial signatures: ", err)
	}

	log.Println("Computing full signature for certificate.")
	fullCert, err := crypto.ComputeFullySignedCert(cert, thresholdKey, partialSigs...)
	if err != nil {
		log.Println("Failed to compute full signature: ", err)
	}

	finishedCerts <- fullCert

	<-ctx.Done()
}
