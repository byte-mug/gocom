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

import "io"
import "bufio"
import "crypto/md5"
import "github.com/valyala/bytebufferpool"

type Header struct{
	tempbuf [512]byte
	alloc   bufferAlloc
	
	headKV  [][2][]byte
	headIdx indexMap
	
	body    *bytebufferpool.ByteBuffer
}
func (r *Header) Reset() {
	r.alloc.reset()
	r.headIdx.clear()
	r.headKV = r.headKV[:0]
	if r.body!=nil { bufpool.Put(r.body) ; r.body=nil }
}
func (r *Header) cbbin(b []byte) []byte {
	nb := r.alloc.alloc(len(b))
	if len(nb)!=len(b) { return nil }
	copy(nb,b)
	return nb
}
func (r *Header) cbstr(b string) []byte {
	nb := r.alloc.alloc(len(b))
	if len(nb)!=len(b) { return nil }
	copy(nb,b)
	return nb
}
func (r *Header) SetHeader(k,v []byte) {
	nk := append(r.tempbuf[:0],k...)
	lowerize(nk)
	sum := md5.Sum(nk)
	hix := r.headIdx.init()
	
	if i,ok := hix[sum] ; ok {
		h := r.headKV[i][1]
		if len(h)<len(v) {
			r.headKV[i][1] = r.cbbin(v)
		} else {
			copy(h,v)
			r.headKV[i][1] = h[:len(v)]
		}
	} else {
		nk = r.cbbin(k)
		nv := r.cbbin(v)
		lowerize(nk)
		r.headKV = append(r.headKV,[2][]byte{nk,nv})
		hix[sum] = len(r.headKV)-1
	}
}
func (r *Header) SetHeaderKV(k,v string) {
	nk := append(r.tempbuf[:0],k...)
	lowerize(nk)
	sum := md5.Sum(nk)
	hix := r.headIdx.init()
	
	if i,ok := hix[sum] ; ok {
		h := r.headKV[i][1]
		if len(h)<len(v) {
			r.headKV[i][1] = r.cbstr(v)
		} else {
			copy(h,v)
			r.headKV[i][1] = h[:len(v)]
		}
	} else {
		nk = r.cbstr(v)
		nv := r.cbstr(v)
		lowerize(nk)
		r.headKV = append(r.headKV,[2][]byte{nk,nv})
		hix[sum] = len(r.headKV)-1
	}
}
func (r *Header) GetHeader(k []byte) []byte {
	if r.headIdx==nil { return nil }
	nk := append(r.tempbuf[:0],k...)
	lowerize(nk)
	sum := md5.Sum(nk)
	i,ok := r.headIdx[sum]
	if !ok { return nil }
	return r.headKV[i][1]
}
func (r *Header) GetHeaderK(k string) []byte {
	if r.headIdx==nil { return nil }
	nk := append(r.tempbuf[:0],k...)
	lowerize(nk)
	sum := md5.Sum(nk)
	i,ok := r.headIdx[sum]
	if !ok { return nil }
	return r.headKV[i][1]
}
func (r *Header) Body() *bytebufferpool.ByteBuffer {
	if r.body==nil { r.body = bufpool.Get() }
	return r.body
}

func (r *Header) rdMethodPath(rd *bufio.Reader) ([]byte,[]byte,error) {
	buf,e := rd.ReadSlice(' ')
	if e!=nil { return nil,nil,e }
	method := r.cbbin(stripTail(buf,' '))
	uperize(method)
	buf,e = rd.ReadSlice('\n')
	if e!=nil { return method,nil,e }
	pth := r.cbbin(stripTail(buf,'\n'))
	return method,pth,nil
}
func (r *Header) rdCodeReason(rd *bufio.Reader) (int,[]byte,error) {
	buf,e := rd.ReadSlice(' ')
	if e!=nil { return 0,nil,e }
	i := 0
	for _,b := range buf {
		if b<'0' || '9'<b { continue }
		i = (i*10)+int(b-'0')
	}
	
	buf,e = rd.ReadSlice('\n')
	if e!=nil { return i,nil,e }
	pth := r.cbbin(stripTail(buf,'\n'))
	return i,pth,nil
}


var contentLength = []byte("content-length")

func (r *Header) readHead(rd *bufio.Reader) error {
	for {
		{
			b,e := rd.ReadByte()
			if e!=nil { return e }
			if b=='\n' { break }
			e = rd.UnreadByte()
			if e!=nil { return e }
		}
		
		buf,e := rd.ReadSlice(':')
		if e!=nil { return e }
		nk := r.cbbin(stripTail(buf,':'))
		lowerize(nk)
		sum := md5.Sum(nk)
		buf,e = rd.ReadSlice('\n')
		if e!=nil { return e }
		
		nv := r.cbbin(stripHead(stripTail(buf,'\n'),' '))
		r.headKV = append(r.headKV,[2][]byte{nk,nv})
		r.headIdx.init()[sum] = len(r.headKV)-1
	}
	return nil
}

func (r *Header) readBody(rd *bufio.Reader) error {
	buf := r.GetHeaderK("content-length")
	i := 0
	for _,b := range buf {
		if b<'0' || '9'<b { continue }
		i = (i*10)+int(b-'0')
	}
	if i==0 {
		if r.body!=nil { r.body.Reset() }
		return nil
	}
	body := r.Body()
	if cap(body.B) < i {
		body.B = make([]byte,i)
	} else {
		body.B = body.B[:i]
	}
	_,e := io.ReadFull(rd, body.B)
	return e
}

func (r *Header) read(rd *bufio.Reader) error {
	e := r.readHead(rd)
	if e!=nil { return e }
	return r.readBody(rd)
}

func (r *Header) write(wd *bufio.Writer) error {
	i := 0
	j := 0
	{
		buf := r.GetHeader(contentLength)
		for _,b := range buf {
			if b<'0' || '9'<b { continue }
			i = (i*10)+int(b-'0')
		}
	}
	if r.body!=nil {
		j = r.body.Len()
	}
	if i!=j {
		r.SetHeader(contentLength,encodeInt(r.tempbuf[:],j))
	}
	
	for _,kv := range r.headKV {
		if _,e := wd.Write(kv[0]) ; e!=nil { return e }
		if _,e := wd.WriteString(": ") ; e!=nil { return e }
		if _,e := wd.Write(kv[1]) ; e!=nil { return e }
		if e := wd.WriteByte('\n') ; e!=nil { return e }
	}
	wd.WriteByte('\n')
	
	if r.body==nil { return nil }
	
	_,e := r.body.WriteTo(wd)
	return e
}


