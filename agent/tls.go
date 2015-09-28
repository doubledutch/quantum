package agent

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

var acceptedCBCCiphers = []uint16{
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	tls.TLS_RSA_WITH_AES_128_CBC_SHA,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
}

// ErrProvideAllCertFiles is used when one of the cert files isn't provided
var ErrProvideAllCertFiles = errors.New("Must provide cert, key, and CA certificate files")

// NewTLSConfigFromEnv creates a new TLS configuration using environment variables
func NewTLSConfigFromEnv() (*tls.Config, error) {
	certFile := os.Getenv("TLS_CERT")
	keyFile := os.Getenv("TLS_KEY")
	caFile := os.Getenv("TLS_CA")
	if certFile == "" || keyFile == "" || caFile == "" {
		return nil, ErrProvideAllCertFiles
	}
	return NewTLSConfig(certFile, keyFile, caFile)
}

// NewTLSConfig creates a new TLS configuration
func NewTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	pem, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("Could not read CA certificate %q: %v", caFile, err)
	}
	if !certPool.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("failed to append certificates from PEM file: %q", caFile)
	}

	return &tls.Config{
		ClientAuth:               tls.RequireAndVerifyClientCert,
		ClientCAs:                certPool,
		Certificates:             []tls.Certificate{cert},
		MinVersion:               tls.VersionTLS10,
		PreferServerCipherSuites: true,
		CipherSuites:             acceptedCBCCiphers,
		RootCAs:                  certPool,
	}, nil
}
