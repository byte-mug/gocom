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


package notrest

import "github.com/byte-mug/gocom/semirpc"
import "time"

func ackquireFunc() semirpc.Response {
	return AckquireResponse()
}

type rpcClient interface{
	Close() error
	Do(req semirpc.Request,resp semirpc.Response) error
	DoTimeout(req semirpc.Request,resp semirpc.Response, timeout <- chan time.Time) error
	DoAsyncSupport(req semirpc.Request,resp semirpc.Response, errch chan error)
}

type Client struct{
	client rpcClient
}
func NewClient(cli semirpc.ClientCodec) *Client {
	return &Client{
		client: semirpc.NewClient(cli, ackquireFunc),
	}
}
func NewARClient(cli *semirpc.ARClient) *Client {
	return &Client{
		client: cli,
	}
}


func (c *Client) Close() error { return c.client.Close() }

func (c *Client) Do(req *Request, resp *Response) error {
	return c.client.Do(req,resp)
}

func (c *Client) DoTimeout(req *Request, resp *Response, timeout <-chan time.Time) error {
	return c.client.DoTimeout(req,resp,timeout)
}

func (c *Client) DoAsyncSupport(req *Request, resp *Response, errch chan error) {
	c.client.DoAsyncSupport(req,resp,errch)
}
