package main

import (
	"MyRPC"
	"MyRPC/codec"
	"encoding/json"
	"log"
	"net"
)

const Addr = "127.0.0.1:8080"

func startServer() {
	l, _ := net.Listen("tcp", Addr)
	conn, err := l.Accept()
	if err != nil {
		log.Printf("rpc server accept error: %s\n", err)
		return
	}
	cc := codec.NewGobCodec(conn)

}
func main() {
	conn, err := net.Dial("tcp", Addr)
	//传入Option
	json.NewEncoder(conn).Encode(MyRPC.DefaultOption)
	cc := codec.NewGobCodec(conn)
	cc.Write()
}
