package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/relab/hotstuff"
	"github.com/relab/hotstuff/config"
	"github.com/relab/hotstuff/crypto"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/grpc/credentials"
)

type replica struct {
	ID         hotstuff.ID
	PeerAddr   string `mapstructure:"peer-address"`
	ClientAddr string `mapstructure:"client-address"`
	Pubkey     string
	Cert       string
}

type options struct {
	BatchSize       int         `mapstructure:"batch-size"`
	Benchmark       bool        `mapstructure:"benchmark"`
	Cert            string      `mapstructure:"cert"`
	ClientAddr      string      `mapstructure:"client-listen"`
	ExitAfter       int         `mapstructure:"exit-after"`
	Input           string      `mapstructure:"input"`
	LeaderID        hotstuff.ID `mapstructure:"leader-id"`
	MaxInflight     uint64      `mapstructure:"max-inflight"`
	Output          string      `mapstructure:"print-commands"`
	PayloadSize     int         `mapstructure:"payload-size"`
	PeerAddr        string      `mapstructure:"peer-listen"`
	PmType          string      `mapstructure:"pacemaker"`
	PrintThroughput bool        `mapstructure:"print-throughput"`
	Privkey         string
	RateLimit       int         `mapstructure:"rate-limit"`
	RootCAs         []string    `mapstructure:"root-cas"`
	SelfID          hotstuff.ID `mapstructure:"self-id"`
	TLS             bool
	ViewTimeout     int `mapstructure:"view-timeout"`
	Replicas        []replica
}

func usage() {
	fmt.Printf("Usage: %s [options]\n", os.Args[0])
	fmt.Println()
	fmt.Println("Loads configuration from ./hotstuff.toml and file specified by --config")
	fmt.Println()
	fmt.Println("Options:")
	pflag.PrintDefaults()
}

func main() {
	pflag.Usage = usage

	// some configuration options can be set using flags
	help := pflag.BoolP("help", "h", false, "Prints this text.")
	configFile := pflag.String("config", "", "The path to the config file")
	//server := pflag.Bool("server", false, "Start a server. If not specified, a client will be started.")

	// shared options
	pflag.Uint32("self-id", 0, "The id for this replica.")
	pflag.Bool("tls", false, "Enable TLS")
	pflag.Int("exit-after", 0, "Number of seconds after which the program should exit.")

	// server options
	pflag.String("privkey", "", "The path to the private key file (server)")
	pflag.String("cert", "", "Path to the certificate (server)")
	pflag.Int("view-timeout", 1000, "How many milliseconds before a view is timed out (server)")
	pflag.Int("batch-size", 100, "How many commands are batched together for each proposal (server)")
	pflag.Bool("print-throughput", false, "Throughput measurements will be printed stdout (server)")
	pflag.String("client-listen", "", "Override the listen address for the client server (server)")
	pflag.String("peer-listen", "", "Override the listen address for the replica (peer) server (server)")

	// client options
	pflag.String("input", "", "Optional file to use for payload data (client)")
	pflag.Bool("benchmark", false, "If enabled, a BenchmarkData protobuf will be written to stdout. (client)")
	pflag.Int("rate-limit", 0, "Limit the request-rate to approximately (in requests per second). (client)")
	pflag.Int("payload-size", 0, "The size of the payload in bytes (client)")
	pflag.Uint64("max-inflight", 10000, "The maximum number of messages that the client can wait for at once (client)")

	pflag.Parse()

	if *help {
		pflag.Usage()
		os.Exit(0)
	}

	var conf options
	err := ReadConfig(&conf, *configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read config: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go func() {
		if conf.ExitAfter > 0 {
			time.Sleep(time.Duration(conf.ExitAfter) * time.Millisecond)
			cancel()
		}
	}()

	runServer(ctx, &conf)
	//runClient(ctx, &conf)
}

func runServer(ctx context.Context, conf *options) {
	privkey, err := crypto.ReadPrivateKeyFile(conf.Privkey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read private key file: %v\n", err)
		os.Exit(1)
	}

	var creds credentials.TransportCredentials
	var tlsCert tls.Certificate
	if conf.TLS {
		creds, tlsCert = loadCreds(conf)
	}

	var clientAddress string
	replicaConfig := config.NewConfig(conf.SelfID, privkey, creds)
	for _, r := range conf.Replicas {
		key, err := crypto.ReadPublicKeyFile(r.Pubkey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read public key file '%s': %v\n", r.Pubkey, err)
			os.Exit(1)
		}

		info := &config.ReplicaInfo{
			ID:      r.ID,
			Address: r.PeerAddr,
			PubKey:  key,
		}

		if r.ID == conf.SelfID {
			// override own addresses if set
			if conf.ClientAddr != "" {
				clientAddress = conf.ClientAddr
			} else {
				clientAddress = r.ClientAddr
			}
			if conf.PeerAddr != "" {
				info.Address = conf.PeerAddr
			}
		}

		replicaConfig.Replicas[r.ID] = info
	}

	srv := newCertificationServer(conf, replicaConfig, &tlsCert)
	err = srv.Start(clientAddress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start HotStuff: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("starting server", conf.SelfID, "on", clientAddress)

	<-ctx.Done()
	srv.Stop()
}

func loadCreds(conf *options) (credentials.TransportCredentials, tls.Certificate) {
	if conf.Cert == "" {
		for _, replica := range conf.Replicas {
			if replica.ID == conf.SelfID {
				conf.Cert = replica.Cert
			}
		}
	}

	tlsCert, err := tls.LoadX509KeyPair(conf.Cert, conf.Privkey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse certificate: %v\n", err)
		os.Exit(1)
	}

	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		// system cert pool is unavailable on windows
		if runtime.GOOS != "windows" {
			fmt.Fprintf(os.Stderr, "Failed to get system cert pool: %v\n", err)
		}
		rootCAs = x509.NewCertPool()
	}

	for _, ca := range conf.RootCAs {
		cert, err := os.ReadFile(ca)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read CA file: %v\n", err)
			os.Exit(1)
		}
		if !rootCAs.AppendCertsFromPEM(cert) {
			fmt.Fprintf(os.Stderr, "Failed to add CA to cert pool.\n")
			os.Exit(1)
		}
	}
	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		RootCAs:      rootCAs,
		ClientCAs:    rootCAs,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	})

	return creds, tlsCert
}

// ReadConfig reads config options from configuration files and command line flags.
func ReadConfig(opts interface{}, secondaryConfig string) (err error) {
	err = viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return err
	}

	// read main config file in working dir
	viper.SetConfigName("hotstuff")
	viper.AddConfigPath(".")
	err = viper.ReadInConfig()
	if err != nil {
		return err
	}

	// read secondary config if requested
	if secondaryConfig != "" {
		viper.SetConfigFile(secondaryConfig)
		err = viper.MergeInConfig()
		if err != nil {
			return err
		}
	}

	err = viper.Unmarshal(opts)
	if err != nil {
		return err
	}

	return nil
}
