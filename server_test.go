package MyRPC

import (
	"fmt"
	"reflect"
	"testing"
)

type Foo int

type Sumable interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~string
}
type Args[T Sumable] struct {
	Num1, Num2 T
}

//定义Sum参数类型
type myType string

func (f Foo) Sum(args Args[myType], reply *myType) error {
	*reply = args.Num1 + args.Num2
	return nil
}

// it's not a exported Method
func (f Foo) sum(args Args[myType], reply *myType) error {

	*reply = args.Num1 + args.Num2
	return nil
}

func _assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assertion failed: "+msg, v...))
	}
}

func TestNewService(t *testing.T) {
	var foo Foo
	s := newService(foo)
	_assert(len(s.methods) == 1, fmt.Sprintf("wrong method length: got %d, expected 1", len(s.methods)))
	mtype1 := s.methods["Sum"]
	mtype2 := s.methods["sum"]
	_assert(mtype1 != nil, fmt.Sprintf("no method 'Sum'"))
	_assert(mtype2 == nil, fmt.Sprintf("no method 'sum'"))
}

func TestMethod_call(t *testing.T) {
	var foo Foo
	s := newService(&foo)
	method := s.methods["Sum"]
	argv, replyv := method.newArgv(), method.newReplyv()
	argv.Set(reflect.ValueOf(Args[myType]{Num1: "hu", Num2: "jizhe"}))
	err := s.call(method, argv, replyv)
	_assert(err == nil, fmt.Sprintf("call 'Sum' failed\n"))
}

func TestFindService(t *testing.T) {
	var foo Foo
	Register(&foo)
	svci, ok := DefaultServer.serviceMap.Load("Foo")
	_assert(ok, "no service")
	svc := svci.(*service)
	method := svc.methods["Sum"]
	argv, replyv := method.newArgv(), method.newReplyv()
	argv.Set(reflect.ValueOf(Args[myType]{Num1: "hu", Num2: "jizhe"}))
	err := svc.call(method, argv, replyv)
	fmt.Println(replyv.Elem())
	_assert(err == nil, fmt.Sprintf("call 'Sum' failed\n"))
}
