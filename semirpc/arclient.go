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

import (
	"time"
	"fmt"
)

type Dialer func(addr string) (ClientCodec,error)

// Auto-Reconnect-Client
type ARClient struct{
	client *Client
	
	Addr   string
	Dial   Dialer
	MkResp func() Response
}
func (c *ARClient) connect() error {
	if c.client!=nil { c.client.Close() ; c.client = nil }
	cc,err := c.Dial(c.Addr)
	if err!=nil { return err }
	c.client = NewClient(cc,c.MkResp)
	return nil
}
func (c *ARClient) doit(req Request,resp Response) (*clientObject,error) {
	if c.client==nil {
		if err := c.connect() ; err!=nil { return nil,err }
	}
	co,err := c.client.doit(req,resp)
	if err!=nil && err!=EPipelineStall { // IO-Error, Connection fail
		if err := c.connect() ; err!=nil { return nil,err }
		return c.client.doit(req,resp)
	}
	return co,err
}
func (c *ARClient) Close() (err error) {
	if c.client!=nil {
		err = c.client.Close()
		c.client = nil
	}
	return
}
func (c *ARClient) Do(req Request,resp Response) error {
	co,e := c.doit(req,resp)
	if co!=nil { defer co.free() }
	if e!=nil { return e }
	e = <- co.signal
	co.signal <- e
	return e
}
func (c *ARClient) DoTimeout(req Request,resp Response, timeout <- chan time.Time) error {
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
func (c *ARClient) DoAsyncSupport(req Request,resp Response, errch chan error) {
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
