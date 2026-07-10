package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"errors"

	"github.com/YazanAbdelal/mixnet/keygen"
)

const (
	MessageSize            = 4096
	RSAKeySize             = 512
	AESKeySize             = 32
	LengthOfCiphertextSize = 2
)

// encryptOnionLayer adds a layer of encryption to the onion using hybrid encryption
// it returns a ciphertext that includes an ephemeral AES key, the length of the encrypted content, the next node, and the AES encrypted content.
func encryptOnionLayer(content []byte, nextNode string, pathToKey string) ([]byte, error) {
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

	// then we use the node's public RSA key to encrypt the the length of the AES key, the length of the ciphertext, and the next destination
	// first we concatenate them
	rsaPlaintext := make([]byte, 0, AESKeySize+LengthOfCiphertextSize+len(nextNode)) // length = 0, capacity = AESKeySize + LengthOfCiphertextSize + len(nextNode)
	rsaPlaintext = append(rsaPlaintext, aesKey...)
	rsaPlaintext = binary.BigEndian.AppendUint16(rsaPlaintext, uint16(len(ciphertext)))
	rsaPlaintext = append(rsaPlaintext, []byte(nextNode)...)

	// now we load the public RSA key of the target node
	rsaKey, err := keygen.LoadPublicKey(pathToKey)
	if err != nil {
		return nil, errors.New("encryptOnionLayer: error loading public key: " + err.Error())
	}

	// now we encrypt using the target node's public RSA key
	rsaCiphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaKey, rsaPlaintext, nil)
	if err != nil {
		return nil, errors.New("encryptOnionLayer: error encrypting with RSA key: " + err.Error())
	}
	return rsaCiphertext, nil
}

// TODO write this function
func OnionEncrypt(msg string, dest string, round int) ([]byte, error) {
	// convert into bytes before
	msgBytes := []byte(msg)

	// encrypt layer by layer
	// TODO remove '_' and handle errors
	// TODO get the correct paths to the keys from the docker compose
	msgBytes, _ = encryptOnionLayer(msgBytes, "", "public_keys/"+dest+"_public_key.pem")
	msgBytes, _ = encryptOnionLayer(msgBytes, dest, "public_keys/server3_public_key.pem")
	msgBytes, _ = encryptOnionLayer(msgBytes, "server-3", "public_keys/server2_public_key.pem")
	msgBytes, _ = encryptOnionLayer(msgBytes, "servee-2", "public_keys/server1_public_key.pem")

	// TODO handle padding here

	return msgBytes, nil
}

// DecryptLayer decrypts an onion layer using the private RSA key of the current node.
// it returns the ephemeral AES key, the length of the content that was encrypted with the AES key and the next node in the path.
func DecryptLayer(encryptedMsg []byte) ([]byte, uint16, string, error) {
	// first we load private RSA key
	keyPath := "keys/private.pem"
	rsaKey, err := keygen.LoadPrivateKey(keyPath)
	if err != nil {
		return nil, 0, "", errors.New("DecryptLayer: error loading private key: " + err.Error())
	}

	// then we decrypt the packet using the private RSA key
	rsaPlaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, rsaKey, encryptedMsg[:RSAKeySize], nil)
	if err != nil {
		return nil, 0, "", errors.New("DecryptLayer: error decrypting using private RSA key: " + err.Error())
	}

	// then we extract the data from the plaintext
	aesKey := rsaPlaintext[:AESKeySize]
	aesLen := binary.BigEndian.Uint16(rsaPlaintext[AESKeySize : AESKeySize+LengthOfCiphertextSize])
	route := string(rsaPlaintext[AESKeySize+LengthOfCiphertextSize:])

	return aesKey, aesLen, route, nil
}
