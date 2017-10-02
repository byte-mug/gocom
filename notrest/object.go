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

import "github.com/valyala/bytebufferpool"

var bufpool bytebufferpool.Pool

const cDown = 'a'-'A'

func lowerize(bs []byte) {
	for i,b := range bs {
		if !(b<'A' || b>'Z') { bs[i] = b+cDown }
	}
}

func uperize(bs []byte) {
	for i,b := range bs {
		if !(b<'a' || b>'z') { bs[i] = b-cDown }
	}
}

func stripTail(buf []byte,post byte) []byte {
	if len(buf)==0 { return nil }
	i := len(buf)
	for i-- ; 0<i ; i-- {
		if buf[i]!=post { return buf[:i+1] }
	}
	return nil
}
func stripHead(buf []byte,post byte) []byte {
	for i,b := range buf {
		if b!=post { return buf[i:] }
	}
	return nil
}
func encodeInt(storage []byte,val int) []byte {
	i := len(storage)
	for i>0 {
		i--
		storage[i] = byte( (val%10) + '0' )
		val /= 10
		if val==0 { break }
	}
	return storage[i:]
}


type bufferAlloc struct{
	basis  []byte
	
	offset int
}
func (b *bufferAlloc) reset() { b.offset = 0 }
func (b *bufferAlloc) alloc(n int) []byte {
	lbb := len(b.basis)
	if lbb==0 {
		lbb = 1<<16
		b.basis = make([]byte,1<<16)
	}
	
	ending := b.offset + n
	if ending > lbb {
		if lbb > (1<<24) {
			return nil
		}
		b.basis = make([]byte,lbb<<1)
	}
	begin := b.offset
	b.offset = ending
	return b.basis[begin:b.offset]
}

type indexMap map[[16]byte]int
func (i *indexMap) init() indexMap {
	if (*i)!=nil { return *i }
	*i = make(indexMap)
	return *i
}
func (i indexMap) clear() {
	if i==nil { return }
	for id := range i { delete(i,id) }
}
