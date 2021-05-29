/*
	KEY GENERATION EXECUTABLE:
		1. Key shares computed and wrapped into compatible data structure (see threshold_key.go)
		2. Root certificate computed and self-signed by replicas

			The CA Root Certificate is a root certificate signed by itself (a threshold of replicas/nodes of the network).
			This is done in the trusted setup phase (initialization/key generation).
			Every saves this certificate to issue new partially signed certificates.
			The Gateway/Aggregator then combines these partial signatures and issues a completely normal looking certificate to the client

		3. Create TLS certificates for secure communication between replicas with root certificate
		4. Computes HotStuff ecdsa keys for replication/consensus logic
		5. Marshall/Write root cert, tls certs, hotstuff priv and pub keys and threshold keys to files
*/

// TODO: add keygen functions for HotStuff

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/relab/hotstuff/crypto/keygen"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/raphasch/hotcertification/crypto"
)

type options struct {
	Num         uint16 `mapstructure:"num"`
	Threshold   uint16 `mapstructure:"threshold"`
	KeySize     int    `mapstructure:"key-size"`
	Destination string
	//TLS     bool // if TLS enabled then also generates tls certificates signed by root CA and writes to file. TODO:
}

func usage() {
	fmt.Printf("Usage: %s [options] [destination]\n", os.Args[0])
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
}

func parseOptions() (options, string) {

	flag.Usage = usage
	help := flag.BoolP("help", "h", false, "Prints this text.")
	flag.Uint16P("num", "n", 3, "The number of replicas to generate keys/certs for.")
	flag.Uint16P("threshold", "t", 2, "The threshold of replicas that can generate a valid signature on a certificate.")
	flag.Int("key-size", 512, "The size of the RSA private key from which the threshold keys are generated. Has to be one of 512/1024/2048/4096.")
	flag.Parse()

	if flag.NArg() < 1 {
		usage()
		os.Exit(1)
	}

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

	return opts, flag.Arg(0)

}

func main() {
	opts, dest := parseOptions()

	fmt.Println("Generating all threshold keys and root certificate.")

	// Generates the threshold keys and a root certificate and writes all to seperate files
	err := crypto.GenerateConfiguration(opts.Threshold, opts.Num, opts.KeySize, dest)
	if err != nil {
		fmt.Print(err)
	}

	fmt.Println("Generating all private keys and TLS certificates.")

	// Generates ecdsa private keys for HotStuff/Replication server
	err = keygen.GenerateConfiguration(dest, false, false, 1, int(opts.Num), "n*", []string{"localhost:8080"})
	if err != nil {
		fmt.Print(err)
	}
}
