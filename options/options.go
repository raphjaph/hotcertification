package options

import (
	"github.com/relab/hotstuff"
)

// Information other replicas in network have to know about each other (public knowledge)
// Private knowledge (threshold key, ecdsa private key) have to given through command line and not in global config file
type Peer struct {
	ID                  hotstuff.ID
	PubKey              string `mapstructure:"pubkey"`
	TLSCert             string `mapstructure:"tls-cert"`
	ClientAddr          string `mapstructure:"client-address"`
	ReplicationPeerAddr string `mapstructure:"replication-peer-address"`
	SigningPeerAddr     string `mapstructure:"signing-peer-address"`
}

type Options struct {
	// The ID of this server
	ID hotstuff.ID `mapstructure:"id"`

	// TLS configs
	RootCA  string `mapstructure:"root-ca"`
	TLS     bool   `mapstructure:"tls"`
	PrivKey string `mapstructure:"privkey"` // privkey has to belong the to the pubkey and should be ecdsa because thresholdkey can't do TLS

	// HotStuff configs
	PmType      string      `mapstructure:"pacemaker"`
	LeaderID    hotstuff.ID `mapstructure:"leader-id"`
	ViewTimeout int         `mapstructure:"view-timeout"`

	// HotCertification and miscellaneous configs
	ThresholdKey string `mapstructure:"thresholdkey"`
	KeySize      int    `mapstructure:"key-size"`
	ConfigFile   string `mapstructure:"config"`
	Peers        []Peer
}
