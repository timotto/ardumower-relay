package server_test

import (
	"code.cloudfoundry.org/lager"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/timotto/ardumower-relay/internal/server"
	"math/big"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Server Suite")
}

func aLogger() lager.Logger {
	return lager.NewLogger("test")
}

type fakeEndpoint struct {
	Response  string
	CallCount int
}

func (e *fakeEndpoint) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	e.CallCount++

	_, _ = w.Write([]byte(e.Response))
}

func httpUrl(s *server.Server, path string) string {
	address, _ := s.Address()
	//goland:noinspection HttpUrlsUsage
	return fmt.Sprintf("http://%v%v", address, path)
}

func httpsUrl(s *server.Server, path string) string {
	_, address := s.Address()
	return fmt.Sprintf("https://%v%v", address, path)
}

type tempPki struct {
	tempDir string
	Cert    string
	Key     string
}

func newTempPki() *tempPki {
	dir, err := os.MkdirTemp(os.TempDir(), "ardumower-relay-test-server-pki-*")
	Expect(err).ToNot(HaveOccurred())

	t := &tempPki{
		tempDir: dir,
		Cert:    path.Join(dir, "cert.pem"),
		Key:     path.Join(dir, "key.pem"),
	}

	Expect(generatePki(t.Key, t.Cert)).ToNot(HaveOccurred())

	return t
}

func (p *tempPki) Close() {
	_ = os.RemoveAll(p.tempDir)
}

func generatePki(keyFilename, certFilename string) error {
	pk, cert, err := generateSelfSignedCertificate("localhost", "ArduMower Relay Server")
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
