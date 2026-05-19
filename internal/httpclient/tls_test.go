package httpclient_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/httpclient"
)

// makeCAPEM generates a self-signed CA cert and returns its PEM bytes.
func makeCAPEM(t *testing.T) []byte {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "qctx-test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	require.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func TestNewWithValidCACertLoads(t *testing.T) {
	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.pem")
	require.NoError(t, os.WriteFile(caPath, makeCAPEM(t), 0o600))

	c, err := httpclient.New(httpclient.Options{CACertPath: caPath, Timeout: 2 * time.Second})
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestNewWithMissingCACertErrors(t *testing.T) {
	_, err := httpclient.New(httpclient.Options{CACertPath: "/no/such/ca.pem", Timeout: 2 * time.Second})
	require.Error(t, err)
}

func TestNewWithInvalidPEMErrors(t *testing.T) {
	dir := t.TempDir()
	caPath := filepath.Join(dir, "bad.pem")
	require.NoError(t, os.WriteFile(caPath, []byte("not a pem"), 0o600))

	_, err := httpclient.New(httpclient.Options{CACertPath: caPath, Timeout: 2 * time.Second})
	require.Error(t, err)
}
