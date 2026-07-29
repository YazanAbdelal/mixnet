package crypto

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

const (
	MessageSize                  = 4096
	RSAKeySizeBytes              = 512
	AESKeySize                   = 32
	ECCKeySize                   = 32
	ECCMetadataBlockSize         = 64
	ECCMetadataBlockSizeAfterGCM = 92 // 12 (nonce) + 16 (tag) = 28 Bytes overhead, and 64 Bytes for the actual metadata
	LengthOfCiphertextSize       = 2
	LengthOfNextNodeSize         = 2
	DummyFlagLSize               = 1
)

// padMessage pads a message with random bytes.
// returns a padded message with size = MessageSize.
func padMessage(msg []byte) ([]byte, error) {
	// calculate padding size
	padSize := MessageSize - len(msg)

	// generate random padding
	padding := make([]byte, padSize)
	_, err := rand.Read(padding)
	if err != nil {
		return nil, errors.New("padMessage: error adding padding: " + err.Error())
	}

	// pad message
	paddedMessage := append(msg, padding...)

	return paddedMessage, nil
}

// getRandomPath selects a random path with the given length from the given servers.
func getRandomPath(servers []string, dest string, pathLen int) []string {
	n := min(pathLen, len(servers))
	shuffled := make([]string, len(servers))
	copy(shuffled, servers)

	for i := len(shuffled) - 1; i > 0; i-- {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		idx := j.Int64()
		shuffled[i], shuffled[idx] = shuffled[idx], shuffled[i]
	}

	return append(shuffled[:n], dest, "")
}

// OnionEncrypt encrypts a message with onion layers and pads it.
// returns a packet with size = MessageSize, and the first node in the mix.
func OnionEncrypt(msg string, dest string, isDummy bool, servers []string, pathLen int, publicKeysPath string, cryptoType string) ([]byte, string, error) {
	msgBytes := []byte(msg)

	randomPath := getRandomPath(servers, dest, pathLen)

	// encrypt inside-out
	var err error
	for i := len(randomPath) - 1; i >= 1; i-- {
		nextNode := randomPath[i]
		switch cryptoType {
		case "rsa":
			path := publicKeysPath + randomPath[i-1] + "-rsa-public.pem"
			msgBytes, err = encryptOnionLayerWithRSA(msgBytes, nextNode, path, isDummy)

		default:
			path := publicKeysPath + randomPath[i-1] + "-ecc-public.pem"
			msgBytes, err = EncryptOnionLayerWithECC(msgBytes, nextNode, path, isDummy)
		}
		if err != nil {
			return nil, "", fmt.Errorf("OnionEncrypt: layer %d: %w", i, err)
		}
	}

	// pad message
	paddedMessage, err := padMessage(msgBytes)
	if err != nil {
		return nil, "", errors.New("OnionEncrypt: " + err.Error())
	}

	return paddedMessage, randomPath[0], nil
}
