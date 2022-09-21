package codec

import (
	"bufio"
	"encoding/gob"
	"log"
	"net"
)

type GobCodec struct {
	conn    net.Conn
	encoder *gob.Encoder
	decoder *gob.Decoder
	buf     *bufio.Writer
}

func (g GobCodec) Close() error {
	return g.conn.Close()
}

func (g GobCodec) ReadHeader(header *Header) error {
	return g.decoder.Decode(header)
}

func (g GobCodec) ReadBody(body interface{}) error {
	return g.decoder.Decode(body)
}

func (g GobCodec) Write(header *Header, body interface{}) (err error) {
	defer func() {
		if err == nil {
			g.buf.Flush()
			g.Close()
		}
	}()
	if err := g.encoder.Encode(header); err != nil {
		log.Printf("rpc encode header error: %s", err)
		return err
	}
	if err := g.encoder.Encode(body); err != nil {
		log.Printf("rpc encode body error: %s", err)
		return err
	}
	return nil
}

func NewGobCodec(conn net.Conn) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn:    conn,
		buf:     buf,
		decoder: gob.NewDecoder(conn),
		encoder: gob.NewEncoder(buf),
	}
}
