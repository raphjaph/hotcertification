package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"io"

	"github.com/niclabs/tcrsa"
)

/*
This data structure has two purposes:

	1. Implementing the "crypto" Signer interface for generating certificates with the "crypto/x509" package
 	2. Hold all the information needed for a replica to generate a partial signature on a certificate

The dummy private key data and the Sign() method just server purpose 1. This is so I can piggy back on the
"crypto/x509" library for generating/parsing/marshalling x509 certificates. As part of this the library signs
the certificate but since I'm using a custom threshold signing scheme I can't use that. So I slice out the
signature part and TBS(to be signed) part of the certificate and later add it back with the signature
generated through the "github.com/niclabs/tcrsa" library.
*/
type ThresholdKey struct {
	KeyShare     *tcrsa.KeyShare
	KeyMeta      *tcrsa.KeyMeta
	HashType     crypto.Hash
	DummyPrivKey *rsa.PrivateKey
}

func NewThresholdKey(keyShare *tcrsa.KeyShare, keyMeta *tcrsa.KeyMeta, keySize int) (*ThresholdKey, error) {

	dummyPrivKey, err := rsa.GenerateKey(rand.Reader, keySize)

	return &ThresholdKey{
		KeyShare:     keyShare,
		KeyMeta:      keyMeta,
		HashType:     crypto.SHA256,
		DummyPrivKey: dummyPrivKey,
	}, err
}

func (key *ThresholdKey) Public() *rsa.PublicKey {
	return key.KeyMeta.PublicKey
}

func (key *ThresholdKey) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	return rsa.SignPKCS1v15(rand, key.DummyPrivKey, key.HashType, digest)
}
