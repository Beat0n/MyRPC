package MyRPC

import (
	"MyRPC/codec"
	"errors"
	"sync"
)

type Call struct {
	Seq          uint64
	ServicMethod string
	Args         any
	Reply        any
	Error        error
	Done         chan *Call
}

func (call *Call) done() {
	call.Done <- call
}

type Client struct {
	cc             codec.Codec
	opt            *Option
	sending        sync.Mutex //protect following
	header         *codec.Header
	mu             sync.Mutex //protect following
	seq            uint64
	unHandledCalls map[uint64]*Call
	closing        bool //user close
	shutdown       bool //server tells client to shutdown
}

var Errshutdown = errors.New("connection already shutdown")

func (c *Client) IsAvailable() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return !c.closing && !c.shutdown
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closing {
		return Errshutdown
	}
	c.closing = true
	return c.cc.Close()
}
