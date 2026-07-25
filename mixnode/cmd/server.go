package main

import (
	"log"
	"net"

	pb "github.com/YazanAbdelal/mixnet/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/YazanAbdelal/mixnet/crypto"
	"github.com/YazanAbdelal/mixnet/mixnode"
)

// runServer starts listening to tcp requests on port 50050 and handles gRPS calls through a mTLS-secured channel.
func runServer(mc *mixnode.MixNode, creds credentials.TransportCredentials) *grpc.Server {
	// start a listener on the given port
	lis, err := net.Listen("tcp", "0.0.0.0:"+ListeningPort)
	if err != nil {
		log.Fatal("Error opening port for listening: " + err.Error())
	}

	// create a new gRPC server and register it
	server := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterMessagingServer(server, mc)

	log.Println("listening on port " + mc.Port)

	go func() {
		// start handling remote calls from gRPC clients
		if err := server.Serve(lis); err != nil {
			log.Fatal("client failed to serve: " + err.Error())
		}
	}()

	return server
}

// receiveMessages receives gRPC messages from peers, decrypts them and forwards them to the next node if there is one.
func receiveMessages(mc *mixnode.MixNode, nodeType string, batchCh chan<- mixnode.BatchEntry) {
	for msg := range mc.Pipe {
		// decrypt onion layer
		decryptedMsg, nextNode, isDummy, err := crypto.DecryptLayer(msg, "/etc/mixnet/keys/private.pem")
		if err != nil {
			log.Fatal("receiveMessages: Error decrypting onion layer: " + err.Error())
			return
		}

		// if this is the last node, print message or drop it.
		if nextNode == "" {
			// if dummy and this is the last node, drop the message
			if isDummy {
				continue
			}

			// if not dummy, log message.
			log.Printf("Received the following message: %q.", string(decryptedMsg))

		} else { // if this is not the last node, forward to the next node
			// send packet to next node using its stub
			if nodeType == "server" {
				log.Printf("Received a packet with size = %v.", len(decryptedMsg))
				batchCh <- mixnode.BatchEntry{Packet: decryptedMsg, NextNode: nextNode}
			} else {
				sendToPeer(mc.Stubs[nextNode], decryptedMsg)
			}
		}
	}
}
