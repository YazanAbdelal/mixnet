package client

import (
	"context"

	pb "github.com/YazanAbdelal/mixnet/proto/gen"
)

type MixClient struct {
	pb.UnimplementedMessagingServer
	ID   int
	Port string
	Pipe chan []byte
}

func NewMixClient(id int, port string) *MixClient {
	return &MixClient{
		ID:   id,
		Port: port,
		Pipe: make(chan []byte),
	}
}

func (c *MixClient) ForwardMessage(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// forward the mesage to the Pipe
	c.Pipe <- req.Payload

	response := &pb.MessageResponse{
		Success: true,
	}

	return response, nil
}
