package main

import (
	"log"
	"net"

	pb "github.com/YazanAbdelal/mixnet/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/YazanAbdelal/mixnet/client"
	"github.com/YazanAbdelal/mixnet/crypto"
)

// runServer starts listening to tcp requests on port 50050 and handles gRPS calls through a mTLS-secured channel.
func runServer(mc *client.MixClient, creds credentials.TransportCredentials) {
	// start a listener on the given port
	lis, err := net.Listen("tcp", "0.0.0.0:50050")
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

// receiveMessages receives gRPC messages from peers, decrypts them and forwards them to the next node if there is one.
func receiveMessages(mc *client.MixClient) {
	for msg := range mc.Pipe {
		// decrypt onion layer
		decryptedMsg, nextNode, err := crypto.DecryptLayer(msg, "/etc/mixnet/keys/private.pem")
		if err != nil {
			log.Fatal("receiveMessages: Error decrypting onion layer: " + err.Error())
			return
		}
		// if this is the last node, print message and return.
		if nextNode == "" {
			log.Printf("Recieved the following message: %q.", string(decryptedMsg))
		} else { // if this is not the last node, forward to the next node
			// create a stub
			stub := connectToPeer(nextNode+":50050", mustLoadClientCreds())
			// send packet to next node using the stub
			sendToPeer(stub, decryptedMsg)
		}
	}
}
