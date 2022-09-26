package MyRPC

import (
	"MyRPC/codec"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
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

func NewClient(conn net.Conn, opt *Option) (*Client, error) {
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("invalid codec type %s", opt.CodecType)
		log.Println("rpc client: codec error:", err)
		return nil, err
	}

	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc client: options error: ", err)
		conn.Close()
		return nil, err
	}

	return newClientCodec(f(conn), opt), nil
}

func newClientCodec(cc codec.Codec, opt *Option) *Client {
	client := &Client{
		cc:             cc,
		seq:            1,
		opt:            opt,
		unHandledCalls: make(map[uint64]*Call),
	}
	go client.receive()
	return client
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

func (c *Client) registerCall(call *Call) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.shutdown || c.closing {
		return 0, Errshutdown
	}

	call.Seq = c.seq
	c.seq++
	c.unHandledCalls[call.Seq] = call
	return call.Seq, nil
}

func (c *Client) removeCall(seq uint64) *Call {
	c.mu.Lock()
	defer c.mu.Unlock()
	call := c.unHandledCalls[seq]
	delete(c.unHandledCalls, seq)
	return call
}

func (c *Client) terminateCalls(err error) {
	c.sending.Lock()
	defer c.sending.Unlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.shutdown = true
	for _, call := range c.unHandledCalls {
		call.Error = err
		call.done()
	}
}

func (c *Client) receive() {

}
