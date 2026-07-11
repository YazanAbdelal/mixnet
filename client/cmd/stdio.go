package main

import (
	"bufio"
	"log"
	"os"

	"github.com/YazanAbdelal/mixnet/client"
	"github.com/YazanAbdelal/mixnet/crypto"
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
func sendLoop(client *client.MixNode, dest string, input <-chan string) {
	// keep reading from channel until it is closed.
	// each message will be forwarded to the destination node, using a stub and calling a procedure remotely (RPC)
	for msg := range input {
		// encrypt the message using onion encryption and pad
		// OnionEncrypt chooses a routing path and returns the first node in the path
		packet, firstNode, err := crypto.OnionEncrypt(msg, dest)
		if err != nil {
			log.Fatal("Error encrypting message: " + err.Error())
		}

		// forward the message to the first node
		sendToPeer(client.Stubs[firstNode], packet)
	}
	log.Println("stdin reached EOF, send loop exiting")
}
