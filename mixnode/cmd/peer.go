package main

import (
	"context"
	"log"
	"time"

	pb "github.com/YazanAbdelal/mixnet/proto/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

const sendRetries = 3

// connectToPeer connects to gRPC server and returns a stub for calling methods on the server remotely.
func connectToPeer(dest string, creds credentials.TransportCredentials) pb.MessagingClient {
	// start a gRPC client, and connect to dest through a TLS encrypted channel
	conn, err := grpc.NewClient(dest, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatal("connectToPeer: erro creating a new client: " + err.Error())
	}
	return pb.NewMessagingClient(conn)
}

// sendToPeer calls the ForwardMessage to a gRPC server using the stub.
// retries up to sendRetries times with backoff before dropping the packet.
func sendToPeer(stub pb.MessagingClient, packet []byte) {
	var err error
	for attempt := range sendRetries {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		_, err = stub.ForwardMessage(ctx, &pb.MessageRequest{
			Payload: []byte(packet),
		})
		cancel()
		if err == nil {
			return
		}
		if attempt < sendRetries-1 {
			backoff := time.Duration(100*(1<<attempt)) * time.Millisecond
			log.Printf("sendToPeer: attempt %d failed, retrying in %v: %v", attempt+1, backoff, err)
			time.Sleep(backoff)
		}
	}
	log.Printf("sendToPeer: all %d attempts failed, dropping packet: %v", sendRetries, err)
	if st, ok := status.FromError(err); ok {
		log.Printf("sendToPeer: gRPC status code=%v message=%q", st.Code(), st.Message())
	}
}
