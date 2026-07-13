package crypto

import (
	"testing"

	"github.com/YazanAbdelal/mixnet/keygen"
)

func TestLoad(t *testing.T) {
	privateKey, publicKey, err := keygen.GenerateKeys(4096)
	if err != nil {
		t.Fatalf("GenerateKeys: %v", err)
	}

	dir := t.TempDir()

	err = keygen.ExportPrivateKey(privateKey, dir, "private.pem")
	if err != nil {
		t.Fatalf("ExportPrivateKey: %v", err)
	}
	err = keygen.ExportPublicKey(publicKey, dir, "public.pem")
	if err != nil {
		t.Fatalf("ExportPublicKey: %v", err)
	}

	pri, err := keygen.LoadPrivateKey(dir + "/private.pem")
	if err != nil {
		t.Fatalf("LoadPrivateKey: %v", err)
	}
	pub, err := keygen.LoadPublicKey(dir + "/public.pem")
	if err != nil {
		t.Fatalf("LoadPublicKey: %v", err)
	}

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
