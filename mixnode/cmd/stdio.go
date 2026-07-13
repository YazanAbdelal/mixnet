package main

import (
	"bufio"
	"log"
	"os"
	"time"

	"github.com/YazanAbdelal/mixnet/crypto"
	"github.com/YazanAbdelal/mixnet/mixnode"
)

const (
	PublicKeysPath = "/etc/mixnet/keys/public/"
)

// readStdin reads from the standard input and sends the messages it receives to the sendLoop function through a receive only channel.
func readStdin() <-chan string {

	ch := make(chan string, 2)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)

		// read from the stdin and send to channel. the channel will forward to the sendLoop function, which in turn remotely calls (RPC) the target's
		// function. The loop exists when an error occurs or when stdin is closed (EOF), or if the buffer exceeds its capacity.
		for scanner.Scan() {
			if input := scanner.Text(); input != "" {
				ch <- input
			}
		}
		// if scanner exited because of an error, print error log
		if err := scanner.Err(); err != nil {
			log.Printf("Error reading stdin: %v\n", err)
		}

		close(ch) // closing the channel causes the sendLoop function to return.
	}()
	return ch
}

// sendLoop reads messages (string) from the stdin, encrypts the message with onion and then forwards them to the next node in the mixnet using a gRPC stub.
func sendLoop(mc *mixnode.MixNode, dest string, input <-chan string, servers []string, pathLen int, clientTick time.Duration) {
	ticker := time.NewTicker(clientTick)
	for range ticker.C {
		select {
		case msg := <-input:
			packet, firstNode, err := crypto.OnionEncrypt(msg, dest, false, servers, pathLen, PublicKeysPath)
			if err != nil {
				log.Printf("Error encrypting message: %v", err)
				continue
			}
			sendToPeer(mc.Stubs[firstNode], packet)
		default:
			packet, firstNode, err := crypto.OnionEncrypt("__DUMMY__", dest, true, servers, pathLen, PublicKeysPath)
			if err != nil {
				log.Printf("Error encrypting message: %v", err)
				continue
			}
			sendToPeer(mc.Stubs[firstNode], packet)
		}

	}
}
