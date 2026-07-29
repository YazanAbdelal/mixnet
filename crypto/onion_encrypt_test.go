package crypto

import (
	"path/filepath"
	"testing"

	"github.com/YazanAbdelal/mixnet/keygen"
)

func generateRSAKeyPair(t testing.TB, dir, name string) {
	t.Helper()

	privKey, pubKey, err := keygen.GenerateRSAKeys(4096)
	if err != nil {
		t.Fatalf("generateRSAKeyPair: %v", err)
	}

	if err := keygen.ExportPrivateRSAKey(privKey, dir, name+"-rsa-private.pem"); err != nil {
		t.Fatalf("generateRSAKeyPair: %v", err)
	}
	if err := keygen.ExportPublicRSAKey(pubKey, filepath.Join(dir, "public"), name+"-rsa-public.pem"); err != nil {
		t.Fatalf("generateRSAKeyPair: %v", err)
	}
}

func generateECCKeyPair(t testing.TB, dir, name string) {
	t.Helper()

	privateKey, publicKey, err := keygen.GenerateECCKeys()
	if err != nil {
		t.Fatal("generateECCKeyPair: " + err.Error())
	}

	if err := keygen.ExportPrivateECCKey(privateKey, dir, name+"-ecc-private.pem"); err != nil {
		t.Fatal("generateECCKeyPair: " + err.Error())
	}

	if err := keygen.ExportPublicECCKey(publicKey, filepath.Join(dir, "public"), name+"-ecc-public.pem"); err != nil {
		t.Fatal("generateECCKeyPair: " + err.Error())
	}
}

func TestOnionLayerRSAEncryption(t *testing.T) {
	baseDir := t.TempDir()
	generateRSAKeyPair(t, baseDir, "client-1")

	msg := "Hello, world!"
	nextNode := ""
	encryptedMsg, err := encryptOnionLayerWithRSA([]byte(msg), nextNode, filepath.Join(baseDir, "public", "client-1-rsa-public.pem"), false)
	if err != nil {
		t.Fatal("TestOnionLayerRSAEncryption: Error encrypting message: " + err.Error())
	}

	decryptedMsg, decryptedNextNode, _, err := DecryptOnionLayerWithRSA(encryptedMsg, filepath.Join(baseDir, "client-1-rsa-private.pem"))
	if err != nil {
		t.Fatal("TestOnionLayerRSAEncryption: Error decrypting message: " + err.Error())
	}

	if string(decryptedMsg) != msg {
		t.Fatalf("TestOnionLayerRSAEncryption: Decrypted message should be %q, got %q instead.", msg, string(decryptedMsg))
	}

	if nextNode != decryptedNextNode {
		t.Fatalf("TestOnionLayerRSAEncryption: Next node should be %q, got %q instead.", nextNode, decryptedNextNode)
	}
}

func TestOnionLayerECCEncryption(t *testing.T) {
	baseDir := t.TempDir()
	generateECCKeyPair(t, baseDir, "client-1")
	generateECCKeyPair(t, baseDir, "client-2")

	msg := "Hello, world!"
	nextNode := ""
	encryptedMsg, err := EncryptOnionLayerWithECC([]byte(msg), nextNode, filepath.Join(baseDir, "public", "client-2-ecc-public.pem"), false)
	if err != nil {
		t.Fatal("TestOnionLayerECCEncryption: Error encrypting message: " + err.Error())
	}

	decryptedMsg, decryptedNextNode, _, err := DecryptOnionLayerWithECC(encryptedMsg, filepath.Join(baseDir, "client-2-ecc-private.pem"))
	if err != nil {
		t.Fatal("TestOnionLayerECCEncryption: Error decrypting message: " + err.Error())
	}

	if string(decryptedMsg) != msg {
		t.Fatalf("TestOnionLayerECCEncryption: Decrypted message should be %q, got %q instead.", msg, string(decryptedMsg))
	}

	if nextNode != decryptedNextNode {
		t.Fatalf("TestOnionLayerECCEncryption: Next node should be %q, got %q instead.", nextNode, decryptedNextNode)
	}
}

func TestOnionEncryptRSA(t *testing.T) {
	baseDir := t.TempDir()
	servers := []string{"server-1", "server-2", "server-3"}
	allNodes := append(servers, "client-1")
	for _, name := range allNodes {
		generateRSAKeyPair(t, baseDir, name)
	}

	msg := "Hello, world!"
	encryptedMsg, firstNode, err := OnionEncrypt(msg, "client-1", false, servers, 3, filepath.Join(baseDir, "public")+"/", "rsa")
	if err != nil {
		t.Error("Error encrypting message: " + err.Error())
		return
	}

	nextNode := firstNode
	for i := 0; i < 4; i++ {
		decryptedMsg, nxt, _, err := DecryptOnionLayerWithRSA(encryptedMsg, filepath.Join(baseDir, nextNode+"-rsa-private.pem"))
		if err != nil {
			t.Errorf("Error decrypting layer %d: %v", i+1, err)
			return
		}
		encryptedMsg = decryptedMsg
		nextNode = nxt
	}

	if string(encryptedMsg) != msg {
		t.Errorf("Decrypted message should be %q, got %q instead.", msg, string(encryptedMsg))
	}
}

func TestOnionEncryptECC(t *testing.T) {
	baseDir := t.TempDir()
	servers := []string{"server-1", "server-2", "server-3"}
	allNodes := append(servers, "client-1")
	for _, name := range allNodes {
		generateECCKeyPair(t, baseDir, name)
	}

	msg := "Hello, world!"
	encryptedMsg, firstNode, err := OnionEncrypt(msg, "client-1", false, servers, 3, filepath.Join(baseDir, "public")+"/", "ecc")
	if err != nil {
		t.Fatal("TestOnionEncryptECC: Error encrypting message: " + err.Error())
	}

	nextNode := firstNode
	for i := 0; i < 4; i++ {
		decryptedMsg, nxt, _, err := DecryptOnionLayerWithECC(encryptedMsg, filepath.Join(baseDir, nextNode+"-ecc-private.pem"))
		if err != nil {
			t.Fatalf("TestOnionEncryptECC: Error decrypting layer %d: %v", i+1, err)
		}
		encryptedMsg = decryptedMsg
		nextNode = nxt
	}

	if string(encryptedMsg) != msg {
		t.Fatalf("TestOnionEncryptECC: Decrypted message should be %q, got %q instead.", msg, string(encryptedMsg))
	}
}

func TestPacketSizeRSA(t *testing.T) {
	baseDir := t.TempDir()
	servers := []string{"server-1", "server-2", "server-3"}
	allNodes := append(servers, "client-2")
	for _, name := range allNodes {
		generateRSAKeyPair(t, baseDir, name)
	}

	msg := "Hello, world!"
	encryptedMsg, firstNode, err := OnionEncrypt(msg, "client-2", false, servers, 3, filepath.Join(baseDir, "public")+"/", "rsa")
	if err != nil {
		t.Error("Error encrypting message: " + err.Error())
		return
	}

	var sizes []int
	sizes = append(sizes, len(encryptedMsg))

	nextNode := firstNode
	for i := 0; i < 3; i++ {
		decryptedMsg, nxt, _, err := DecryptOnionLayerWithRSA(encryptedMsg, filepath.Join(baseDir, nextNode+"-rsa-private.pem"))
		if err != nil {
			t.Errorf("Error decrypting layer %d: %v", i+1, err)
			return
		}
		sizes = append(sizes, len(decryptedMsg))
		encryptedMsg = decryptedMsg
		nextNode = nxt
	}

	for i := 1; i < len(sizes); i++ {
		if sizes[i] != sizes[0] {
			t.Errorf("Packets are not of equal size: %v", sizes)
			return
		}
	}
}

func TestPacketSizeECC(t *testing.T) {
	baseDir := t.TempDir()
	servers := []string{"server-1", "server-2", "server-3"}
	allNodes := append(servers, "client-2")
	for _, name := range allNodes {
		generateECCKeyPair(t, baseDir, name)
	}

	msg := "Hello, world!"
	encryptedMsg, firstNode, err := OnionEncrypt(msg, "client-2", false, servers, 3, filepath.Join(baseDir, "public")+"/", "ecc")
	if err != nil {
		t.Error("Error encrypting message: " + err.Error())
		return
	}

	var sizes []int
	sizes = append(sizes, len(encryptedMsg))

	nextNode := firstNode
	for i := 0; i < 3; i++ {
		decryptedMsg, nxt, _, err := DecryptOnionLayerWithECC(encryptedMsg, filepath.Join(baseDir, nextNode+"-ecc-private.pem"))
		if err != nil {
			t.Errorf("Error decrypting layer %d: %v", i+1, err)
			return
		}
		sizes = append(sizes, len(decryptedMsg))
		encryptedMsg = decryptedMsg
		nextNode = nxt
	}

	for i := 1; i < len(sizes); i++ {
		if sizes[i] != sizes[0] {
			t.Errorf("Packets are not of equal size: %v", sizes)
			return
		}
	}
}

func BenchmarkOnionEncryptRSA(b *testing.B) {
	baseDir := b.TempDir()
	servers := []string{"server-1", "server-2", "server-3"}
	allNodes := append(servers, "client-1")
	for _, name := range allNodes {
		generateRSAKeyPair(b, baseDir, name)
	}

	msg := "Hello, world! This is a benchmark test message to measure performance."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := OnionEncrypt(msg, "client-1", false, servers, 3, filepath.Join(baseDir, "public")+"/", "rsa")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOnionEncryptECC(b *testing.B) {
	baseDir := b.TempDir()
	servers := []string{"server-1", "server-2", "server-3"}
	allNodes := append(servers, "client-1")
	for _, name := range allNodes {
		generateECCKeyPair(b, baseDir, name)
	}

	msg := "Hello, world! This is a benchmark test message to measure performance."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := OnionEncrypt(msg, "client-1", false, servers, 3, filepath.Join(baseDir, "public")+"/", "ecc")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFullRoundTripRSA(b *testing.B) {
	baseDir := b.TempDir()
	servers := []string{"server-1", "server-2", "server-3"}
	allNodes := append(servers, "client-1")
	for _, name := range allNodes {
		generateRSAKeyPair(b, baseDir, name)
	}

	msg := "Hello, world! This is a benchmark test message to measure performance."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encryptedMsg, firstNode, err := OnionEncrypt(msg, "client-1", false, servers, 3, filepath.Join(baseDir, "public")+"/", "rsa")
		if err != nil {
			b.Fatal(err)
		}

		nextNode := firstNode
		for j := 0; j < 4; j++ {
			decryptedMsg, nxt, _, err := DecryptOnionLayerWithRSA(encryptedMsg, filepath.Join(baseDir, nextNode+"-rsa-private.pem"))
			if err != nil {
				b.Fatalf("Error decrypting layer %d: %v", j+1, err)
			}
			encryptedMsg = decryptedMsg
			nextNode = nxt
		}

		if string(encryptedMsg) != msg {
			b.Fatalf("Decrypted message should be %q, got %q instead.", msg, string(encryptedMsg))
		}
	}
}

func BenchmarkFullRoundTripECC(b *testing.B) {
	baseDir := b.TempDir()
	servers := []string{"server-1", "server-2", "server-3"}
	allNodes := append(servers, "client-1")
	for _, name := range allNodes {
		generateECCKeyPair(b, baseDir, name)
	}

	msg := "Hello, world! This is a benchmark test message to measure performance."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encryptedMsg, firstNode, err := OnionEncrypt(msg, "client-1", false, servers, 3, filepath.Join(baseDir, "public")+"/", "ecc")
		if err != nil {
			b.Fatal(err)
		}

		nextNode := firstNode
		for j := 0; j < 4; j++ {
			decryptedMsg, nxt, _, err := DecryptOnionLayerWithECC(encryptedMsg, filepath.Join(baseDir, nextNode+"-ecc-private.pem"))
			if err != nil {
				b.Fatalf("Error decrypting layer %d: %v", j+1, err)
			}
			encryptedMsg = decryptedMsg
			nextNode = nxt
		}

		if string(encryptedMsg) != msg {
			b.Fatalf("Decrypted message should be %q, got %q instead.", msg, string(encryptedMsg))
		}
	}
}
