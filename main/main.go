package main

import (
	"MyRPC"
	"MyRPC/codec"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

const Addr = "127.0.0.1:8080"

func startServer() {
	l, err := net.Listen("tcp", Addr)
	if err != nil {
		log.Printf("rpc server listen error: %s\n", err)
		return
	}
	log.Println("start rpc server...")
	MyRPC.Accept(l)
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
		cc.Write(h, fmt.Sprintf("geerpc req %d", h.Seq))
		cc.ReadHeader(h)
		log.Printf("header: %v\n", h)
		var reply string
		cc.ReadBody(&reply)
		log.Printf("reply: %s\n", reply)
	}
}
