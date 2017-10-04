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


package fastcodec

import "github.com/valyala/batcher"
import "github.com/pierrec/lz4"
import "github.com/byte-mug/gocom/semirpc"
import "io"
import "bytes"
import "bufio"
import "sync"
import "encoding/binary"
import "time"

var pool_workItem sync.Pool
type workItem struct{
	tmpbuf   [16]byte
	buf      bytes.Buffer
	bwri     *bufio.Writer // Adapter
}
func (wi *workItem) writeId(id uint64) {
	l := binary.PutUvarint(wi.tmpbuf[:], id)
	wi.buf.Write(wi.tmpbuf[:l])
}

func acquireWI() *workItem {
	wi := pool_workItem.Get()
	if wi==nil {
		nwi := &workItem{}
		nwi.bwri = bufio.NewWriterSize(&(nwi.buf),128)
		return nwi
	}
	return wi.(*workItem)
}
func releaseWI(wi *workItem) {
	wi.bwri.Flush()
	wi.buf.Reset()
	pool_workItem.Put(wi)
}

type PlainCodec struct{
	closer  io.Closer
	reader  *bufio.Reader
	lzrd    *lz4.Reader
	lzwr    *lz4.Writer
	buffer  []byte
	mutex   sync.Mutex
	coalesc batcher.Batcher
}
func (pc *PlainCodec) initialize(c io.ReadWriteCloser) {
	rd  := lz4.NewReader(c)
	wri := lz4.NewWriter(c)
	
	pc.closer = c
	pc.reader = bufio.NewReader(rd)
	pc.lzrd   = rd
	pc.lzwr   = wri
	pc.buffer = make([]byte,16)
	
	pc.coalesc.Func = pc.write
	pc.coalesc.MaxDelay = time.Millisecond * 50
	pc.coalesc.Start()
}
func (c *PlainCodec) Close() error {
	c.coalesc.Stop()
	return c.closer.Close()
}

func (c *PlainCodec) write(batch []interface{}) {
	c.mutex.Lock(); defer c.mutex.Unlock()
	for _,el := range batch {
		wi := el.(*workItem)
		c.lzwr.Write(wi.buf.Bytes())
		releaseWI(wi)
	}
	c.lzwr.Flush()
}
func (c *PlainCodec) RecvId() (uint64,error) {
	return binary.ReadUvarint(c.reader)
}



func NewServerCodec(c io.ReadWriteCloser) *ServerCodec {
	sc := new(ServerCodec)
	sc.initialize(c)
	return sc
}

type ServerCodec struct{
	PlainCodec
}
func (c *ServerCodec) Send(id uint64,r semirpc.Response) error {
	wi := acquireWI()
	wi.writeId(id)
	e := r.WriteResp(wi.bwri)
	wi.bwri.Flush()
	c.coalesc.Push(wi)
	return e
}
func (c *ServerCodec) Recv(r semirpc.Request) error {
	return r.ReadReq(c.reader)
}



func NewClientCodec(c io.ReadWriteCloser) *ClientCodec {
	sc := new(ClientCodec)
	sc.initialize(c)
	return sc
}

type ClientCodec struct{
	PlainCodec
}
func (c *ClientCodec) Send(id uint64,r semirpc.Request) error {
	wi := acquireWI()
	wi.writeId(id)
	e := r.WriteReq(wi.bwri)
	wi.bwri.Flush()
	c.coalesc.Push(wi)
	return e
}
func (c *ClientCodec) Recv(r semirpc.Response) error {
	return r.ReadResp(c.reader)
}


