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
	"strconv"
	"strings"

	"github.com/relab/hotstuff"
	hsconfig "github.com/relab/hotstuff/config"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/raphasch/hotcertification/client"
	"github.com/raphasch/hotcertification/crypto"
	"github.com/raphasch/hotcertification/signing"
)

// Information other replicas in network have to know about each other (public knowledge)
// Private knowledge (threshold key, ecdsa private key) have to given through command line and not in global config file
type peer struct {
	ID                  hotstuff.ID
	PubKey              string `mapstructure:"pubkey"`
	TLSCert             string `mapstructure:"tls-cert"`
	ClientAddr          string `mapstructure:"client-address"`
	ReplicationPeerAddr string `mapstructure:"replication-peer-address"`
	SigningPeerAddr     string `mapstructure:"signing-peer-address"`
}

type options struct {
	// The ID of this server
	ID int `mapstructure:"id"`

	// TLS configs
	RootCA  string `mapstructure:"root-ca"`
	TLS     bool   `mapstructure:"tls"`
	PrivKey string `mapstructure:"privkey"` // privkey has to belong the to the pubkey and should be ecdsa because thresholdkey can't do TLS

	// HotStuff configs
	PmType   string      `mapstructure:"pacemaker"`
	LeaderID hotstuff.ID `mapstructure:"leader-id"`

	// HotCertification and miscellaneous configs
	ThresholdKey string `mapstructure:"thresholdkey"`
	KeySize      int    `mapstructure:"key-size"`
	ConfigFile   string `mapstructure:"config"`
	Peers        []peer
}

func usage() {
	fmt.Printf("Usage: %s [options]\n", os.Args[0])
	fmt.Println()
	fmt.Println("Loads configuration from ./hotcertification.yml and file specified by --config")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
}

func parseOptionsAndConfig() options {
	flag.Usage = usage

	help := flag.BoolP("help", "h", false, "Prints this text.")
	config := flag.String("config", "", "The path to the config file in case it isn't in working directory.")
	thresholdkey := flag.String("thresholdkey", "", "The path to the threshold key file")
	id := flag.Int("id", 0, "The ID of this server.")

	//tls := flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	//privkey := flag.String("privkey", "", "The path to the private key used for TLS")

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

	var opts options
	err = viper.Unmarshal(&opts)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	return opts
}

// TODO: parse peer info into hotstuff/config.ReplicaConfig and pass that struct into NewReplicationServer()
func getReplicaConfig(opts *options) *hsconfig.ReplicaConfig {
	return &hsconfig.ReplicaConfig{}
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
	pendingCSRs := make(chan *x509.CertificateRequest, 10)
	pendingCerts := make(chan *x509.Certificate, 10)
	finishedCerts := make(chan *x509.Certificate, 10)

	log.Println("Setting up servers.")

	// Parsing signing peer information
	signingPeers := make([]string, len(opts.Peers))
	for i, peer := range opts.Peers {
		signingPeers[i] = peer.SigningPeerAddr
	}
	signingServer := signing.NewSigningServer(thresholdKey, signingPeers, pendingCerts, finishedCerts)
	clientServer := client.NewClientServer(pendingCSRs, finishedCerts)

	signingPort, err := strconv.Atoi(strings.Split(opts.Peers[opts.ID-1].SigningPeerAddr, ":")[1])
	if err != nil {
		log.Println(err)
	}
	go signingServer.Start(signingPort)

	clientPort, err := strconv.Atoi(strings.Split(opts.Peers[opts.ID-1].ClientAddr, ":")[1])
	if err != nil {
		log.Println(err)
	}
	go clientServer.Start(clientPort)

	log.Println("Started server go routines.")

	// The logic for validating a csr and transforming into certificate
	// need root cert for this
	// This logic is only executed by the server that directly handles the client's request

	rootCA, err := crypto.ReadCertFile(opts.RootCA)
	if err != nil {
		log.Println(err)
	}

	csr := <-pendingCSRs
	log.Println("Processing CSR.")
	cert, err := crypto.GenerateCert(csr, rootCA, thresholdKey)
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
