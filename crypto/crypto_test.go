package crypto

import "testing"

func TestLoad(t *testing.T) {
	// create keys:
	privateKey, publicKey, err := GenerateKeys(4096)

	// save keys
	err = ExportPrivateKey(privateKey, "tmp", "private.pem")
	if err != nil {
		t.Error("Error exporting private key: " + err.Error())
		return
	}
	err = ExportPublicKey(publicKey, "tmp", "public.pem")
	if err != nil {
		t.Error("Error exporting public key: " + err.Error())
		return
	}

	// load keys
	pri, err := LoadPrivateKey("tmp/private.pem")
	if err != nil {
		t.Error("Error loading private key: " + err.Error())
	}
	pub, err := LoadPublicKey("tmp/public.pem")
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
	privateKey, publicKey, err := GenerateKeys(4096)

	msgBytes := []byte("Hello, world!")

	encryptedBytes, err := EncryptMessage(msgBytes, publicKey, 32)
	if err != nil {
		t.Error("Error encrypting message: " + err.Error())
		return
	}

	decryptedMessage, err := DecryptMessage(encryptedBytes, privateKey)
	if err != nil {
		t.Error("Error dectypting message: " + err.Error())
		return
	}

	if string(decryptedMessage) != string(msgBytes) {
		t.Errorf("original message and decrypted message are not the same:\n Original: %s\n Decrypted: %s\n", string(msgBytes), string(decryptedMessage))
	}
}
