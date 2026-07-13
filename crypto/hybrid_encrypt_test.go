package crypto

import (
	"testing"

	"github.com/YazanAbdelal/mixnet/keygen"
)

func TestLoad(t *testing.T) {
	// create keys:
	privateKey, publicKey, err := keygen.GenerateKeys(4096)

	// save keys
	err = keygen.ExportPrivateKey(privateKey, "tmp", "private.pem")
	if err != nil {
		t.Error("Error exporting private key: " + err.Error())
		return
	}
	err = keygen.ExportPublicKey(publicKey, "tmp", "public.pem")
	if err != nil {
		t.Error("Error exporting public key: " + err.Error())
		return
	}

	// load keys
	pri, err := keygen.LoadPrivateKey("tmp/private.pem")
	if err != nil {
		t.Error("Error loading private key: " + err.Error())
	}
	pub, err := keygen.LoadPublicKey("tmp/public.pem")
	if err != nil {
		t.Error("Error loading public key: " + err.Error())
	}

	// compare
	if !privateKey.Equal(pri) {
		t.Error("Loaded private key is not equal to the original key.")
	}
	if !publicKey.Equal(pub) {
		t.Error("Loaded public key is not equal to the original key.")
	}
}

func TestEncrypt(t *testing.T) {
	// create keys:
	privateKey, publicKey, err := keygen.GenerateKeys(4096)

	msgBytes := []byte("Hello, world!")

	encryptedBytes, err := EncryptMessage(msgBytes, publicKey, 32)
	if err != nil {
		t.Error("Error encrypting message: " + err.Error())
		return
	}

	decryptedMessage, err := DecryptMessage(encryptedBytes, privateKey)
	if err != nil {
		t.Error("Error decrypting message: " + err.Error())
		return
	}

	if string(decryptedMessage) != string(msgBytes) {
		t.Errorf("original message and decrypted message are not the same:\n Original: %s\n Decrypted: %s\n", string(msgBytes), string(decryptedMessage))
	}
}
