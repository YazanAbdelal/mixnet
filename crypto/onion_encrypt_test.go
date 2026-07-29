package crypto

import (
	"path/filepath"
	"testing"

	"github.com/YazanAbdelal/mixnet/keygen"
)

func generateKeyPair(t testing.TB, dir, name string) {
	t.Helper()

	privKey, pubKey, err := keygen.GenerateRSAKeys(4096)
	if err != nil {
		t.Fatalf("GenerateKeys: %v", err)
	}

	if err := keygen.ExportPrivateRSAKey(privKey, dir, name+"-private.pem"); err != nil {
		t.Fatalf("ExportPrivateKey: %v", err)
	}
	if err := keygen.ExportPublicRSAKey(pubKey, filepath.Join(dir, "public"), name+"-public.pem"); err != nil {
		t.Fatalf("ExportPublicKey: %v", err)
	}
}

func TestOnionLayerEncryption(t *testing.T) {
	baseDir := t.TempDir()
	generateKeyPair(t, baseDir, "client-1")

	msg := "Hello, world!"
	nextNode := ""
	encryptedMsg, err := encryptOnionLayer([]byte(msg), nextNode, filepath.Join(baseDir, "public", "client-1-public.pem"), false)
	if err != nil {
		t.Error("Error encrypting message: " + err.Error())
		return
	}

	decryptedMsg, decryptedNextNode, _, err := DecryptLayer(encryptedMsg, filepath.Join(baseDir, "client-1-private.pem"))
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
	baseDir := t.TempDir()
	servers := []string{"server-1", "server-2", "server-3"}
	allNodes := append(servers, "client-1")
	for _, name := range allNodes {
		generateKeyPair(t, baseDir, name)
	}

	msg := "Hello, world!"
	encryptedMsg, firstNode, err := OnionEncrypt(msg, "client-1", false, servers, 3, filepath.Join(baseDir, "public")+"/")
	if err != nil {
		t.Error("Error encrypting message: " + err.Error())
		return
	}

	nextNode := firstNode
	for i := 0; i < 4; i++ {
		decryptedMsg, nxt, _, err := DecryptLayer(encryptedMsg, filepath.Join(baseDir, nextNode+"-private.pem"))
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

func TestPacketSize(t *testing.T) {
	baseDir := t.TempDir()
	servers := []string{"server-1", "server-2", "server-3"}
	allNodes := append(servers, "client-2")
	for _, name := range allNodes {
		generateKeyPair(t, baseDir, name)
	}

	msg := "Hello, world!"
	encryptedMsg, firstNode, err := OnionEncrypt(msg, "client-2", false, servers, 3, filepath.Join(baseDir, "public")+"/")
	if err != nil {
		t.Error("Error encrypting message: " + err.Error())
		return
	}

	var sizes []int
	sizes = append(sizes, len(encryptedMsg))

	nextNode := firstNode
	for i := 0; i < 3; i++ {
		decryptedMsg, nxt, _, err := DecryptLayer(encryptedMsg, filepath.Join(baseDir, nextNode+"-private.pem"))
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

func BenchmarkOnionEncrypt(b *testing.B) {
	baseDir := b.TempDir()
	servers := []string{"server-1", "server-2", "server-3"}
	allNodes := append(servers, "client-1")
	for _, name := range allNodes {
		generateKeyPair(b, baseDir, name)
	}

	msg := "Hello, world! This is a benchmark test message to measure performance."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := OnionEncrypt(msg, "client-1", false, servers, 3, filepath.Join(baseDir, "public")+"/")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFullRoundTrip(b *testing.B) {
	baseDir := b.TempDir()
	servers := []string{"server-1", "server-2", "server-3"}
	allNodes := append(servers, "client-1")
	for _, name := range allNodes {
		generateKeyPair(b, baseDir, name)
	}

	msg := "Hello, world! This is a benchmark test message to measure performance."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encryptedMsg, firstNode, err := OnionEncrypt(msg, "client-1", false, servers, 3, filepath.Join(baseDir, "public")+"/")
		if err != nil {
			b.Fatal(err)
		}

		nextNode := firstNode
		for j := 0; j < 4; j++ {
			decryptedMsg, nxt, _, err := DecryptLayer(encryptedMsg, filepath.Join(baseDir, nextNode+"-private.pem"))
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
