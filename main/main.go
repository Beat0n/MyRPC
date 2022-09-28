package main

import (
	"MyRPC"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo) ToUpper(s string, reply *string) error {
	*reply = strings.ToUpper(s)
	return nil
}

func startServer(addr chan string) {
	var foo Foo
	if err := MyRPC.Register(&foo); err != nil {
		log.Fatal("register error:", err)
	}
	// pick a free port
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	MyRPC.Accept(l)
}

func main() {
	log.SetFlags(0)
	addr := make(chan string)
	go startServer(addr)

	for n := 0; n < 5; n++ {
		go func() {
			client, _ := MyRPC.Dial("tcp", <-addr)
			defer client.Close()
			time.Sleep(1e9)

			// send request & receive response
			var wg sync.WaitGroup
			for i := 0; i < 5; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					args := &Args{Num1: i, Num2: i * i}
					var reply int
					if err := client.Call("Foo.Sum", args, &reply); err != nil {
						log.Fatal("call Foo.Sum error:", err)
					}
					log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
				}(i)
			}
			var stringReply string
			if err := client.Call("Foo.ToUpper", "hjz", &stringReply); err != nil {
				log.Fatal("call Foo.ToUpper error:", err)
			}
			log.Println(stringReply)
			wg.Wait()
		}()
	}
	time.Sleep(10e9)
}
