package mixnode

import (
	"context"
	"log"

	pb "github.com/YazanAbdelal/mixnet/proto/gen"
)

type MixNode struct {
	pb.UnimplementedMessagingServer
	Port  string
	Pipe  chan []byte
	Stubs map[string]pb.MessagingClient
}

func NewMixNode(port string, stubs map[string]pb.MessagingClient) *MixNode {
	return &MixNode{
		Port:  port,
		Pipe:  make(chan []byte),
		Stubs: stubs,
	}
}

func (c *MixNode) ForwardMessage(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	log.Printf("Received a Packet of length: %v.", len(req.Payload))

	// forward the mesage to the Pipe
	c.Pipe <- req.Payload

	response := &pb.MessageResponse{
		Success: true,
	}

	return response, nil
}
