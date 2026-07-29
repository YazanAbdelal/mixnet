package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/YazanAbdelal/mixnet/keygen"
)

const (
	RSAKeySize = 4096
)

// generateRSAKeysForNode generates an RSA key pair for a single node
// and exports both the private and public PEM files.
func generateRSAKeysForNode(nodeType string, index int, keysFolder, publicKeysFolder string) error {
	privateKeyPath := nodeType + "-" + strconv.Itoa(index) + "-rsa-private.pem"
	publicKetPath := nodeType + "-" + strconv.Itoa(index) + "-rsa-public.pem"
	private, public, err := keygen.GenerateRSAKeys(RSAKeySize)
	if err != nil {
		return fmt.Errorf("error generating RSA keys: %w", err)
	}
	err = keygen.ExportPrivateRSAKey(private, keysFolder, privateKeyPath)
	if err != nil {
		return fmt.Errorf("error exporting RSA private key: %w", err)
	}
	return keygen.ExportPublicRSAKey(public, publicKeysFolder, publicKetPath)
}

// generateECCKeysForNode generates an ECC key pair for a single node
// and exports both the private and public PEM files.
func generateECCKeysForNode(nodeType string, index int, keysFolder, publicKeysFolder string) error {
	privateKeyPath := nodeType + "-" + strconv.Itoa(index) + "-ecc-private.pem"
	publicKetPath := nodeType + "-" + strconv.Itoa(index) + "-ecc-public.pem"
	private, public, err := keygen.GenerateECCKeys()
	if err != nil {
		return fmt.Errorf("error generating ECC keys: %w", err)
	}
	err = keygen.ExportPrivateECCKey(private, keysFolder, privateKeyPath)
	if err != nil {
		return fmt.Errorf("error exporting ECC private key: %w", err)
	}
	return keygen.ExportPublicECCKey(public, publicKeysFolder, publicKetPath)
}

func main() {
	clientCount := flag.Int("clients", 2, "number of clients")
	serverCount := flag.Int("servers", 3, "number of mix nodes")
	cryptoType := flag.String("crypto-type", "both", "key type to generate: rsa, ecc, or both")

	flag.Parse()

	fmt.Printf("Generating %s key pairs for %d clients and %d servers... ",
		*cryptoType, *clientCount, *serverCount)

	keysFolder := "keys"
	publicKeysFolder := keysFolder + "/public"

	for i := range *clientCount {
		index := i + 1
		switch *cryptoType {
		case "ecc":
			if err := generateECCKeysForNode("client", index, keysFolder, publicKeysFolder); err != nil {
				fmt.Println(err)
				return
			}
		case "rsa":
			if err := generateRSAKeysForNode("client", index, keysFolder, publicKeysFolder); err != nil {
				fmt.Println(err)
				return
			}
		default:
			if err := generateRSAKeysForNode("client", index, keysFolder, publicKeysFolder); err != nil {
				fmt.Println(err)
				return
			}
			if err := generateECCKeysForNode("client", index, keysFolder, publicKeysFolder); err != nil {
				fmt.Println(err)
				return
			}
		}
	}

	for i := range *serverCount {
		index := i + 1
		switch *cryptoType {
		case "ecc":
			if err := generateECCKeysForNode("server", index, keysFolder, publicKeysFolder); err != nil {
				fmt.Println(err)
				return
			}
		case "rsa":
			if err := generateRSAKeysForNode("server", index, keysFolder, publicKeysFolder); err != nil {
				fmt.Println(err)
				return
			}
		default:
			if err := generateRSAKeysForNode("server", index, keysFolder, publicKeysFolder); err != nil {
				fmt.Println(err)
				return
			}
			if err := generateECCKeysForNode("server", index, keysFolder, publicKeysFolder); err != nil {
				fmt.Println(err)
				return
			}
		}
	}

	fmt.Print("Done.\n")

}
