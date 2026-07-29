package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/YazanAbdelal/mixnet/crypto"
	"github.com/YazanAbdelal/mixnet/mixnode"
)

const (
	PublicKeysPath = "/etc/mixnet/keys/public/"
)

type sendEntry struct {
	message string
	dest    string
}

// interactiveInput prompts the user for a message and a recipient from the
// available clients, then sends the entry to the returned channel.
func interactiveInput(clients []string) <-chan sendEntry {
	ch := make(chan sendEntry, 2)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("Enter message: ")
			msg, _ := reader.ReadString('\n')
			msg = strings.TrimSpace(msg)
			if msg == "" {
				continue
			}
			fmt.Println("\nAvailable recipients:")
			for i, c := range clients {
				fmt.Printf("  %d: %s\n", i+1, c)
			}
			fmt.Printf("Choose recipient (1-%d): ", len(clients))
			choiceStr, _ := reader.ReadString('\n')
			choiceStr = strings.TrimSpace(choiceStr)
			choice, err := strconv.Atoi(choiceStr)
			if err != nil || choice < 1 || choice > len(clients) {
				fmt.Println("Invalid choice")
				continue
			}
			ch <- sendEntry{message: msg, dest: clients[choice-1]}
		}
	}()
	return ch
}

// sendLoop reads message entries from the input channel, encrypts them as onion
// layers and forwards them through the mixnet. When no real input is ready it
// sends dummy cover traffic to a random client.
func sendLoop(ctx context.Context, mc *mixnode.MixNode, input <-chan sendEntry, servers []string, clients []string, pathLen int, clientTick time.Duration, cryptoType string) {
	ticker := time.NewTicker(clientTick)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			select {
			case entry := <-input:
				packet, firstNode, err := crypto.OnionEncrypt(entry.message, entry.dest, false, servers, pathLen, PublicKeysPath, cryptoType)
				if err != nil {
					log.Printf("Error encrypting message: %v", err)
					continue
				}
				sendToPeer(mc.Stubs[firstNode], packet)
			default:
				idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(clients))))
				dummyDest := clients[idx.Int64()]
				packet, firstNode, err := crypto.OnionEncrypt("__DUMMY__", dummyDest, true, servers, pathLen, PublicKeysPath, cryptoType)
				if err != nil {
					log.Printf("Error encrypting message: %v", err)
					continue
				}
				sendToPeer(mc.Stubs[firstNode], packet)
			}
		case <-ctx.Done():
			return
		}
	}
}
