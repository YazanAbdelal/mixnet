package server

import (
	pb "github.com/YazanAbdelal/mixnet/proto/gen"
)

type MixServer struct {
	pb.UnimplementedMessagingServer
	ID   int
	Port string
	Pipe chan string
}

func NewMixServer(id int, port string) *MixServer {
	return &MixServer{
		ID:   id,
		Port: port,
		Pipe: make(chan string),
	}
}
