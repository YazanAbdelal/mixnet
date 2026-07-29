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
	Port        string
	Pipe        chan []byte // for receiving packets
	Stubs       map[string]pb.MessagingClient
	ReplayCache *ReplayCache
}

func NewMixNode(ctx context.Context, port string, stubs map[string]pb.MessagingClient) *MixNode {
	return &MixNode{
		Port:        port,
		Pipe:        make(chan []byte, 100),
		Stubs:       stubs,
		ReplayCache: NewReplayCache(ctx),
	}
}

// ForwardMessage is called by gRPC clients.
// receives the request and forwards the payload through a channel to be handled elsewhere.
func (c *MixNode) ForwardMessage(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// drop replays silently -- always return success so the caller
	// cannot distinguish a replay from a legitimate packet
	if c.ReplayCache != nil && c.ReplayCache.IsReplay(req.Payload) {
		return &pb.MessageResponse{Success: true}, nil
	}

	c.Pipe <- req.Payload

	return &pb.MessageResponse{Success: true}, nil
}
