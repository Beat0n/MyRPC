package MyRPC

import (
	"MyRPC/codec"
	"encoding/json"
	"log"
	"net"
	"reflect"
	"sync"
)

const MagicNumber = 0x000001

type Option struct {
	MagicNumber int // MagicNumber marks this is a rpc request
	CodecType   codec.Type
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

type Server struct {
	serviceMap sync.Map
}

func NewServer() *Server {
	return &Server{}
}

func (server *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Printf("rpc server accept error: %s\n", err)
			return
		}
		go server.ServeConn(conn)
	}
}

func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}

func (server *Server) ServeConn(conn net.Conn) {
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Printf("rpc server decode option error: %s\n", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server receive invalid magicnumber option: %d\n", opt.MagicNumber)
		return
	}
	f, ok := codec.NewCodecFuncMap[opt.CodecType]
	if !ok {
		log.Printf("rpc server receive invalid codectype option: %s\n", opt.CodecType)
	}
	cc := f(conn)

	//保证并发处理请求时能够正确回复
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)

	//1.read request header

	//2.read request

	//3/handle request
	for {
		req, err := server.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}
			req.header.Error = err.Error()
			server.sendResponse(req)
		}
	}

}

func (server *Server) readRequest(cc codec.Codec) (req *request, err error) {

}

var DefaultServer = NewServer()

type request struct {
	header       *codec.Header
	argv, replyv reflect.Value
	svc          *service
	methodtype   *MethodType
}

type MethodType struct {
	method    *reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
}

type service struct {
	//映射结构体名称
	name string
	//映射结构体类型
	typ reflect.Type
	//结构体实例
	rcvr reflect.Value
	//结构体方法
	method map[string]*MethodType
}
