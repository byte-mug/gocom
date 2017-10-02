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


package route

import "fmt"
import "sort"
import "github.com/byte-mug/gocom/notrest"

type Handler func(req *notrest.Request, resp *notrest.Response, rest []byte)

func min(a, b int) int {
	if a<b { return a }
	return b
}

type treeElement struct{
	relat  *listElement
	absol  *listElement
	sub    map[string]*treeElement
	prefix int
}

func (te* treeElement) perform(path []byte) ([]byte,*treeElement) {
	
	for {
		lp := min(len(path),te.prefix)
		ste,sok := te.sub[string(path[:lp])]
		if !sok {
			return path[lp:],te
		}
		te = ste
		path = path[lp:]
	}
	
	panic("...")
}
func (te* treeElement) String() string {
	return fmt.Sprintf("{[%d]%v}",te.prefix,te.sub)
}


type listElement struct{
	method  string
	prefix  string
	isAbsol bool
	handler Handler
}
func (a listElement) ptr() *listElement {
	p := new(listElement)
	*p = a
	return p
}

func (a listElement) less(b listElement) bool {
	if a.isAbsol && !b.isAbsol { return true }
	if !a.isAbsol && b.isAbsol { return false }
	
	return len(a.prefix) < len(b.prefix)
}
type listElementList []listElement
func (a listElementList) Len() int { return len(a) }
func (a listElementList) Less(i, j int) bool { return a[i].less(a[j]) }
func (a listElementList) Swap(i, j int) { a[i],a[j] = a[j],a[i] }

func genTree2(elems []listElement) map[string]*treeElement {
	lem := make(map[string][]listElement)
	for _,le := range elems {
		me := le.method
		lem[me] = append(lem[me],le)
	}
	m := make(map[string]*treeElement,len(lem))
	for k,v := range lem {
		m[k] = genTree1(v)
	}
	return m
}

func genTree1(elems []listElement) *treeElement {
	sort.Stable(listElementList(elems))
	return convert(elems)
}
func convert(elems []listElement) *treeElement {
	te := &treeElement{
		sub: make(map[string]*treeElement),
	}
	lem := make(map[string][]listElement)
	
	n := len(elems)
	pl := 0
	for i,le := range elems {
		if le.prefix=="" {
			te.relat = le.ptr()
		} else { pl = len(le.prefix) ; n = i ; break }
	}
	te.prefix = pl
	elems = elems[n:]
	
	n = len(elems)
	for i,le := range elems {
		if le.isAbsol { n = i ; break }
		
		str := le.prefix[:pl]
		le.prefix = le.prefix[pl:]
		
		lem[str] = append(lem[str],le)
	}
	elems = elems[n:]
	
	n = len(elems)
	for i,le := range elems {
		if le.prefix=="" {
			te.absol = le.ptr()
		} else { n = i ; break }
	}
	elems = elems[n:]
	
	for _,le := range elems {
		pre := le.prefix
		rest := ""
		if len(pre)>pl {
			rest = pre[pl:]
			pre = pre[:pl]
		}
		le.prefix = rest
		lem[pre] = append(lem[pre],le)
	}
	
	for k,v := range lem {
		if len(v)==0 { continue }
		te.sub[k] = convert(v)
	}
	
	return te
}


