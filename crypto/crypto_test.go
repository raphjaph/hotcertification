package crypto_test

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/niclabs/tcrsa"

	"github.com/raphasch/hotcertification/crypto"
)

const (
	threshold   = 3
	num         = 4
	keySize     = 512
	destination = "test_keys"
)

func TestWriteReadEquality(t *testing.T) {

	thresholdKeys, err := crypto.ComputeTresholdKeys(threshold, num, keySize)
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}

	err = os.MkdirAll(destination, 0755)
	if err != nil {
		panic(fmt.Errorf("cannot create '%s' directory: %w", destination, err))
	}

	// Write keys to file
	for i, key := range thresholdKeys {
		thresholdKeyPath := filepath.Join(destination, fmt.Sprintf("n%v.thresholdkey", i+1))
		err = crypto.WriteThresholdKeyFile(key, thresholdKeyPath)
		if err != nil {
			panic(fmt.Sprintf("%v", err))
		}
	}

	// Read keys from file and compare
	readThresholdKeys := make([]*crypto.ThresholdKey, len(thresholdKeys))
	for i, key := range thresholdKeys {
		thresholdKeyPath := filepath.Join(destination, fmt.Sprintf("n%v.thresholdkey", i+1))
		readThresholdKeys[i], err = crypto.ReadThresholdKeyFile(thresholdKeyPath)
		if err != nil {
			panic(fmt.Sprintf("%v", err))
		}
		if key.KeyShare.Id != readThresholdKeys[i].KeyShare.Id {
			t.Errorf("key shares ids are different.")
		}

		for i, b := range readThresholdKeys[i].KeyShare.Si {
			if key.KeyShare.Si[i] != b {
				t.Errorf("key shares are different.")
			}
		}
		keySize := key.KeyMeta.PublicKey.Size()
		readKeySize := readThresholdKeys[i].KeyMeta.PublicKey.Size()
		if keySize != readKeySize {
			t.Errorf("public key sizes not the same: %v vs. %v", keySize, readKeySize)
		}
	}
}

func TestSigningProcess(t *testing.T) {

	fmt.Println("Generating configuration and writing to files")
	err := crypto.GenerateConfiguration(threshold, num, keySize, destination)
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}

	fmt.Println("New client")
	clientPrivKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}

	fmt.Println("New CSR for client")
	csr, err := crypto.GenerateCSR(clientPrivKey)
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}

	fmt.Println("Reading root cert")
	rootCA, err := crypto.ReadCertFile(destination + "/root.crt")
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}

	fmt.Println("Reading dummy key")
	dummyKey, err := crypto.ReadThresholdKeyFile(destination + "/n1.thresholdkey")
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}

	fmt.Println("Generating a certificate for signing")
	cert, err := crypto.GenerateCert(csr, rootCA, dummyKey)
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}

	sigShares := make(tcrsa.SigShareList, threshold)
	fmt.Println("Starting partially signing test")
	for i := 0; i < threshold; i++ {

		fmt.Print("Reading a key from file -> ")
		thresholdKey, err := crypto.ReadThresholdKeyFile(destination + fmt.Sprintf("/n%v.thresholdkey", i+1))
		if err != nil {
			panic(fmt.Sprintf("%v", err))
		}

		fmt.Println("Partially signing with key")
		sigShares[i], err = crypto.ComputePartialSignature(cert, thresholdKey)
		if err != nil {
			panic(fmt.Sprintf("%v", err))
		}
		/*
			certHash := sha256.Sum256(cert.RawTBSCertificate)
			certPKCS1, err := tcrsa.PrepareDocumentHash(thresholdKey.KeyMeta.PublicKey.Size(), thresholdKey.HashType, certHash[:])
			if err != nil {
				panic(fmt.Sprintf("%v", err))
			}

			sigShares[i], err = thresholdKey.KeyShare.Sign(certPKCS1, thresholdKey.HashType, thresholdKey.KeyMeta)
			if err != nil {
				panic(fmt.Sprintf("%v", err))
			}
			if err := sigShares[i].Verify(certPKCS1, thresholdKey.KeyMeta); err != nil {
				panic(fmt.Sprintf("%v", err))
			}
		*/
	}

	fmt.Println("Computing the full signature on the certificate")
	fullCert, err := crypto.ComputeFullySignedCert(cert, dummyKey, sigShares...)
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}

	fmt.Println("Verifying signature on certificate")
	if fullCert.CheckSignatureFrom(rootCA) != nil {
		fmt.Println(fullCert.CheckSignatureFrom(rootCA))
		t.Errorf("signature verification failed")
	}
}
