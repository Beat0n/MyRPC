package codec

import (
	"io"
	"net"
)

type Header struct {
	ServiceMethod string
	seq           uint64
	Error         string
}

type Type string

const (
	GobType Type = "application/gob"
)

type Codec interface {
	io.Closer
	ReadHeader(header *Header) error
	ReadBody(body interface{}) error
	Write(header *Header, body interface{}) error
}

type NewCodecFunc func(conn net.Conn) Codec

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
