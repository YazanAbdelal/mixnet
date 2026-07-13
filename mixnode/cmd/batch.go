package main

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/YazanAbdelal/mixnet/mixnode"
	pb "github.com/YazanAbdelal/mixnet/proto/gen"
)

// batchFlusher receives packets from other nodes, and sends them in batches.
// A batch is flushed when either batchSize messages have been collected (threshold), or flushTimeout has elapsed since the last flush (timeout).
// The timeout guarantees forward progress even under low traffic.
func batchFlusher(stubs map[string]pb.MessagingClient, batchChan <-chan mixnode.BatchEntry,
	batchSize int, flushTimeout time.Duration) {
	batch := make([]mixnode.BatchEntry, 0, batchSize)
	ticker := time.NewTicker(flushTimeout)
	defer ticker.Stop()

	for {
		select {
		case entry := <-batchChan:
			batch = append(batch, entry)
			if len(batch) >= batchSize {
				go flushBatch(stubs, batch[:batchSize])
				batch = batch[batchSize:]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				go flushBatch(stubs, batch)
				batch = batch[:0]
			}
		}
	}
}

// flushBatch sends each packet in the batch to its destination.
func flushBatch(stubs map[string]pb.MessagingClient, toSend []mixnode.BatchEntry) {
	for i := len(toSend) - 1; i > 0; i-- {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		idx := j.Int64()
		toSend[i], toSend[idx] = toSend[idx], toSend[i]
	}

	for _, entry := range toSend {
		sendToPeer(stubs[entry.NextNode], entry.Packet)
	}
}
