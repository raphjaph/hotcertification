package bls

import (
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"os"

	kyber "go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/bls"
	"go.dedis.ch/kyber/v3/sign/tbls"
)

const privateKeyFileType = "HOTCERTIFICATION AUTHORITY PRIVATE KEY SHARE"
const publicKeyFileType = "HOTCERTIFICATION AUTHORITY PUBLIC KEY"

type thresholdScheme struct {
	t       int
	n       int
	suite   pairing.Suite
	pubPoly *share.PubPoly
}

func NewThresholdScheme(t int, n int) *thresholdScheme {
	suite := bn256.NewSuite()
	ts := &thresholdScheme{
		t:     t,
		n:     n,
		suite: suite,
	}
	return ts
}

func (ts *thresholdScheme) SetPublicKey(pubkey *PublicKey) {
	ts.pubPoly = pubkey.pubPoly
}

func (ts *thresholdScheme) Sign(message []byte, key *PrivateKey) []byte {
	sig, _ := tbls.Sign(ts.suite, key.Share, message)
	return sig
}

func (ts *thresholdScheme) Verify(message []byte, signatures [][]byte) (bool bool, err error) {
	// Commit() because using bls to verify and it needs p(0) of the curve (the constant term of the polynomial)
	err = bls.Verify(ts.suite, ts.pubPoly.Commit(), message, ts.AggregateSignatures(message, signatures))
	if err != nil {
		return false, err
	}

	return true, nil
}

func (ts *thresholdScheme) VerifyPartialSig(message []byte, signature []byte) bool {
	return (nil == tbls.Verify(ts.suite, ts.pubPoly, message, signature))
}

func (ts *thresholdScheme) VerifyAggregate(message []byte, aggr_sigs []byte) bool {
	// Commit() because using bls to verify and it needs p(0) of the curve (the constant term of the polynomial)
	return (nil == bls.Verify(ts.suite, ts.pubPoly.Commit(), message, aggr_sigs))
}

func (ts *thresholdScheme) AggregateSignatures(message []byte, signatures [][]byte) []byte {
	aggr_sigs, _ := tbls.Recover(ts.suite, ts.pubPoly, message, signatures, ts.t, ts.n)
	return aggr_sigs
}

func (ts *thresholdScheme) ReadPrivateKeyFile(keyFile string) (privateKey *PrivateKey, err error) {
	block, err := readPemFile(keyFile)
	if err != nil {
		return nil, err
	}

	if block.Type != privateKeyFileType {
		return nil, fmt.Errorf("file type did not match")
	}

	shareIndex := binary.BigEndian.Uint16(block.Bytes[:2])
	// G1 or G2?
	scalar := ts.suite.G2().Scalar()
	err = scalar.UnmarshalBinary(block.Bytes[2:])

	privateKey = &PrivateKey{
		&share.PriShare{
			I: int(shareIndex),
			V: scalar,
		},
	}

	return privateKey, err
}

func (ts *thresholdScheme) ReadPublicKeyFile(keyFile string) (publicKey *PublicKey, err error) {
	block, err := readPemFile(keyFile)
	if err != nil {
		return nil, err
	}

	if block.Type != publicKeyFileType {
		return nil, fmt.Errorf("file type did not match")
	}

	polyCommit := ts.suite.G2().Point()
	err = polyCommit.UnmarshalBinary(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key")
	}

	publicKey = &PublicKey{
		share.NewPubPoly(ts.suite.G2(), ts.suite.G2().Point().Base(), []kyber.Point{polyCommit}),
	}

	return publicKey, err
}

// just wrappers around the private share and commitment polynomial
type PrivateKey struct {
	Share *share.PriShare
}
type PublicKey struct {
	pubPoly *share.PubPoly
}

func GenerateKeys(ts *thresholdScheme) (privKeys []*PrivateKey, pubKey *PublicKey) {
	secret := ts.suite.G1().Scalar().Pick(ts.suite.RandomStream())
	// polynomial of degree t (t-ten Grades); points on this curve become the key shares
	// in case secret == nil then suite.RandomStream generates new secret
	priPoly := share.NewPriPoly(ts.suite.G2(), ts.t, secret, ts.suite.RandomStream())
	// public key: commitment polynomial belonging to secret sharing polynomial
	pubPoly := priPoly.Commit(ts.suite.G2().Point().Base())
	pubKey = &PublicKey{pubPoly: pubPoly}

	privKeys = make([]*PrivateKey, ts.n)
	for i, share := range priPoly.Shares(ts.n) {
		privKeys[i] = &PrivateKey{Share: share}
	}

	return privKeys, pubKey
}

func WritePrivateKeyToFile(filePath string, key *PrivateKey) (err error) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	// This seems overly complicated; find better way
	// A private key has an index i and a value p(i) where p(x) is the secret polynomial(PriPoly)
	shareIndex := make([]byte, 2)
	binary.BigEndian.PutUint16(shareIndex, uint16(key.Share.I))
	scalar, err := key.Share.V.MarshalBinary()
	if err != nil {
		return err
	}
	marshalledBits := append(shareIndex, scalar...)

	block := &pem.Block{
		Type:  privateKeyFileType,
		Bytes: marshalledBits,
	}
	err = pem.Encode(file, block)

	if cerr := file.Close(); err == nil {
		err = cerr
	}
	return
}

func WritePublicKeyFile(filePath string, key *PublicKey) (err error) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return
	}

	defer func() {
		if cerr := file.Close(); err == nil {
			err = cerr
		}
	}()

	marshalledBits, err := key.pubPoly.Commit().MarshalBinary()
	if err != nil {
		return
	}

	block := &pem.Block{
		Type:  publicKeyFileType,
		Bytes: marshalledBits,
	}

	err = pem.Encode(file, block)
	return
}

// helper function
func readPemFile(file string) (b *pem.Block, err error) {
	d, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	b, _ = pem.Decode(d)
	if b == nil {
		return nil, fmt.Errorf("failed to decode PEM")
	}
	return b, nil
}
