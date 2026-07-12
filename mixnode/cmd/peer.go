package main

import (
	"context"
	"log"
	"time"

	pb "github.com/YazanAbdelal/mixnet/proto/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func connectToPeer(dest string, creds credentials.TransportCredentials) pb.MessagingClient {
	// start a gRPC client, and connect to dest through a TLS encrypted channel
	conn, err := grpc.NewClient(dest, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatal("connectToPeer: erro creating a new client: " + err.Error())
	}
	return pb.NewMessagingClient(conn)
}

func sendToPeer(stub pb.MessagingClient, packet []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err := stub.ForwardMessage(ctx, &pb.MessageRequest{
		Payload: []byte(packet),
	})
	if err != nil {
		log.Fatal("sendToPeer: error forwarding message to peer: " + err.Error())
	}
}
