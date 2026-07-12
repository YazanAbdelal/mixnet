package crypto

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"

	"google.golang.org/grpc/credentials"
)

// LoadServerTLSCredentials loads TLS certificate for the server.
func LoadServerTLSCredentials(certFile, keyFile, caFile string) (credentials.TransportCredentials, error) {
	serverCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, errors.New("LoadServerTLSCredentials: " + err.Error())
	}

	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, errors.New("LoadServerTLSCredentials: " + err.Error())
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, errors.New("LoadServerTLSCredentials: failed to parse CA certificate")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(tlsConfig), nil
}

// LoadClientTLSCredentials loads TLS cerificates for the client.
func LoadClientTLSCredentials(certFile, keyFile, caFile string) (credentials.TransportCredentials, error) {
	clientCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, errors.New("LoadClientTLSCredentials: " + err.Error())
	}

	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, errors.New("LoadClientTLSCredentials: " + err.Error())
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, errors.New("LoadClientTLSCredentials: failed to parse CA certificate")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	return credentials.NewTLS(tlsConfig), nil
}
