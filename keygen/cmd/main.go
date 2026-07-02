package main

import (
	"flag"
	"fmt"
	"strconv"

	"https://github.com/YazanAbdelal/mixnet/keygen"
)

func main() {
	clientCount := flag.Int("clients", 2, "number of clients")
	serverCount := flag.Int("servers", 3, "number of mix nodes")

	flag.Parse()

	fmt.Printf("Generating %d pairs of keys for the clients and %d pairs of keys for the servers...", *clientCount, *serverCount)

	keysFolder := "keys"
	for i := range *clientCount {
		privateKeyPath := "client_" + strconv.Itoa(i+1) + "_private.pem"
		publicKetPath := "client_" + strconv.Itoa(i+1) + "_public.pem"
		private, public, err := keygen.GenerateKeys()
	}
}
