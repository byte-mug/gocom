/*
Copyright (c) 2017 Simon Schmidt

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/


package codec

import "github.com/byte-mug/gocom/semirpc"
import "io"
import "bufio"
import "sync"
import "encoding/binary"

type ServerCodec struct{
	closer io.Closer
	reader *bufio.Reader
	writer *bufio.Writer
	buffer []byte
	mutex  sync.Mutex
}
func (c *ServerCodec) Close() error { return c.closer.Close() }
func (c *ServerCodec) Send(id uint64,r semirpc.Response) error {
	c.mutex.Lock(); defer c.mutex.Unlock()
	l := binary.PutUvarint(c.buffer, id)
	_,e := c.writer.Write(c.buffer[:l])
	if e!=nil { return e }
	e = r.WriteResp(c.writer)
	if e!=nil { return e }
	e = c.writer.Flush()
	return e
}
func (c *ServerCodec) RecvId() (uint64,error) {
	return binary.ReadUvarint(c.reader)
}
func (c *ServerCodec) Recv(r semirpc.Request) error {
	return r.ReadReq(c.reader)
}

func NewServerCodec(c io.ReadWriteCloser) *ServerCodec {
	return &ServerCodec{
		closer: c,
		reader: bufio.NewReader(c),
		writer: bufio.NewWriter(c),
		buffer: make([]byte,16),
	}
}


type ClientCodec struct{
	closer io.Closer
	reader *bufio.Reader
	writer *bufio.Writer
	buffer []byte
	mutex  sync.Mutex
}
func (c *ClientCodec) Close() error { return c.closer.Close() }
func (c *ClientCodec) Send(id uint64,r semirpc.Request) error {
	c.mutex.Lock(); defer c.mutex.Unlock()
	l := binary.PutUvarint(c.buffer, id)
	_,e := c.writer.Write(c.buffer[:l])
	if e!=nil { return e }
	e = r.WriteReq(c.writer)
	if e!=nil { return e }
	e = c.writer.Flush()
	return e
}
func (c *ClientCodec) RecvId() (uint64,error) {
	return binary.ReadUvarint(c.reader)
}
func (c *ClientCodec) Recv(r semirpc.Response) error {
	return r.ReadResp(c.reader)
}

func NewClientCodec(c io.ReadWriteCloser) *ClientCodec {
	return &ClientCodec{
		closer: c,
		reader: bufio.NewReader(c),
		writer: bufio.NewWriter(c),
		buffer: make([]byte,16),
	}
}

