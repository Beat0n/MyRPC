package main

import (
	"MyRPC"
	"MyRPC/codec"
	"encoding/json"
	"log"
	"net"
	"time"
)

const Addr = "127.0.0.1:8080"

type Foo struct{}

type Args struct {
	Arg1, Arg2 int
}

type Body struct {
	Args
	Reply *int
}

func (f *Foo) Sum(args Args, reply *int) error {
	*reply = args.Arg1 + args.Arg2
	return nil
}

func startServer() {
	var foo Foo

	l, err := net.Listen("tcp", Addr)
	if err != nil {
		log.Printf("rpc server listen error: %s\n", err)
		return
	}
	log.Println("start rpc server...")
	MyRPC.Accept(l)
	MyRPC.Register(&foo)
}
func main() {
	go startServer()

	conn, err := net.Dial("tcp", Addr)
	defer conn.Close()
	if err != nil {
		log.Printf("rpc client dial error: %s\n", err)
		return
	}
	time.Sleep(2e9)
	//传入Option
	json.NewEncoder(conn).Encode(MyRPC.DefaultOption)
	cc := codec.NewGobCodec(conn)
	for i := 0; i < 5; i++ {
		h := &codec.Header{
			Seq:           uint64(i),
			ServiceMethod: "Foo.Sum",
		}
		var reply float64
		args := Args{
			i,
			i * i,
		}
		body := &Body{
			Args:  args,
			Reply: new(int),
		}
		cc.Write(h, body)
		cc.ReadHeader(h)
		log.Printf("header: %v\n", h)

		cc.ReadBody(&reply)
		log.Printf("reply: %.2f\n", reply)
	}
}
