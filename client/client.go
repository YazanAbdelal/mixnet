package client

import (
	"context"
	"log"

	pb "github.com/YazanAbdelal/mixnet/proto/gen"
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
