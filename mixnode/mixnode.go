package mixnode

import (
	"context"

	pb "github.com/YazanAbdelal/mixnet/proto/gen"
)

type BatchEntry struct {
	Packet   []byte
	NextNode string
}

type MixNode struct {
	pb.UnimplementedMessagingServer
	Port  string
	Pipe  chan []byte // for receiving packets
	Stubs map[string]pb.MessagingClient
}

func NewMixNode(port string, stubs map[string]pb.MessagingClient) *MixNode {
	return &MixNode{
		Port:  port,
		Pipe:  make(chan []byte, 100),
		Stubs: stubs,
	}
}

// ForwardMessage is called by gRPC clients.
// receives the request and forwards the payload through a channel to be handled elsewhere.
func (c *MixNode) ForwardMessage(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {
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
