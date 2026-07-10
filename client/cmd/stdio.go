package main

import (
	"bufio"
	"context"
	"log"
	"os"

	pb "github.com/YazanAbdelal/mixnet/proto/gen"
)

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

func sendLoop(stub pb.MessagingClient, input <-chan string) {
	// keep reading from channel until it is closed.
	// each message will be forwarded to the destination node, using a stub and calling a procedure remotely (RPC)
	for msg := range input {
		// TODO should encrypt and add padding before sending
		_, err := stub.ForwardMessage(context.Background(), &pb.MessageRequest{
			Payload: []byte(msg),
		})
		if err != nil {
			log.Fatal(err.Error())
		}
	}
	log.Println("stdin reached EOF, send loop exiting")
}
