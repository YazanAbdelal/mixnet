package crypto

import (
	"encoding/binary"
	"log"

	"google.golang.org/protobuf/proto"

	"github.com/YazanAbdelal/mixnet/keygen"
	pb "github.com/YazanAbdelal/mixnet/proto/gen"
)

const (
	MessageSize = 4096
	RSAKeySize  = 512
)

// ENCRYPTION:
// each layer is encrypted as follows:
// 1. first we generate an ephemeral
func encryptOnionLayer(content []byte, nextNode string, pathToKey string) ([]byte, error) {
	// // make a new layer struct that includes both the content of the message and the next node
	// newLayer := &pb.OnionLayer{
	// 	NextNode: nextNode,
	// 	Payload:  content,
	// }

	// we will not use protobuf for the layers
	// instead we will use fixed-sized fields for each message
	// 2 bytes for the length of the message
	prefix := make([]byte, 2)
	binary.BigEndian.PutUint16(prefix, uint16(len(content))) // fill the prefix with the length of the content
	packet := append(prefix, content...)

	// convert struct into bytes
	newLayerSerialized := serializeLayer(newLayer)

	// encrypt the entire layer (content + next node) to protect path privacy
	key, err := keygen.LoadPublicKey(pathToKey)
	if err != nil {
		return nil, err
	}
	return EncryptMessage(newLayerSerialized, key, 32) // TODO is 32 good?
}

func OnionEncrypt(msg string, dest string, round int) []byte {
	// convert into bytes before
	msgBytes := []byte(msg)

	// encrypt layer by layer
	// TODO remove '_' and handle errors
	msgBytes, _ = encryptOnionLayer(msgBytes, "", "public_keys/"+dest+"_public_key.pem")
	msgBytes, _ = encryptOnionLayer(msgBytes, dest, "public_keys/server3_public_key.pem")
	msgBytes, _ = encryptOnionLayer(msgBytes, "server-3", "public_keys/server2_public_key.pem")
	msgBytes, _ = encryptOnionLayer(msgBytes, "servee-2", "public_keys/server1_public_key.pem")

	return msgBytes
}

func DecryptLayer(encryptedMsg []byte) []byte {
	// first decrypt
	// TODO handle errors
	privateKey, _ := keygen.LoadPrivateKey("private_key/" + "_private_key.pem")
	serializedLayer, _ := DecryptMessage(encryptedMsg, privateKey)

	// now we make an empty container for the OnionLayer struct
	layer := &pb.OnionLayer{}

	// bytes -> OnionLayer struct
	err := proto.Unmarshal(serializedLayer, layer)
	if err != nil {
		log.Fatalf("Error while umarshalling: %v\n", err)
	}

	return layer.Payload
}
