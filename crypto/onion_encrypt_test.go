package crypto

import (
	"testing"

	"github.com/YazanAbdelal/mixnet/keygen"
)

func TestOnionLayerEncryption(t *testing.T) {
	privateKey, publicKey, err := keygen.GenerateKeys(RSAKeySizeBytes * 8)
	if err != nil {
		t.Error("Error generating RSA keys: " + err.Error())
		return
	}

	err = keygen.ExportPrivateKey(privateKey, "keys", "private.pem")
	if err != nil {
		t.Error("Error exporting private key: " + err.Error())
	}
	err = keygen.ExportPublicKey(publicKey, "keys/public", "public.pem")
	if err != nil {
		t.Error("Error exporting public key: " + err.Error())
	}

	msg := "Hello, world!"
	nextNode := "node 1"
	encryptedMsg, err := encryptOnionLayer([]byte(msg), nextNode, "./keys/public/public.pem")
	if err != nil {
		t.Error("Error encrypting message: " + err.Error())
		return
	}

	decryptedMsg, decryptedNextNode, err := DecryptLayer(encryptedMsg)
	if err != nil {
		t.Error("Error decrypting message: " + err.Error())
		return
	}

	if string(decryptedMsg) != msg {
		t.Errorf("decrypted message should be %q, got %q instead.", msg, string(decryptedMsg))
	}

	if nextNode != decryptedNextNode {
		t.Errorf("next node should be %q, got %q instead.", nextNode, decryptedNextNode)
	}
}
