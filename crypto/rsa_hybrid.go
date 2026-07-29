package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"errors"

	"github.com/YazanAbdelal/mixnet/keygen"
)

// packRSA concatenates the AES key, dummy flag byte, next node and the ciphertext length into one slice.
// This function is called before encryptOnionLayerWithRSA to embed the metadata in the ciphertext for security.
func packRSA(aesKey []byte, isDummy bool, nextNode string, cipherTextLength uint16) []byte {

	// make an array of bytes to store the key, ciphertext, and metadata
	// length = 0, capacity = AESKeySize + LengthOfCiphertextSize + len(nextNode)
	rsaPlaintext := make([]byte, 0, AESKeySize+LengthOfCiphertextSize+DummyFlagLSize+len(nextNode))

	// add the AES key first
	rsaPlaintext = append(rsaPlaintext, aesKey...)

	// then add the length of the ciphertext
	rsaPlaintext = binary.BigEndian.AppendUint16(rsaPlaintext, cipherTextLength)

	// then add the dummy flag byte
	if isDummy {
		rsaPlaintext = append(rsaPlaintext, 1)
	} else {
		rsaPlaintext = append(rsaPlaintext, 0)
	}

	// then add the next node ID and return
	return append(rsaPlaintext, []byte(nextNode)...)
}

// unpackRSA unpacks the plaintext to AES key, AES ciphertext length, dummy flag and next node.
func unpackRSA(rsaPlaintext []byte) ([]byte, uint16, byte, string) {
	aesKey := rsaPlaintext[:AESKeySize]
	aesLen := binary.BigEndian.Uint16(rsaPlaintext[AESKeySize : AESKeySize+LengthOfCiphertextSize])
	dummyFlag := rsaPlaintext[AESKeySize+LengthOfCiphertextSize]
	nextNode := string(rsaPlaintext[AESKeySize+LengthOfCiphertextSize+DummyFlagLSize:])

	return aesKey, aesLen, dummyFlag, nextNode
}

// encryptOnionLayerWithRSA adds a layer of encryption to the onion using hybrid encryption.
// It returns a ciphertext that includes an ephemeral AES key, the length of the encrypted content, the next node, and the AES encrypted content.
func encryptOnionLayerWithRSA(content []byte, nextNode string, pathToKey string, isDummy bool) ([]byte, error) {
	// first we generate an ephemeral AES key
	aesKey, err := genAESKey(AESKeySize)
	if err != nil {
		return nil, errors.New("encryptOnionLayerWithRSA: error generating AES key: " + err.Error())
	}

	// then we encrypt the content using the AES key
	ciphertext, err := encryptWithAES(content, aesKey)
	if err != nil {
		return nil, errors.New("encryptOnionLayerWithRSA: error encrypting using AES key: " + err.Error())
	}

	// then we use the node's public RSA key to encrypt the length of the AES key, the length of the ciphertext, the dummy flag and the next destination
	// first we concatenate them
	rsaPlaintext := packRSA(aesKey, isDummy, nextNode, uint16(len(ciphertext)))

	// now we load the public RSA key of the target node
	rsaKey, err := keygen.LoadPublicRSAKey(pathToKey)
	if err != nil {
		return nil, errors.New("encryptOnionLayerWithRSA: error loading public key: " + err.Error())
	}

	// now we encrypt using the target node's public RSA key
	// the resulting ciphertext here is 512 bytes long because we used a 4096 bit long RSA key
	rsaCiphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaKey, rsaPlaintext, nil)
	if err != nil {
		return nil, errors.New("encryptOnionLayerWithRSA: error encrypting with RSA key: " + err.Error())
	}

	return append(rsaCiphertext, ciphertext...), nil
}

// DecryptOnionLayerWithRSA decrypts an onion layer using the private RSA key of the current node.
// it returns the unencrypted content and the next node in the path.
// if nextNode != "", the message is padded to size = MessageSize.
func DecryptOnionLayerWithRSA(encryptedMsg []byte, privateKeyPath string) ([]byte, string, bool, error) {
	// first we load private RSA key
	rsaKey, err := keygen.LoadPrivateRSAKey(privateKeyPath)
	if err != nil {
		return nil, "", false, errors.New("DecryptOnionLayerWithRSA: error loading private key: " + err.Error())
	}

	// then we decrypt the packet using the private RSA key
	rsaPlaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, rsaKey, encryptedMsg[:RSAKeySizeBytes], nil)
	if err != nil {
		return nil, "", false, errors.New("DecryptOnionLayerWithRSA: error decrypting using private RSA key: " + err.Error())
	}

	// then we extract the data from the plaintext
	aesKey, aesLen, dummyFlag, nextNode := unpackRSA(rsaPlaintext)

	// use AES key to decrypt
	aesCiphertext := encryptedMsg[RSAKeySizeBytes : RSAKeySizeBytes+aesLen]
	content, err := decryptAES(aesCiphertext, aesKey)
	if err != nil {
		return nil, "", false, errors.New("DecryptOnionLayerWithRSA: error decrypting using ephemeral AES key: " + err.Error())
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
		return nil, "", false, errors.New("DecryptOnionLayerWithRSA: " + err.Error())
	}
	return paddedPacket, nextNode, false, nil
}
