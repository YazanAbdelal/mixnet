package main

import (
	"flag"

	"github.com/YazanAbdelal/mixnet/client"
)

func main() {
	// parse user options
	port := flag.String("port", "50050", "port number")
	dest := flag.String("dest", "client-2:50050", "destination address")
	flag.Parse()

	// create a new client
	mc := client.NewMixClient(1, *port)

	// start a thread for printing messages recieved using RPC
	go receiveMessages(mc)

	// start a thread for reading from stdin and forwarding messages using RPC
	stub := connectToPeer(*dest, mustLoadClientCreds())
	go sendLoop(stub, readStdin())

	// start the gRPC server in the main thread (forwards to recieveMessages thread)
	runServer(mc, mustLoadServerCreds())
}
