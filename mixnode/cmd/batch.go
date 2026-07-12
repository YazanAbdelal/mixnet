package main

import (
	"math/rand"
	"time"

	"github.com/YazanAbdelal/mixnet/mixnode"
	pb "github.com/YazanAbdelal/mixnet/proto/gen"
)

// batchFlusher receives packets from other nodes, and sends them periodically in batches.
func batchFlusher(stubs map[string]pb.MessagingClient, batchChan <-chan mixnode.BatchEntry) {
	batch := make([]mixnode.BatchEntry, 0, BatchSize)
	ticker := time.NewTicker(time.Duration(BatchInterval) * time.Microsecond)
	defer ticker.Stop()

	for {
		select { // simultaneously check both channels
		case entry := <-batchChan:
			batch = append(batch, entry)
		case <-ticker.C:
			if len(batch) >= BatchSize {
				// running this as a goroutine is essential, without it we get a deadlock, and the context in sendToPeer runs out.
				go flushBatch(stubs, batch[:BatchSize])
				batch = batch[BatchSize:]
			}
		}
	}
}

// flushBatch sends each packet in the batch to its destiation.
func flushBatch(stubs map[string]pb.MessagingClient, toSend []mixnode.BatchEntry) {
	// shuffle batch first.
	rand.Shuffle(len(toSend), func(i, j int) {
		toSend[i], toSend[j] = toSend[j], toSend[i]
	})

	// send each packet to the respective node
	for _, entry := range toSend {
		sendToPeer(stubs[entry.NextNode], entry.Packet)
	}
}
