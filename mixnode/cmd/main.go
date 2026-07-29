package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
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
	cryptoType := flag.String("crypto", "rsa", "crypto type (rsa or ecc)")
	flag.Parse()

	cfg, err := mixnode.LoadConfig(*cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	var privKeyPath string
	if *cryptoType == "ecc" {
		privKeyPath = "/etc/mixnet/keys/ecc_private.pem"
	} else {
		privKeyPath = "/etc/mixnet/keys/private.pem"
	}

	nodes := append(cfg.Servers, cfg.Clients...)
	stubs := initStubs(*nodeName, nodes)

	// this context is for handling SIGINT and SIGTERM
	// we send it to the goroutines so they can exit gracefully (no messages missed) in case the process receives any of these signals.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	mc := mixnode.NewMixNode(ctx, ListeningPort, stubs)

	batchCh := make(chan mixnode.BatchEntry, 100)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		receiveMessages(ctx, mc, *nodeType, batchCh, *cryptoType, privKeyPath)
	}()

	if *nodeType == "client" {
		clientTick := time.Duration(cfg.ClientTickUs) * time.Microsecond
		wg.Add(1)
		go func() {
			defer wg.Done()
			sendLoop(ctx, mc, *destName, readStdin(), cfg.Servers, cfg.PathLen, clientTick, *cryptoType)
		}()
	} else {
		flushTimeout := time.Duration(cfg.FlushTimeoutMs) * time.Millisecond
		wg.Add(1)
		go func() {
			defer wg.Done()
			batchFlusher(ctx, stubs, batchCh, cfg.BatchSize, flushTimeout)
		}()
	}

	server := runServer(mc, mustLoadServerCreds())
	<-sigCh
	log.Println("Shutting down...")

	// wait until all messages are received and then shuts down
	server.GracefulStop()
	cancel()
	wg.Wait()
}
