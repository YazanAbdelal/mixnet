package crypto

import (
	"crypto/ecdh"
	"encoding/binary"
	"errors"

	"github.com/YazanAbdelal/mixnet/keygen"
)

// packECCMetadata packs all metadata in a 64-Byte block.
// Metadata for each layer is comprised of: The dummy flag, the length of content, the length of the next node's ID, the next node's ID.
// The rest of the 64 Bytes is padding.
func packECCMetadata(nextNode string, isDummy bool, contentLen uint16) []byte {
	// dummy flag (1 Byte) + length of content (2 Bytes) + length of next node (2 Bytes) + next node

	// make an array of bytes to store the plaintext
	eccPlaintext := make([]byte, 0, ECCMetadataBlockSize)

	// add the dummy flag byte
	if isDummy {
		eccPlaintext = append(eccPlaintext, 1)
	} else {
		eccPlaintext = append(eccPlaintext, 0)
	}

	// add the length of the ciphertext
	lengthOfContent := uint16(contentLen)
	eccPlaintext = binary.BigEndian.AppendUint16(eccPlaintext, lengthOfContent)

	// add the length of the next node
	lengthOfNextNode := uint16(len(nextNode))
	eccPlaintext = binary.BigEndian.AppendUint16(eccPlaintext, lengthOfNextNode)

	// add the next node
	eccPlaintext = append(eccPlaintext, []byte(nextNode)...)

	// pad to ECCMetdataBlockSize
	if len(eccPlaintext) < ECCMetadataBlockSize {
		padding := make([]byte, ECCMetadataBlockSize-len(eccPlaintext))
		eccPlaintext = append(eccPlaintext, padding...)
	}

	return eccPlaintext
}

// unpackECCMetadata unpacks a metadata block and returns the next node, the content size, and the dummy flag.
func unpackECCMetadata(plaintext []byte) (string, uint16, byte) {
	// extract the dummy flag
	dummyFlag := plaintext[0]

	// extract the length of the ciphertext
	ciphertextSize := binary.BigEndian.Uint16(plaintext[DummyFlagLSize : DummyFlagLSize+LengthOfCiphertextSize])

	// extract the size of the next node ID
	nextNodeSize := binary.BigEndian.Uint16(plaintext[DummyFlagLSize+LengthOfCiphertextSize : DummyFlagLSize+LengthOfCiphertextSize+LengthOfNextNodeSize])

	// extract the next node
	nextNode := plaintext[DummyFlagLSize+LengthOfCiphertextSize+LengthOfNextNodeSize : DummyFlagLSize+LengthOfCiphertextSize+LengthOfNextNodeSize+nextNodeSize]

	return string(nextNode), ciphertextSize, dummyFlag
}

// EncryptOnionLayerWithECC encrypts a slice of bytes using ECC.
func EncryptOnionLayerWithECC(unencryptedBytes []byte, nextNode string, pathToKey string, isDummy bool) ([]byte, error) {
	// load public key
	publicKey, err := keygen.LoadPublicECCKey(pathToKey)
	if err != nil {
		return nil, errors.New("EncryptOnionLayerWithECC: " + err.Error())
	}

	// generate ephemeral keys
	ephPrivateKey, ephPublicKey, err := keygen.GenerateECCKeys()
	if err != nil {
		return nil, errors.New("EncryptOnionLayerWithECC: " + err.Error())
	}

	// get shared AES key
	aesKey, err := DeriveAESKey(ephPrivateKey, publicKey)
	if err != nil {
		return nil, errors.New("EncryptOnionLayerWithECC: " + err.Error())
	}

	// encrypt content
	encryptedBytes, err := encryptWithAES(unencryptedBytes, aesKey)
	if err != nil {
		return nil, errors.New("EncryptOnionLayerWithECC: " + err.Error())
	}

	// pack metadata with plaintext
	unencryptedMetadata := packECCMetadata(nextNode, isDummy, uint16(len(encryptedBytes)))

	// encrypt metadata
	encryptedMetadata, err := encryptWithAES(unencryptedMetadata, aesKey)
	if err != nil {
		return nil, errors.New("EncryptOnionLayerWithECC: " + err.Error())
	}

	// concatenate ephemeral public key + AES encrypted metadata block + AES encrypted content
	packet := append(ephPublicKey.Bytes(), encryptedMetadata...)
	packet = append(packet, encryptedBytes...)

	return packet, nil
}

// DecryptOnionLayerWithECC decrypts a slice of bytes that was encrypted using ECC.
func DecryptOnionLayerWithECC(encryptedPacket []byte, privateKeyPath string) ([]byte, string, bool, error) {
	// load private key
	privateKey, err := keygen.LoadPrivateECCKey(privateKeyPath)
	if err != nil {
		return nil, "", false, errors.New("DecryptOnionLayerWithECC: " + err.Error())
	}

	// extract ephemeral public key
	ephPublicKey, err := ecdh.X25519().NewPublicKey(encryptedPacket[:ECCKeySize])
	if err != nil {
		return nil, "", false, errors.New("DecryptOnionLayerWithECC: Error extracting ephemeral public key: " + err.Error())
	}

	// derive AES key
	aesKey, err := DeriveAESKey(privateKey, ephPublicKey)
	if err != nil {
		return nil, "", false, errors.New("DecryptOnionLayerWithECC: " + err.Error())
	}

	// extract metadata
	encryptedMetadataBlock := encryptedPacket[ECCKeySize : ECCKeySize+ECCMetadataBlockSizeAfterGCM]

	// decrypt metadata
	metadataBlock, err := decryptAES(encryptedMetadataBlock, aesKey)
	if err != nil {
		return nil, "", false, errors.New("DecryptOnionLayerWithECC: Error decrypting metadata" + err.Error())
	}

	// unpack metadata
	nextNode, contentSize, dummyFlag := unpackECCMetadata(metadataBlock)

	// extract content
	encryptedContent := encryptedPacket[ECCKeySize+ECCMetadataBlockSizeAfterGCM : ECCKeySize+ECCMetadataBlockSizeAfterGCM+contentSize]

	// decrypt content using the AES key
	decryptedContent, err := decryptAES(encryptedContent, aesKey)
	if err != nil {
		return nil, "", false, errors.New("DecryptOnionLayerWithECC: " + err.Error())
	}

	// final hop: drop dummy silently, display real
	if nextNode == "" {
		if dummyFlag == 1 {
			return nil, "", true, nil
		}
		return decryptedContent, "", false, nil
	}

	// intermediate hop: re-pad and forward (real and dummy both go through)
	paddedPacket, err := padMessage(decryptedContent)
	if err != nil {
		return nil, "", false, errors.New("DecryptOnionLayerWithRSA: " + err.Error())
	}

	return paddedPacket, nextNode, false, nil
}
