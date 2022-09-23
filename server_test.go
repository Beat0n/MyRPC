package MyRPC

import (
	"fmt"
	"reflect"
	"testing"
)

type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

// it's not a exported Method
func (f Foo) sum(args Args, reply *int) error {
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
	argv.Set(reflect.ValueOf(Args{Num1: 1, Num2: 3}))
	err := s.call(method, argv, replyv)
	_assert(err == nil, fmt.Sprintf("call 'Sum' failed\n"))
}
