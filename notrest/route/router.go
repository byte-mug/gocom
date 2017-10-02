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

import "github.com/byte-mug/gocom/notrest"
import "strings"
import "fmt"

type Router struct{
	rawElems []listElement
	tree     map[string]*treeElement
}
func (r *Router) MethodRaw(method string,prefix string,isAbsolute bool,handler Handler) {
	r.rawElems = append(r.rawElems,listElement{ method,prefix,isAbsolute,handler } )
}
func (r *Router) Prepare() {
	r.tree = genTree2(r.rawElems)
}
func (r *Router) Handle(req *notrest.Request, resp *notrest.Response) {
	if r.tree==nil { panic("BUG: You must call Prepare before use") }
	te,ok := r.tree[string(req.Method())]
	if !ok {
		resp.Status(404)
		return
	}
	pth,te := te.perform(req.Path())
	
	if len(pth)==0 && te.absol!=nil {
		te.absol.handler(req,resp,pth)
		return
	} else if te.relat!=nil {
		te.relat.handler(req,resp,pth)
		return
	}
	resp.Status(404)
}

func (r *Router) Method(method string,path string,handler Handler) {
	isAbso := strings.HasSuffix(path,"*")
	path = strings.TrimSuffix(path,"*")
	r.MethodRaw(method,path,isAbso,handler)
}
func (r *Router) GET(path string,handler Handler) { r.Method("GET",path,handler) }
func (r *Router) POST(path string,handler Handler) { r.Method("POST",path,handler) }
func (r *Router) PUT(path string,handler Handler) { r.Method("PUT",path,handler) }
func (r *Router) DELETE(path string,handler Handler) { r.Method("DELETE",path,handler) }
func (r *Router) String() string {
	return fmt.Sprint(r.tree)
}
