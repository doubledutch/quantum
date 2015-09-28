package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

var clientCipherSuites = []uint16{
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
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
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS10,
		CipherSuites: clientCipherSuites,
		RootCAs:      certPool,
	}, nil
}
