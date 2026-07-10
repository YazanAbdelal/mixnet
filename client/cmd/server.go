package main

import (
	"log"
	"net"

	pb "github.com/YazanAbdelal/mixnet/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/YazanAbdelal/mixnet/client"
)

func runServer(mc *client.MixClient, creds credentials.TransportCredentials) {
	// start a listener on the given port
	lis, err := net.Listen("tcp", "0.0.0.0:50050")
	if err != nil {
		log.Fatal("Error opening port for listening: " + err.Error())
	}

	// create a new gRPC server and register it
	server := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterMessagingServer(server, mc)

	log.Println("listening on port " + mc.Port)

	// start handling remote calls from gRPC clinets
	if err := server.Serve(lis); err != nil {
		log.Fatal("client failed to serve: " + err.Error())
	}
}

func receiveMessages(mc *client.MixClient) {
	// print messages that were sent using the RPCs
	for msg := range mc.Pipe {
		log.Println("Printing that was received from the pipe.")
		log.Println(msg)
	}
}
