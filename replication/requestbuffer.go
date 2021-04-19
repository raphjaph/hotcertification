package main

import (
	"container/list"
	"fmt"
	"sync"

	"github.com/raphasch/hotcertification/protocol"
	"github.com/relab/hotstuff"
	"google.golang.org/protobuf/proto"
)

type reqBuffer struct {
	mut           sync.Mutex
	batchSize     int
	serialNumbers map[uint32]uint64 // highest proposed serial number per client ID
	cache         list.List
	marshaler     proto.MarshalOptions
	unmarshaler   proto.UnmarshalOptions
}

func newReqBuffer(batchSize int) *reqBuffer {
	return &reqBuffer{
		batchSize:     batchSize,
		serialNumbers: make(map[uint32]uint64),
		marshaler:     proto.MarshalOptions{Deterministic: true},
		unmarshaler:   proto.UnmarshalOptions{DiscardUnknown: true},
	}
}

func (cache *reqBuffer) addRequest(req *protocol.Request) {
	cache.mut.Lock()
	defer cache.mut.Unlock()
	/* if serialNo := cache.serialNumbers[req.GetClientID()]; serialNo >= req.GetSequenceNumber() {
		// command is too old
		return
	} */
	if req.Data == "Invalid CSR" {

	}

	cache.cache.PushBack(req)
}

// transform the requests into a HotStuff Command (string)
func (cache *reqBuffer) GetCommand() *hotstuff.Command {
	cache.mut.Lock()
	defer cache.mut.Unlock()

	if cache.cache.Len() == 0 {
		return nil
	}

	batch := new(protocol.Batch)

	for i := 0; i < cache.batchSize; i++ {
		elem := cache.cache.Front()
		if elem == nil {
			break
		}
		cache.cache.Remove(elem)
		request := elem.Value.(*protocol.Request)
		/* if serialNo := cache.serialNumbers[cmd.GetClientID()]; serialNo >= cmd.GetSequenceNumber() {
			// command is too old, can't propose
			continue
		} */
		batch.Requests = append(batch.Requests, request)
	}

	b, err := cache.marshaler.Marshal(batch)
	if err != nil {
		return nil
	}

	cmd := hotstuff.Command(b)
	return &cmd
}

// HotStuff only works with strings (Command) so this string has to be parsed with protocol buffers to get the single Requests
// This function checks wether a CSR is valid/well-formed/...
func (cache *reqBuffer) Accept(cmd hotstuff.Command) bool {
	// Parses from string to Request format
	batch := new(protocol.Batch)
	err := cache.unmarshaler.Unmarshal([]byte(cmd), batch)
	if err != nil {
		return false
	}

	cache.mut.Lock()
	defer cache.mut.Unlock()
	for _, request := range batch.GetRequests() {
		/* if serialNo := cache.serialNumbers[cmd.GetClientID()]; serialNo >= cmd.GetSequenceNumber() {
			// command is too old, can't accept
			return false
		}
		cache.serialNumbers[cmd.GetClientID()] = cmd.GetSequenceNumber()
		*/

		// IMPORTANT: This is where the logic for validating a CSR goes
		if request.Data == "Invalid CSR" {
			fmt.Println("Invalid Certificate Signing Request")
			return false
		}
	}

	return true
}
