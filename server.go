package MyRPC

import (
	"MyRPC/codec"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
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
	serviceMap sync.Map //map[servicename]*service
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

func (server *Server) Register(rcvr interface{}) error {
	s := newService(rcvr)
	if _, loaded := server.serviceMap.LoadOrStore(s.name, s); loaded {
		return errors.New(fmt.Sprintf("rpc server: service %s already exists", s.name))
	}
	return nil
}

func Register(rcvr interface{}) error {
	return DefaultServer.Register(rcvr)
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
	req.svc, req.methods, err = server.findService(header.ServiceMethod)
	if err != nil {
		log.Println("rpc serveer find service error: " + err.Error())
		return req, err
	}
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
		log.Printf("rpc server send response error: %s", err)
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

	methodType := svc.methods[methodName]
	if methodType == nil {
		err := errors.New(fmt.Sprintf("rpc server cannot find the method: %s of service: %s\n", methodName, serviceName))
		return svc, nil, err
	}
	return svc, methodType, nil
}

var DefaultServer = NewServer()

type request struct {
	header       *codec.Header
	argv, replyv reflect.Value
	svc          *service
	methods      *MethodType
}

type MethodType struct {
	//方法
	method reflect.Method
	//方法调用需要的参数类型
	ArgType reflect.Type
	//方法调用的结果参数类型
	ReplyType reflect.Type
	numCalls  uint64
}

func (m *MethodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

func (mType *MethodType) newArgv() reflect.Value {
	var argv reflect.Value
	//// arg may be a pointer type, or a value type
	if mType.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(mType.ArgType.Elem())
	} else {
		argv = reflect.New(mType.ArgType).Elem()
	}
	return argv
}

func (mType *MethodType) newReplyv() reflect.Value {
	//reply must be a pointer type
	replyv := reflect.New(mType.ReplyType.Elem())
	switch mType.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(mType.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(mType.ReplyType.Elem(), 0, 0))
	}
	return replyv
}

type service struct {
	//映射结构体名称
	name string
	//映射结构体类型
	typ reflect.Type
	//结构体实例
	rcvr reflect.Value
	//结构体方法
	methods map[string]*MethodType
}

func newService(rcvr interface{}) *service {
	s := new(service)
	s.rcvr = reflect.ValueOf(rcvr)
	//Indirect:   if rcvr.Kind() != Pointer { return rcvr } else return rcvr.Elem()
	s.name = reflect.Indirect(s.rcvr).Type().Name()
	s.typ = reflect.TypeOf(rcvr)
	if !ast.IsExported(s.name) {
		log.Fatalf("rpc server: invalid service %s", s.name)
	}
	s.registerMethods()
	return s
}

func (s *service) registerMethods() {
	s.methods = make(map[string]*MethodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		mType := method.Type
		//符合条件的方法：3个入参（self， arg， reply），1个出参（error），arg和reply可用
		if mType.NumIn() != 3 || mType.NumOut() != 1 || mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		if !isExportedOrBuiltinType(mType.In(1)) || !isExportedOrBuiltinType(mType.In(2)) {
			continue
		}
		argType, replyType := mType.In(1), mType.In(2)
		s.methods[method.Name] = &MethodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
		log.Printf("rpc server: register %s.%s\n", s.name, method.Name)
	}
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

func (s *service) call(m *MethodType, argv, replyv reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)

	f := m.method.Func
	returnValues := f.Call([]reflect.Value{s.rcvr, argv, replyv})
	if err := returnValues[0].Interface(); err != nil {
		log.Printf("rpc server: call error: %s\n", err)
		return err.(error)
	}
	return nil
}
