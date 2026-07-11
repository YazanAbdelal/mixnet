package main

import (
	"log"

	"github.com/YazanAbdelal/mixnet/crypto"
	"google.golang.org/grpc/credentials"
)

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
