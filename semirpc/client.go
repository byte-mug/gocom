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


package semirpc

import "sync"
import "sync/atomic"
import "time"
import "fmt"

type ClientCodec interface {
	Close() error
	
	// Threadsafe
	Send(id uint64,r Request) error
	
	// Not threadsafe
	RecvId() (uint64,error)
	Recv(r Response) error
}


var pool_clientObject sync.Pool
type clientObject struct {
	mutex  sync.Mutex
	resp   Response
	signal chan error
	refc   *int32
}
func (c *clientObject) free() {
	if atomic.AddInt32(c.refc,-1) > 0 { return }
	select {
	case <- c.signal:
	default:
	}
	c.resp = nil
	pool_clientObject.Put(c)
}
func (c *clientObject) cancel() {
	c.mutex.Lock()
	c.resp = nil
	c.mutex.Unlock()
}

type Client struct {
	client ClientCodec
	respmk func() Response
	
	mutex  sync.Mutex
	hmap   map[uint64]*clientObject
	nextid uint64
}

func NewClient(client ClientCodec,respmk func() Response) *Client {
	c := &Client{
		client:client,
		respmk:respmk,
		hmap  :make(map[uint64]*clientObject),
	}
	go c.input()
	return c
}

func (c *Client) input() error {
	defer func (){
		c.mutex.Lock()
			for id,co := range c.hmap {
				co.cancel()
				select {
				case co.signal <- EClosed:
				default:
				}
				delete(c.hmap,id)
			}
		c.mutex.Unlock()
	}()
	for{
		id,err := c.client.RecvId()
		if err!=nil { return err }
		
		c.mutex.Lock()
			co := c.hmap[id]
			delete(c.hmap,id)
		c.mutex.Unlock()
		
		var resp Response
		
		if co!=nil {
			co.mutex.Lock()
			resp = co.resp
		}
		
		if resp!=nil { resp = c.respmk() }
		
		err = c.client.Recv(resp)
		
		if co!=nil { co.mutex.Unlock(); co.signal <- err; co.free() }
		if err!=nil { return err }
	}
}
func (c *Client) Close() error { return c.client.Close() }

func (c *Client) doit(req Request,resp Response) (*clientObject,error) {
	var co *clientObject
	elem := pool_clientObject.Get()
	if elem!=nil {
		co = elem.(*clientObject)
		co.resp = resp
	} else {
		co = &clientObject{
			resp: resp,
			signal: make(chan error,1),
			refc: new(int32),
		}
	}
	atomic.AddInt32(co.refc,1)
	
	c.mutex.Lock()
	_,ok := c.hmap[c.nextid]
	c.nextid++
	for i:=0 ; ok && i<3 ; i++ {
		_,ok = c.hmap[c.nextid]
		c.nextid++
	}
	if ok {
		c.mutex.Unlock()
		return co,EPipelineStall
	}
	id := c.nextid
	c.hmap[id] = co
	c.mutex.Unlock()
	
	atomic.AddInt32(co.refc,1)
	
	err := c.client.Send(id,req)
	if err!=nil {
		c.client.Close()
		return nil,err // discard the *clientObject to the GC.
	}
	
	return co,nil
}
func (c *Client) Do(req Request,resp Response) error {
	co,e := c.doit(req,resp)
	if co!=nil { defer co.free() }
	if e!=nil { return e }
	e = <- co.signal
	co.signal <- e
	return e
}
func (c *Client) DoTimeout(req Request,resp Response, timeout <- chan time.Time) error {
	co,e := c.doit(req,resp)
	if co!=nil { defer co.free() }
	if e!=nil { return e }
	select {
	case e := <- co.signal:
		co.signal <- e
		return e
	case tm := <- timeout:
		co.cancel()
		return fmt.Errorf("Timeout -> %v",tm)
	}
	panic("unreachable")
}
func (c *Client) DoAsyncSupport(req Request,resp Response, errch chan error) {
	co,e := c.doit(req,resp)
	if co!=nil { defer co.free() }
	if e!=nil { return }
	select {
	case e := <- co.signal:
		co.signal <- e
		errch <- e
	case e := <- errch:
		co.cancel()
		errch <- e
	}
}


