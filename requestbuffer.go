package hotcertification

import (
	"container/list"
	"context"
	"sync"

	"github.com/raphasch/hotcertification/client"
	"github.com/relab/hotstuff"
	"google.golang.org/protobuf/proto"
)

type ReqBuffer struct {
	mut         sync.Mutex
	mod         *hotstuff.HotStuff
	buffer      list.List
	marshaler   proto.MarshalOptions
	unmarshaler proto.UnmarshalOptions
}

func NewReqBuffer() *ReqBuffer {
	return &ReqBuffer{
		marshaler:   proto.MarshalOptions{Deterministic: true},
		unmarshaler: proto.UnmarshalOptions{DiscardUnknown: true},
	}
}

// InitModule gives the module a reference to the HotStuff object.
func (r *ReqBuffer) InitModule(hs *hotstuff.HotStuff, _ *hotstuff.ConfigBuilder) {
	r.mod = hs
}

func (r *ReqBuffer) AddRequest(req *client.CSR) {
	r.mut.Lock()
	defer r.mut.Unlock()

	r.buffer.PushBack(req)
}

// Get returns a batch of commands to propose.
// Batching not supported yet
func (r *ReqBuffer) Get(ctx context.Context) (cmd hotstuff.Command, ok bool) {
	// Batching not supported yet
	batch := new(client.Batch)

	r.mut.Lock()
	defer r.mut.Unlock()

	elem := r.buffer.Front()
	if elem == nil {
		return "", false
	}

	// TODO: if i have this i get a deadlock
	//r.buffer.Remove(elem)

	req := elem.Value.(*client.CSR)
	batch.CSRs = append(batch.CSRs, req)

	b, err := r.marshaler.Marshal(batch)
	if err != nil {
		r.mod.Logger().Errorf("Failed to marshal batch: %v", err)
		return "", false
	}

	// HotStuff replicates strings; the semantics are validated by the Accept() function
	cmd = hotstuff.Command(b)
	return cmd, true
}

// Accept returns true if the replica can accept the batch.
func (r *ReqBuffer) Accept(cmd hotstuff.Command) bool {
	batch := new(client.Batch)
	err := r.unmarshaler.Unmarshal([]byte(cmd), batch)
	if err != nil {
		r.mod.Logger().Errorf("Failed to unmarshal batch: %v", err)
		return false
	}

	r.mut.Lock()
	defer r.mut.Unlock()

	for _, req := range batch.GetCSRs() {
		// TODO: validate the CSR here
		r.mod.Logger().Infof("Validating the request from client %v", req.ClientID)
	}

	return true
}

// Proposed updates the serial numbers such that we will not accept the given batch again.
func (r *ReqBuffer) Proposed(cmd hotstuff.Command) {
	batch := new(client.Batch)
	err := r.unmarshaler.Unmarshal([]byte(cmd), batch)
	if err != nil {
		r.mod.Logger().Errorf("Failed to unmarshal batch: %v", err)
		return
	}

	r.mut.Lock()
	defer r.mut.Unlock()

	// Updating the state of a specific CSR
	for _, req := range batch.GetCSRs() {
		r.mod.Logger().Infof("Updating state of CSR from client %v to replicating.", req.ClientID)
	}
}

var _ hotstuff.Acceptor = (*ReqBuffer)(nil)
