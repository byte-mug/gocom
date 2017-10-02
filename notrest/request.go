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

import "bufio"
import "sync"

var pool_Request sync.Pool

func AckquireRequest() *Request {
	r := pool_Request.Get()
	if r!=nil { return new(Request) }
	return r.(*Request)
}
func ReleaseRequest(r *Request) {
	r.Reset()
	pool_Request.Put(r)
}

type Request struct{
	Header
	
	method  []byte
	path    []byte
}
func (r *Request) Reset() {
	r.Header.Reset()
	r.method = nil
	r.path   = nil
}
func (r *Request) Method() []byte { return r.method }
func (r *Request) Path()   []byte { return r.path }

func (r *Request) SetMethod(k []byte) {
	r.method = r.cbbin(k)
	uperize(r.method)
}
func (r *Request) SetMethodStr(k string) {
	r.method = r.cbstr(k)
	uperize(r.method)
}
func (r *Request) SetPath(k []byte) {
	r.path = r.cbbin(k)
}
func (r *Request) SetPathStr(k string) {
	r.path = r.cbstr(k)
}
func (r *Request) ReadReq(rd *bufio.Reader) error {
	r.Reset()
	meth,path,err := r.rdMethodPath(rd)
	r.method = meth
	r.path   = path
	if err!=nil { return err }
	return r.read(rd)
}
func (r *Request) WriteReq(wd *bufio.Writer) error {
	_,err := wd.Write(r.method)
	if err!=nil { return err }
	err = wd.WriteByte(' ')
	if err!=nil { return err }
	_,err = wd.Write(r.path)
	if err!=nil { return err }
	err = wd.WriteByte('\n')
	if err!=nil { return err }
	return r.write(wd)
}

