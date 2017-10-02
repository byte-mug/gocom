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

type Response struct{
	Header
	
	code   int
	reason []byte
}
func (r *Response) Reset() {
	r.Header.Reset()
	r.code   = 0
	r.reason = nil
}

func (r *Response) Code  () int    { return r.code   }
func (r *Response) Reason() []byte { return r.reason }

func (r *Response) SetCode     (i int   ) { r.code   = i          }
func (r *Response) SetReason   (b []byte) { r.reason = r.cbbin(b) }
func (r *Response) SetReasonStr(b string) { r.reason = r.cbstr(b) }

func (r *Response) ReadResp(rd *bufio.Reader) error {
	r.Reset()
	code,reason,err := r.rdCodeReason(rd)
	r.code   = code
	r.reason = reason
	if err!=nil { return err }
	return r.read(rd)
}
func (r *Response) WriteResp(wd *bufio.Writer) error {
	code := encodeInt(r.tempbuf[:],r.code)
	_,err := wd.Write(code)
	if err!=nil { return err }
	err = wd.WriteByte(' ')
	if err!=nil { return err }
	_,err = wd.Write(r.reason)
	if err!=nil { return err }
	err = wd.WriteByte('\n')
	if err!=nil { return err }
	return r.write(wd)
}

