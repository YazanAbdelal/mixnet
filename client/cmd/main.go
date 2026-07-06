package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"net"
	"os"

	pb "github.com/YazanAbdelal/mixnet/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/YazanAbdelal/mixnet/client"
)

func runServer(mc *client.MixClient, creds credentials.TransportCredentials) {
	// start a listener on the given port
	lis, err := net.Listen("tcp", "0.0.0.0:"+mc.Port)
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

func connectToPeer(dest string, creds credentials.TransportCredentials) pb.MessagingClient {
	// start a gRPC client, and connect to dest through a TLS encrypted channel
	conn, err := grpc.NewClient(dest, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatal("Could not create a new client: " + err.Error())
	}
	return pb.NewMessagingClient(conn)
}

func readStdin() <-chan string {
	ch := make(chan string, 2)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if input := scanner.Text(); input != "" {
				ch <- input
			}
		}
		close(ch)
	}()
	return ch
}

func sendLoop(stub pb.MessagingClient, input <-chan string) {
	for msg := range input {
		_, err := stub.PrintMessage(context.Background(), &pb.MessageRequest{Payload: msg})
		if err != nil {
			log.Fatal(err.Error())
		}
	}
	log.Println("stdin reached EOF, send loop exiting")
}

func main() {
	port := flag.String("port", "50050", "port number")
	dest := flag.String("dest", "client-2:50050", "destination address")
	flag.Parse()

	mc := client.NewMixClient(1, *port)

	go receiveMessages(mc)

	stub := connectToPeer(*dest, mustLoadClientCreds())
	go sendLoop(stub, readStdin())

	runServer(mc, mustLoadServerCreds())
}
