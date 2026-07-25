package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/YazanAbdelal/mixnet/mixnode"

	pb "github.com/YazanAbdelal/mixnet/proto/gen"
)

const ListeningPort = "50050"

// initStubs creates a stub for each of the nodes in the mixnet except the current node.
func initStubs(nodeName string, nodes []string) map[string]pb.MessagingClient {
	stubs := make(map[string]pb.MessagingClient)
	for _, node := range nodes {
		if node != nodeName {
			stubs[node] = connectToPeer(node+":"+ListeningPort, mustLoadClientCreds())
		}
	}
	return stubs
}

func main() {
	nodeName := flag.String("name", "", "node's name")
	destName := flag.String("dest", "", "recipient's name")
	nodeType := flag.String("type", "client", "node type (client or server)")
	cfgPath := flag.String("config", "/etc/mixnet/config.json", "path to config file")
	flag.Parse()

	cfg, err := mixnode.LoadConfig(*cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	nodes := append(cfg.Servers, cfg.Clients...)
	stubs := initStubs(*nodeName, nodes)

	// this context is for handling SIGINT and SIGTERM
	// we send it to the runServer function so it can exit gracefully (no messages missed) in case it receives any of these signals.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mc := mixnode.NewMixNode(ListeningPort, stubs)

	batchCh := make(chan mixnode.BatchEntry, 100)
	go receiveMessages(mc, *nodeType, batchCh)

	if *nodeType == "client" {
		clientTick := time.Duration(cfg.ClientTickUs) * time.Microsecond
		go sendLoop(mc, *destName, readStdin(), cfg.Servers, cfg.PathLen, clientTick)
	} else {
		flushTimeout := time.Duration(cfg.FlushTimeoutMs) * time.Millisecond
		go batchFlusher(stubs, batchCh, cfg.BatchSize, flushTimeout)
	}

	server := runServer(mc, mustLoadServerCreds())
	<-ctx.Done()
	log.Println("Shutting down...")

	// wait until all messages are received and then shuts down
	server.GracefulStop()
}
