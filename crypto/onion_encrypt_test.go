package crypto

import (
	"testing"
)

func TestOnionLayerEncryption(t *testing.T) {
	msg := "Hello, world!"
	nextNode := ""
	encryptedMsg, err := encryptOnionLayer([]byte(msg), nextNode, "./keys/public/client-1-public.pem", false)
	if err != nil {
		t.Error("Error encrypting message: " + err.Error())
		return
	}

	decryptedMsg, decryptedNextNode, _, err := DecryptLayer(encryptedMsg, "./keys/client-1-private.pem")
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

func TestOnionEncrypt(t *testing.T) {
	msg := "Hello, world!"
	encryptedMsg, nextNode, err := OnionEncrypt(msg, "client-2", false)
	if err != nil {
		t.Error("Error encrypting message: " + err.Error())
		return
	}

	decryptedMsg, nextNode, _, err := DecryptLayer(encryptedMsg, "./keys/server-1-private.pem")
	if err != nil {
		t.Error("Error decrypting 1st layer: " + err.Error())
		return
	}
	decryptedMsg, nextNode, _, err = DecryptLayer(decryptedMsg, "./keys/"+nextNode+"-private.pem")
	if err != nil {
		t.Error("Error decrypting 2nd layer: " + err.Error())
		return
	}
	decryptedMsg, nextNode, _, err = DecryptLayer(decryptedMsg, "./keys/"+nextNode+"-private.pem")
	if err != nil {
		t.Error("Error decrypting 3rd layer: " + err.Error())
		return
	}
	decryptedMsg, nextNode, _, err = DecryptLayer(decryptedMsg, "./keys/"+nextNode+"-private.pem")
	if err != nil {
		t.Error("Error decrypting 4th layer: " + err.Error())
		return
	}

	decryptedMsgStr := string(decryptedMsg)

	if decryptedMsgStr != msg {
		t.Errorf("Decrypted message should be %q, got %q instead.", msg, decryptedMsgStr)
		return
	}

}

func TestPacketSize(t *testing.T) {
	msg := "Hello, world!"
	encryptedMsg, nextNode, err := OnionEncrypt(msg, "client-2", false)
	if err != nil {
		t.Error("Error encrypting message: " + err.Error())
		return
	}

	decryptedMsg1, nextNode, _, err := DecryptLayer(encryptedMsg, "./keys/server-1-private.pem")
	if err != nil {
		t.Error("Error decrypting 1st layer: " + err.Error())
		return
	}
	decryptedMsg2, nextNode, _, err := DecryptLayer(decryptedMsg1, "./keys/"+nextNode+"-private.pem")
	if err != nil {
		t.Error("Error decrypting 2nd layer: " + err.Error())
		return
	}
	decryptedMsg3, nextNode, _, err := DecryptLayer(decryptedMsg2, "./keys/"+nextNode+"-private.pem")
	if err != nil {
		t.Error("Error decrypting 3rd layer: " + err.Error())
		return
	}
	// no need to check last layer becasue the function does not pad when it reaches the target.

	if len(encryptedMsg) != len(decryptedMsg1) ||
		len(decryptedMsg1) != len(decryptedMsg2) ||
		len(decryptedMsg2) != len(decryptedMsg3) {
		t.Errorf("Packets are not of equal size: !(%v = %v = %v = %v)", len(encryptedMsg), len(decryptedMsg1), len(decryptedMsg2), len(decryptedMsg3))
		return
	}
}
