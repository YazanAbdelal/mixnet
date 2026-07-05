package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"net"
	"os"

	"github.com/YazanAbdelal/mixnet/crypto"
	pb "github.com/YazanAbdelal/mixnet/proto/gen"
	"google.golang.org/grpc"
)

type MixClient struct {
	pb.UnimplementedMessagingServer
	ID   int
	Port string
	Pipe chan string
}

func NewMixClient(id int, port string) *MixClient {
	return &MixClient{
		ID:   id,
		Port: port,
		Pipe: make(chan string),
	}
}

func (c *MixClient) PrintMessage(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	log.Println(req.Payload)
	log.Println("This was printed from inside PrintMessage().")

	log.Println("Sending message to pipe.")
	c.Pipe <- req.Payload
	log.Println("Back inside PrintMessage()")

	response := &pb.MessageResponse{
		Success: true,
	}

	return response, nil
}

func main() {
	portPtr := flag.String("port", "50050", "port number")
	destPortPtr := flag.String("dest", "client2:50051", "destination port number")
	flag.Parse()

	client := NewMixClient(1, *portPtr)

	serverCreds, err := crypto.LoadServerTLSCredentials(
		"/etc/mixnet/certs/tls.crt",
		"/etc/mixnet/certs/tls.key",
		"/etc/mixnet/certs/ca.crt",
	)
	if err != nil {
		log.Fatal("Failed to load server TLS credentials: " + err.Error())
	}

	go func() {
		lis, err := net.Listen("tcp", "0.0.0.0:"+client.Port)
		if err != nil {
			log.Fatal("Error opening port for listening: " + err.Error())
			return
		}

		server := grpc.NewServer(grpc.Creds(serverCreds))

		pb.RegisterMessagingServer(server, client)

		log.Println("listening on port " + client.Port)

		if err := server.Serve(lis); err != nil {
			log.Fatal("client failed to serve: " + err.Error())
		}
	}()

	go func() {
		for msg := range client.Pipe {
			log.Println("Printing that was received from the pipe.")
			log.Println(msg)
		}
	}()

	clientCreds, err := crypto.LoadClientTLSCredentials(
		"/etc/mixnet/certs/tls.crt",
		"/etc/mixnet/certs/tls.key",
		"/etc/mixnet/certs/ca.crt",
	)
	if err != nil {
		log.Fatal("Failed to load client TLS credentials: " + err.Error())
	}

	conn, err := grpc.NewClient(*destPortPtr, grpc.WithTransportCredentials(clientCreds))
	if err != nil {
		log.Fatal("Could not create a new client: " + err.Error())
		return
	}
	stub := pb.NewMessagingClient(conn)

	userInput := make(chan string, 2)

	go func() {
		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			input := scanner.Text()
			if input != "" {
				userInput <- input
			}
		}

	}()

	for {
		input := <-userInput

		ctx := context.Background()
		req := &pb.MessageRequest{
			Payload: input,
		}

		_, err := stub.PrintMessage(ctx, req)
		if err != nil {
			log.Fatal(err.Error())
			return
		}
	}
}
