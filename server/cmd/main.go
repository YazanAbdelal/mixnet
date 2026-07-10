package main

import (
	"log"
	"net"

	"github.com/YazanAbdelal/mixnet/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/YazanAbdelal/mixnet/proto/gen"
)

func runServer(s *server.MixServer, creds credentials.TransportCredentials) {
	// start a listener on the given port
	lis, err := net.Listen("tcp", "0.0.0.0:"+s.Port)
	if err != nil {
		log.Fatal("Error opening port for listening: " + err.Error())
	}

	// create a new gRPC server and register it
	server := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterMessagingServer(server, s)

	log.Println("listening on port " + s.Port)

	// start handling remote calls from gRPC clinets
	if err := server.Serve(lis); err != nil {
		log.Fatal("client failed to serve: " + err.Error())
	}
}

func main() {

	return
}
