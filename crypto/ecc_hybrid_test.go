package crypto

import (
	"github.com/YazanAbdelal/mixnet/keygen"

	"testing"
)

func TestECCEncrypt(t *testing.T) {
	// gen keys
	privateKey1, publicKey1, err := keygen.GenerateECCKeys()
	if err != nil {
		t.Error("TestECCEncrypt: " + err.Error())
	}
	privateKey2, publicKey2, err := keygen.GenerateECCKeys()
	if err != nil {
		t.Error("TestECCEncrypt: " + err.Error())
	}

	// message to be encrypted and then decrypted
	msg := []byte("Hello, World!")

	// encrypt
	encrypedMsg, err := EncryptWithECC(msg, privateKey1, publicKey2)
	if err != nil {
		t.Error("TestECCEncrypt: " + err.Error())
	}

	// decrypt
	decryptedMsg, err := DecryptWithECC(encrypedMsg, privateKey2, publicKey1)
	if err != nil {
		t.Error("TestECCEncrypt: " + err.Error())
	}

	// test correctness
	if string(msg) != string(decryptedMsg) {
		t.Errorf("TestECCEncrypt: got %q, supposed to be %q.", string(decryptedMsg), string(msg))
	}

}
