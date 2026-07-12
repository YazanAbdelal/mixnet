package main

import (
	"log"

	"github.com/YazanAbdelal/mixnet/crypto"
	"google.golang.org/grpc/credentials"
)

// mustLoadServerCreds loads TLS credintials for the server.
func mustLoadServerCreds() credentials.TransportCredentials {
	creds, err := crypto.LoadServerTLSCredentials(
		"/etc/mixnet/certs/tls.crt",
		"/etc/mixnet/certs/tls.key",
		"/etc/mixnet/certs/ca.crt",
	)
	if err != nil {
		log.Fatal("Failed to load server TLS credentials: " + err.Error())
	}
	return creds
}

// mustLoadClientCreds loads TLS credintials for the client.
func mustLoadClientCreds() credentials.TransportCredentials {
	creds, err := crypto.LoadClientTLSCredentials(
		"/etc/mixnet/certs/tls.crt",
		"/etc/mixnet/certs/tls.key",
		"/etc/mixnet/certs/ca.crt",
	)
	if err != nil {
		log.Fatal("Failed to load client TLS credentials: " + err.Error())
	}
	return creds
}
