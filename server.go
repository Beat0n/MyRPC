package MyRPC

import (
	"MyRPC/codec"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
)

const MagicNumber = 0x000001

type Option struct {
	MagicNumber int        `json:"magicNumber"` // MagicNumber marks this is a rpc request
	CodecType   codec.Type `json:"codecType"`
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

var invalidRequest = struct{}{}

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

	//4.send response
	for {
		req, err := server.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}
			req.header.Error = err.Error()
			server.sendResponse(cc, req.header, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go server.handleRequest(cc, req, wg, sending)
	}

	wg.Wait()
	defer cc.Close()
}

func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	header := new(codec.Header)
	err := cc.ReadHeader(header)
	if err != nil {
		log.Printf("rpc server read request header error: %s\n", err)
		return nil, err
	}
	req := &request{header: header}
	req.svc, req.methodtype = server.findService(header.ServiceMethod)
	req.argv = reflect.New(reflect.TypeOf(""))
	if err = cc.ReadBody(req.argv.Interface()); err != nil {
		log.Println("rpc server: read argv err:", err)
	}
	return req, nil
}

func (server *Server) handleRequest(cc codec.Codec, req *request, wg *sync.WaitGroup, sending *sync.Mutex) {
	defer wg.Done()
	log.Println("Server:", req.header, req.argv.Elem())
	req.replyv = reflect.ValueOf(fmt.Sprintf("geerpc resp %d", req.header.Seq))
	server.sendResponse(cc, req.header, req.replyv.Interface(), sending)
}

func (server *Server) sendResponse(cc codec.Codec, header *codec.Header, body interface{}, sending *sync.Mutex) error {
	sending.Lock()
	defer sending.Unlock()

	if err := cc.Write(header, body); err != nil {
		log.Println("rpc server send response error: %d", err)
		return err
	}
	return nil
}

func (server *Server) findService(serviceMethod string) (*service, *MethodType, error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot <= 0 {
		err := errors.New("rpc server receive malformed service method")
		return nil, nil, err
	}
	serviceName := serviceMethod[:dot]
	methodName := serviceMethod[dot+1:]

	svci, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err := errors.New("rpc server cannot find service: " + serviceName)
		return nil, nil, err
	}
	svc := svci.(*service)

}

var DefaultServer = NewServer()

type request struct {
	header       *codec.Header
	argv, replyv reflect.Value
	svc          *service
	methodtype   *MethodType
}

type MethodType struct {
	//方法
	method *reflect.Method
	//方法调用需要的参数类型
	ArgType reflect.Type
	//方法调用的结果参数类型
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
