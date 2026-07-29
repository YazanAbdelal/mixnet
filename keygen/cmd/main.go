package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/YazanAbdelal/mixnet/keygen"
)

const (
	KeySize = 4096
)

func main() {
	clientCount := flag.Int("clients", 2, "number of clients")
	serverCount := flag.Int("servers", 3, "number of mix nodes")

	flag.Parse()

	fmt.Printf("Generating %d pairs of keys for the clients and %d pairs of keys for the servers... ", *clientCount, *serverCount)

	keysFolder := "keys"
	publicKeysFolder := keysFolder + "/public"
	for i := range *clientCount {
		privateKeyPath := "client-" + strconv.Itoa(i+1) + "-rsa-private.pem"
		publicKetPath := "client-" + strconv.Itoa(i+1) + "-rsa-public.pem"
		private, public, err := keygen.GenerateRSAKeys(KeySize)
		if err != nil {
			fmt.Print("Error generating keys: " + err.Error())
			return
		}
		err = keygen.ExportPrivateRSAKey(private, keysFolder, privateKeyPath)
		if err != nil {
			fmt.Print("Error exporting private key: " + err.Error())
			return
		}
		err = keygen.ExportPublicRSAKey(public, publicKeysFolder, publicKetPath)
		if err != nil {
			fmt.Print("Error exporting public key: " + err.Error())
			return
		}
	}

	for i := range *serverCount {
		privateKeyPath := "server-" + strconv.Itoa(i+1) + "-rsa-private.pem"
		publicKetPath := "server-" + strconv.Itoa(i+1) + "-rsa-public.pem"
		private, public, err := keygen.GenerateRSAKeys(KeySize)
		if err != nil {
			fmt.Print("Error generating keys: " + err.Error())
			return
		}
		err = keygen.ExportPrivateRSAKey(private, keysFolder, privateKeyPath)
		if err != nil {
			fmt.Print("Error exporting private key: " + err.Error())
			return
		}
		err = keygen.ExportPublicRSAKey(public, publicKeysFolder, publicKetPath)
		if err != nil {
			fmt.Print("Error exporting public key: " + err.Error())
			return
		}
	}

	fmt.Printf("Done.\n")
}
