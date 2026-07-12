package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"

	random "math/rand"

	"github.com/YazanAbdelal/mixnet/keygen"
)

const (
	MessageSize            = 4096
	RSAKeySizeBytes        = 512
	AESKeySize             = 32
	LengthOfCiphertextSize = 2
	DummyFlagLSize         = 1
)

// encryptOnionLayer adds a layer of encryption to the onion using hybrid encryption.
// it returns a ciphertext that includes an ephemeral AES key, the length of the encrypted content, the next node, and the AES encrypted content.
func encryptOnionLayer(content []byte, nextNode string, pathToKey string, isDummy bool) ([]byte, error) {
	// first we generate an ephemeral AES key
	aesKey, err := genAESKey(AESKeySize)
	if err != nil {
		return nil, errors.New("encryptOnionLayer: error generating AES key: " + err.Error())
	}

	// then we encrypt the content using the AES key
	ciphertext, err := encryptWithAES(content, aesKey)
	if err != nil {
		return nil, errors.New("encryptOnionLayer: error encrypting using AES key: " + err.Error())
	}

	// then we use the node's public RSA key to encrypt the the length of the AES key, the length of the ciphertext, the dummy flag and the next destination
	// first we concatenate them
	rsaPlaintext := make([]byte, 0, AESKeySize+LengthOfCiphertextSize+DummyFlagLSize+len(nextNode)) // length = 0, capacity = AESKeySize + LengthOfCiphertextSize + len(nextNode)
	rsaPlaintext = append(rsaPlaintext, aesKey...)
	rsaPlaintext = binary.BigEndian.AppendUint16(rsaPlaintext, uint16(len(ciphertext)))
	if isDummy {
		rsaPlaintext = append(rsaPlaintext, 1)
	} else {
		rsaPlaintext = append(rsaPlaintext, 0)
	}
	rsaPlaintext = append(rsaPlaintext, []byte(nextNode)...)

	// now we load the public RSA key of the target node
	rsaKey, err := keygen.LoadPublicKey(pathToKey)
	if err != nil {
		return nil, errors.New("encryptOnionLayer: error loading public key: " + err.Error())
	}

	// now we encrypt using the target node's public RSA key
	// the resulting ciphertext here is 512 bytes long because we used a 4096 bit long RSA key
	rsaCiphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaKey, rsaPlaintext, nil)
	if err != nil {
		return nil, errors.New("encryptOnionLayer: error encrypting with RSA key: " + err.Error())
	}

	return append(rsaCiphertext, ciphertext...), nil
}

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

// getRandomPath takes a slice of server names and the destination's name, shuffles the servers and adds the destination to the path.
func getRandomPath(servers []string, dest string) []string {
	// shuffle the servers
	random.Shuffle(len(servers), func(i int, j int) {
		servers[i], servers[j] = servers[j], servers[i]
	})

	// add the dest and the empty string to the end
	return append(servers, dest, "")
}

// OnionEncrypt encrypts a message with 4 onion layers and pads it.
// returns a packet with size = MessageSize, and the first node in the mix.
func OnionEncrypt(msg string, dest string, isDummy bool) ([]byte, string, error) {
	// convert into bytes before
	msgBytes := []byte(msg)

	// choose path randomly
	servers := []string{"server-1", "server-2", "server-3"}
	randomPath := getRandomPath(servers, dest)

	// encrypt inside-out
	var err error
	for i := len(randomPath) - 1; i >= 1; i-- {
		nextNode := randomPath[i]
		path := "/etc/mixnet/keys/public/" + randomPath[i-1] + "-public.pem"
		msgBytes, err = encryptOnionLayer(msgBytes, nextNode, path, isDummy)
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

// DecryptLayer decrypts an onion layer using the private RSA key of the current node.
// it returns the unencrypted content and the next node in the path.
// if nextNode != "", the message is padded to size = MessageSize.
func DecryptLayer(encryptedMsg []byte, privateKeyPath string) ([]byte, string, bool, error) {
	// first we load private RSA key
	rsaKey, err := keygen.LoadPrivateKey(privateKeyPath)
	if err != nil {
		return nil, "", false, errors.New("DecryptLayer: error loading private key: " + err.Error())
	}

	// then we decrypt the packet using the private RSA key
	rsaPlaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, rsaKey, encryptedMsg[:RSAKeySizeBytes], nil)
	if err != nil {
		return nil, "", false, errors.New("DecryptLayer: error decrypting using private RSA key: " + err.Error())
	}

	// then we extract the data from the plaintext
	aesKey := rsaPlaintext[:AESKeySize]
	aesLen := binary.BigEndian.Uint16(rsaPlaintext[AESKeySize : AESKeySize+LengthOfCiphertextSize])
	dummyFlag := rsaPlaintext[AESKeySize+LengthOfCiphertextSize]
	nextNode := string(rsaPlaintext[AESKeySize+LengthOfCiphertextSize+DummyFlagLSize:])

	// use AES key to decrypt
	aesCiphertext := encryptedMsg[RSAKeySizeBytes : RSAKeySizeBytes+int(aesLen)]
	content, err := decryptAES(aesCiphertext, aesKey)
	if err != nil {
		return nil, "", false, errors.New("DecryptLayer: error decrypting using ephemeral AES key: " + err.Error())
	}

	// final hop: drop dummy silently, display real
	if nextNode == "" {
		if dummyFlag == 1 {
			return nil, "", true, nil
		}
		return content, "", false, nil
	}

	// intermediate hop: re-pad and forward (real and dummy both go through)
	paddedPacket, err := padMessage(content)
	if err != nil {
		return nil, "", false, errors.New("DecryptLayer: " + err.Error())
	}
	return paddedPacket, nextNode, false, nil
}
