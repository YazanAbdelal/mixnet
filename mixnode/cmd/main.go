package main

import (
	"flag"

	"github.com/YazanAbdelal/mixnet/mixnode"

	pb "github.com/YazanAbdelal/mixnet/proto/gen"
)

const (
	BatchSize     = 10
	BatchInterval = 200
	ListeningPort = "50050"
)

var nodes = []string{"client-1", "client-2", "server-1", "server-2", "server-3"}

// initStubs creates a stub for each of the nodes in the mixnet except the current node.
func initStubs(nodeName string) map[string]pb.MessagingClient {
	stubs := make(map[string]pb.MessagingClient)
	for _, node := range nodes {
		if node != nodeName { // we do not want to add a stub from the current node to itself.
			stubs[node] = connectToPeer(node+":"+ListeningPort, mustLoadClientCreds())
		}
	}
	return stubs
}

func main() {
	nodeName := flag.String("name", "", "node's name")
	destName := flag.String("dest", "", "recipient's name")
	nodeType := flag.String("type", "client", "node type (client or server)")
	flag.Parse()
	// make stubs to all the other nodes
	stubs := initStubs(*nodeName)

	// create a new client
	mc := mixnode.NewMixNode(ListeningPort, stubs)

	// start a thread for printing (or storing) or forwarding messages recieved using RPC
	batchCh := make(chan mixnode.BatchEntry, 100)
	go receiveMessages(mc, *nodeType, batchCh)

	// only clients can send messages:
	if *nodeType == "client" {
		// sending messages to dest like this:
		// user inputs message using stdin   ---channel--->   sendLoop   ---stub--->   calls ForwardMessage method on the dest
		// then the dest either prints (or stores) it, or forwards it to the next node if it is a mixnet.
		go sendLoop(mc, *destName, readStdin()) // readStdin opens a channel from stdin to the sendLoop function
	} else {
		go batchFlusher(stubs, batchCh)
	}

	// start the gRPC server in the main thread (forwards to recieveMessages thread)
	runServer(mc, mustLoadServerCreds())
}
