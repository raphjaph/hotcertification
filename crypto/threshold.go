package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/niclabs/tcrsa"
	"google.golang.org/protobuf/proto"

	serial "github.com/raphasch/hotcertification/crypto/serialization"
)

const thresholdKeyFileType = "HOTCERTIFICATION THRESHOLD KEY"

const keySize = 512

func ComputePartialSignature(certificate *x509.Certificate, key *ThresholdKey) (partialSig *tcrsa.SigShare, err error) {
	// extract the RawTBSCertificate and sign that
	// TBS = to be signed
	tbsCert := certificate.RawTBSCertificate

	// hash the certificate data
	certHash := sha256.Sum256(tbsCert)
	// padding hash to conform to PKCS1 standard
	certPaddedHash, err := tcrsa.PrepareDocumentHash(key.KeyMeta.PublicKey.Size(), key.HashType, certHash[:])
	if err != nil {
		return nil, err
	}

	partialSig, err = key.KeyShare.Sign(certPaddedHash, key.HashType, key.KeyMeta)

	// TODO: Do I need this? Kick out in optimization
	if err := partialSig.Verify(certPaddedHash, key.KeyMeta); err != nil {
		return nil, err
	}

	return partialSig, err
}

func ComputeFullySignedCert(certificate *x509.Certificate, key *ThresholdKey, partialSigs ...*tcrsa.SigShare) (*x509.Certificate, error) {
	// extract the RawTBSCertificate and sign that
	// TBS = to be signed
	tbsCert := certificate.RawTBSCertificate

	// hash the certificate data
	certHash := sha256.Sum256(tbsCert)
	// padding hash to conform to PKCS1 standard
	certPaddedHash, err := tcrsa.PrepareDocumentHash(key.KeyMeta.PublicKey.Size(), key.HashType, certHash[:])
	if err != nil {
		return nil, err
	}

	var signatures tcrsa.SigShareList = partialSigs

	// TODO: This throws an error
	// verify partial signature
	for _, share := range signatures {
		if err := share.Verify(certPaddedHash, key.KeyMeta); err != nil {
			return nil, err
		}
	}

	signature, err := signatures.Join(certPaddedHash, key.KeyMeta)
	// insert the full rsa signature
	certificate.Signature = signature
	certificate.SignatureAlgorithm = x509.SHA256WithRSA

	return certificate, err
}

func ComputeTresholdKeys(threshold uint16, n uint16, keySize int) (thresholdKeys []*ThresholdKey, err error) {
	// trusted dealer computes key shares
	keyShares, keyMeta, err := tcrsa.NewKey(keySize, threshold, n, nil)
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
	thresholdKeys = make([]*ThresholdKey, n)
	for i, share := range keyShares {
		// wrapping key in the crypto/signer interface
		thresholdKeys[i], err = NewThresholdKey(share, keyMeta, keySize)
		if err != nil {
			return nil, err
		}
	}
	return thresholdKeys, err
}

func GenerateCert(csr *x509.CertificateRequest, issuer *x509.Certificate, issuerKey *ThresholdKey) (cert *x509.Certificate, err error) {
	sn, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	cert = &x509.Certificate{
		SerialNumber:          sn,
		Issuer:                pkix.Name{CommonName: "HotCertification Authority"},
		SignatureAlgorithm:    csr.SignatureAlgorithm,
		PublicKeyAlgorithm:    csr.PublicKeyAlgorithm,
		Subject:               csr.Subject,
		EmailAddresses:        csr.EmailAddresses,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
	}

	bytes, err := x509.CreateCertificate(rand.Reader, cert, issuer, csr.PublicKey, issuerKey.DummyPrivKey)
	if err != nil {
		fmt.Println(err)
	}

	return x509.ParseCertificate(bytes)
}

func GenerateCSR(clientPrivKey *rsa.PrivateKey) (cert *x509.CertificateRequest, err error) {
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

func GenerateRootCert(key *ThresholdKey) (cert *x509.Certificate, err error) {
	sn, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}

	caTmpl := &x509.Certificate{
		SerialNumber:          sn,
		Subject:               pkix.Name{CommonName: "HotCertificationAuthority"},
		Issuer:                pkix.Name{CommonName: "HotCertificationAuthority"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	// self-signed with dummyPrivKey
	// TODO: sign through threshold signing mechanism (tcrsa); see above
	caBytes, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, key.Public(), key.DummyPrivKey)
	if err != nil {
		return nil, err
	}

	return x509.ParseCertificate(caBytes)
}

func KeyToBytes(key *ThresholdKey) ([]byte, error) {

	// I'm using protocol buffers to parse the key into bytes and then later unmarshall it back into its data structure
	protobufKey := &serial.ThresholdKey{
		KeyShare: &serial.KeyShare{
			Si: key.KeyShare.Si,
			Id: uint32(key.KeyShare.Id),
		},
		KeyMeta: &serial.KeyMeta{
			PublicKey: &serial.RSAPublicKey{
				N: key.KeyMeta.PublicKey.N.Bytes(),
				E: int32(key.KeyMeta.PublicKey.E),
			},
			K: uint32(key.KeyMeta.K),
			L: uint32(key.KeyMeta.L),
			VerificationKey: &serial.VerificationKey{
				V: key.KeyMeta.VerificationKey.V,
				U: key.KeyMeta.VerificationKey.U,
				I: key.KeyMeta.VerificationKey.I,
			},
		},
	}

	return proto.Marshal(protobufKey)
}

func KeyFromBytes(bytes []byte, keySize int) (*ThresholdKey, error) {

	key := &serial.ThresholdKey{}
	err := proto.Unmarshal(bytes, key)
	if err != nil {
		return nil, err
	}

	keyShare := &tcrsa.KeyShare{
		Si: key.KeyShare.Si,
		Id: uint16(key.KeyShare.Id),
	}

	N := big.NewInt(8)
	N.SetBytes(key.KeyMeta.PublicKey.N)

	keyMeta := &tcrsa.KeyMeta{
		PublicKey: &rsa.PublicKey{
			N: N,
			E: int(key.KeyMeta.PublicKey.E),
		},
		K: uint16(key.KeyMeta.K),
		L: uint16(key.KeyMeta.L),
		VerificationKey: &tcrsa.VerificationKey{
			V: key.KeyMeta.VerificationKey.V,
			U: key.KeyMeta.VerificationKey.U,
			I: key.KeyMeta.VerificationKey.I,
		},
	}

	return NewThresholdKey(keyShare, keyMeta, keySize)
}

func WriteThresholdKeyFile(key *ThresholdKey, filePath string) (err error) {
	// make so it creates directories as well

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return
	}

	defer func() {
		if cerr := file.Close(); err == nil {
			err = cerr
		}
	}()

	bytes, err := KeyToBytes(key)
	if err != nil {
		return
	}

	block := &pem.Block{
		Type:  thresholdKeyFileType,
		Bytes: bytes,
	}

	err = pem.Encode(file, block)
	if err != nil {
		return err
	}

	return
}

func WriteCertFile(cert *x509.Certificate, filePath string) (err error) {
	// make so it creates directories as well
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return
	}

	defer func() {
		if cerr := file.Close(); err == nil {
			err = cerr
		}
	}()

	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}

	return pem.Encode(file, block)
}

func ReadThresholdKeyFile(keyFile string) (*ThresholdKey, error) {
	raw, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM")
	}

	if block.Type != thresholdKeyFileType {
		return nil, fmt.Errorf("file type did not match")
	}

	key, err := KeyFromBytes(block.Bytes, keySize)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func ReadCertFile(certFile string) (cert *x509.Certificate, err error) {
	raw, err := os.ReadFile(certFile)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, fmt.Errorf("failed to decode key")
	}

	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("file type did not match")
	}

	cert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

/*
	TRUSTED SETUP LOGIC:
		1. Key shares computed and wrapped into compatible data structure (see threshold_key.go)
		2. Root certificate computed and self-signed by replicas
		3. Create TLS certificates for secure communication between replicas with root certificate
		4. Computes HotStuff ecdsa keys for replication/consensus logic
		5. Marshall/Write root cert, tls certs, hotstuff priv and pub keys and threshold keys to files
*/

/*
This generates the necessary keys and certificates for issuing x509 certificates.
	a. Generate threshold keys and write to files
	b. Geneate the self-signed certificate of the configuration/group that represents the CA
	c. Write that CA Certificate to file

*/

// put this function into threshold.go ?
func GenerateConfiguration(t uint16, n uint16, keySize int, destination string) (err error) {
	// create directory for keys
	err = os.MkdirAll(destination, 0755)
	if err != nil {
		return fmt.Errorf("cannot create '%s' directory: %w", destination, err)
	}

	// trusted dealer computes key shares
	thresholdKeys, err := ComputeTresholdKeys(t, n, keySize)
	if err != nil {
		return err
	}

	// Write threshold keys to files
	for i, key := range thresholdKeys {

		thresholdKeyPath := filepath.Join(destination, fmt.Sprintf("p%v.thresholdkey", i+1))
		err = WriteThresholdKeyFile(key, thresholdKeyPath)
		if err != nil {
			return err
		}
	}

	// wrapping key in the crypto/signer interface
	caKey := thresholdKeys[0]

	// create a Root certificate for TLS (self-signed)
	caRootCertificate, err := GenerateRootCert(caKey)
	if err != nil {
		return err
	}

	sigShares := make(tcrsa.SigShareList, t)
	for i := 0; i < int(t); i++ {
		sigShares[i], err = ComputePartialSignature(caRootCertificate, thresholdKeys[i])
		if err != nil {
			return err
		}
	}

	/*
		The CA Root Certificate is a root certificate signed by itself (a threshold of replicas/nodes of the network).
		This is done in the trusted setup phase (initialization/key generation).
		Every saves this certificate to issue new partially signed certificates.
		The Gateway/Aggregator then combines these partial signatures and issues a completely normal looking certificate to the client.
	*/
	caRootCertificate, err = ComputeFullySignedCert(caRootCertificate, caKey, sigShares...)
	if err != nil {
		return err
	}

	err = WriteCertFile(caRootCertificate, destination+"/root.crt")
	if err != nil {
		return err
	}
	return
}
