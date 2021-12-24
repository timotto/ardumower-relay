package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"
)

//go:generate go run .

func main() {
	if err := run(keyFilename, certFilename); err != nil {
		log.Fatal(err)
	}
}

const (
	dnsName      = "localhost"
	subjectOrg   = "ArduMower Relay Server"
	certFilename = "cert.example.pem"
	keyFilename  = "key.example.pem"
)

func run(keyFilename, certFilename string) error {
	pk, cert, err := generateSelfSignedCertificate(dnsName, subjectOrg)
	if err != nil {
		return err
	}

	if err := writePemFile(keyFilename, "PRIVATE KEY", 0600, pk); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	if err := writePemFile(certFilename, "CERTIFICATE", 0644, cert); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	return nil
}

func generateSelfSignedCertificate(dnsName, subjectOrg string) ([]byte, []byte, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	pk, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal private key: %v", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{subjectOrg},
		},
		DNSNames:  []string{dnsName},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(8760 * time.Hour),

		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},

		BasicConstraintsValid: true,
	}

	cert, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	return pk, cert, nil
}

func writePemFile(filename, blockType string, perm os.FileMode, data []byte) error {
	block, err := encodeSinglePemBlock(blockType, data)
	if err != nil {
		return err
	}

	log.Printf("writing %v to %v", blockType, filename)
	if err := os.WriteFile(filename, block, perm); err != nil {
		return err
	}

	return nil
}

func encodeSinglePemBlock(blockType string, data []byte) ([]byte, error) {
	if pemData := pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: data}); pemData == nil {
		return nil, fmt.Errorf("PEM encoding failed")
	} else {
		return pemData, nil
	}
}
