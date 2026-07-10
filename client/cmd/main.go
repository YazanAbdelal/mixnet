package main

import (
	"flag"

	"github.com/YazanAbdelal/mixnet/client"
)

func main() {
	destName := flag.String("dest", "client-2", "recipient's name")
	flag.Parse()

	// create a new client
	mc := client.NewMixClient(1, "50050")

	// start a thread for printing (or storing) or forwarding messages recieved using RPC
	go receiveMessages(mc)

	// start a thread for reading from stdin and forwarding messages using RPC
	// first node in the mix is always "server-1"
	stub := connectToPeer("server-1:50050", mustLoadClientCreds())

	// sending messages to dest like this:
	// user inputs message using stdin   ---channel--->   sendLoop   ---stub--->   calls ForwardMessage method on the dest
	// then the dest either prints (or stores) it, or forwards it to the next node if it is a mixnet.
	go sendLoop(stub, *destName, readStdin()) // readStdin opens a channel from stdin to the sendLoop function

	// start the gRPC server in the main thread (forwards to recieveMessages thread)
	runServer(mc, mustLoadServerCreds())
}
