package main

import (
	"log"

	pb "github.com/YazanAbdelal/mixnet/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func connectToPeer(dest string, creds credentials.TransportCredentials) pb.MessagingClient {
	// start a gRPC client, and connect to dest through a TLS encrypted channel
	conn, err := grpc.NewClient(dest, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatal("Could not create a new client: " + err.Error())
	}
	return pb.NewMessagingClient(conn)
}
