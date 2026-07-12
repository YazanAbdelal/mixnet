## mixnet

A source-routed mix network (mixnet) written in Go, deployed via Docker Compose. Messages are wrapped in multiple layers of hybrid encryption (RSA-4096 + AES-256-GCM), routed through a random subset of mix servers, padded to a uniform size, and shuffled in batches — making sender-recipient linkability infeasible for a local adversary.

---

### Setup

The script `./scripts/script.sh` generates everything needed to run the network:

1. **RSA key pairs** (4096-bit) — one pair per node, used for onion encryption. Private keys stay with each node; public keys are shared into a `keys/public/` directory.
2. **mTLS certificates** — a private Certificate Authority (CA) is created, then each node gets a certificate signed by the CA. These are used for authenticated, encrypted gRPC channels between nodes.
3. **config.json** — lists all servers and clients, plus tunable parameters (path length, batch size, timing).
4. **docker-compose.yml** — one service per node with the correct volumes, flags, and network configuration.

After the script finishes, you must edit `docker-compose.yml` to set each client's `--dest` flag to its target (e.g., `--dest client-2`). Then build and start:

```bash
docker compose up -d --build
```

To send messages, attach to a client container and type:

```bash
docker attach client-1
```

---

### Onion Encryption

Each layer of the onion is encrypted using hybrid encryption:

```
  [ 512-byte RSA ciphertext ][ AES-GCM ciphertext ]
```

**Encryption** (inside-out):

1. Generate an ephemeral 32-byte AES key.
2. Encrypt the inner content (padding + plaintext from the previous layer, or the raw message for the innermost layer) using AES-256-GCM. GCM provides authenticated encryption (integrity + confidentiality) and includes a random nonce.
3. Compute `l = len(ciphertext)`.
4. Pack the following into a plaintext buffer:
   - The ephemeral AES key (32 bytes)
   - `l` as a 2-byte big-endian integer
   - A 1-byte dummy flag (1 = dummy, 0 = real)
   - The next node's name (as a string, variable length)
5. Encrypt this buffer with the next node's **public RSA key** using RSA-OAEP (SHA-256). Since the RSA key is 4096 bits, the output is always 512 bytes.
6. Prepend the RSA ciphertext to the AES ciphertext: `[512 bytes][AES output]`.

**Decryption** (outside-in):

1. Read the first 512 bytes of the packet.
2. Decrypt them using the current node's **private RSA key** to recover: AES key, `l`, dummy flag, next node.
3. Read the next `l` bytes as the AES ciphertext.
4. Decrypt the AES ciphertext using the ephemeral AES key to recover the content.
5. If the next node is empty (""), this is the final destination — print the content (or drop it if the dummy flag is set).
6. Otherwise, pad the content with random bytes to exactly 4096 bytes and forward to the next node.

**Padding**: After all onion layers are applied, the outermost layer is padded to exactly 4096 bytes with random data. This is done before every hop as well. This ensures all packets are identical in size regardless of their plaintext length.

**Dummy messages**: When a client has no real message to send, it sends a dummy packet (`"__DUMMY__"` with the dummy flag set to 1). These are forwarded through the entire path like real packets, but silently dropped at the final destination. This prevents an adversary from distinguishing active senders from idle ones.

---

### Path Selection

The client picks `path_len` random servers from the pool for each message (without replacement), then appends the destination and an empty-string sentinel:

```
[server-3, server-1, server-2, client-2, ""]
```

The encryption goes inside-out:
- Layer 5 (innermost): encrypt for client-2, nextNode = ""
- Layer 4: encrypt for server-2, nextNode = "client-2"
- Layer 3: encrypt for server-1, nextNode = "server-2"
- Layer 2 (outermost): encrypt for server-3, nextNode = "server-1"

The message is sent to `randomPath[0]` (server-3).

---

### Sending Messages

1. The user types a message into stdin.
2. `readStdin()` reads the input and sends it through a channel to `sendLoop()`.
3. `sendLoop()` fires every `client_tick_us` microseconds. If a real message is waiting, it encrypts it with `OnionEncrypt()`; otherwise it encrypts a dummy.
4. The encrypted packet (always 4096 bytes) is sent to the first mix node via gRPC (`ForwardMessage`).

---

### Receiving and Forwarding Messages

1. Each node runs a gRPC server on port 50050 (mTLS-secured).
2. When a packet arrives, `ForwardMessage()` checks the replay cache first (SHA-256 hash of the raw bytes, dropped silently if seen within 30s). Then it pushes the packet onto a `Pipe` channel.
3. `receiveMessages()` reads from the Pipe, calls `DecryptLayer()` with the node's private RSA key, and gets the decrypted content, the next node, and the dummy flag.
4. If this is the final destination (nextNode == ""):
   - If dummy, drop silently.
   - Otherwise, print the decrypted message.
5. Otherwise, re-pad to 4096 bytes and push the packet to the batch channel along with the next node's name.

---

### Batching and Shuffling

1. A `batchFlusher` goroutine collects `BatchEntry` items (packet + next node) from the batch channel.
2. When `batchSize` entries are collected, it immediately spawns a goroutine to send them.
3. A ticker fires every `flushTimeoutMs` — if any entries are sitting in the buffer, they are flushed as a partial batch to guarantee forward progress under low traffic.
4. Before forwarding, the batch is randomly shuffled so the order of output packets does not match the order of input packets.

---

### Replay Protection

Each node maintains an in-memory `ReplayCache`:

- On arrival, the raw encrypted packet is hashed with SHA-256.
- The hash is looked up in a `map[[32]byte]time.Time`.
- If found within 30 seconds, the packet is silently dropped (still returns `Success: true` to the caller, so an attacker cannot distinguish a replay from a legitimate packet).
- If not found, the hash is stored with the current timestamp.
- A background goroutine evicts entries older than 30 seconds every 10 seconds.

---

### Configuration Parameters

All values live in `config.json` — edit and restart without rebuilding:

| Parameter | Default | Description |
|-----------|---------|-------------|
| `path_len` | 3 | Number of mix hops per message |
| `batch_size` | 10 | Messages collected before forced flush |
| `client_tick_us` | 200000 | Microseconds between client sends (200000 = 2 msg/s) |
| `flush_timeout_ms` | 1 | Max wait (ms) before a partial batch is flushed |

---

### Project Structure

```
mixnet/
├── mixnode/           # Node logic (gRPC handlers, batching, config, replay cache)
├── crypto/            # Onion encryption, AES-GCM, RSA-OAEP, mTLS
├── keygen/            # RSA key pair generation
├── proto/             # protobuf definitions + generated gRPC code
├── scripts/           # Setup script (keys, certs, compose, config)
├── keys/              # RSA keys (generated by script)
├── certs/             # TLS certificates (generated by script)
├── config.json        # Network parameters (generated by script)
└── docker-compose.yml # Container definitions (generated by script)
```