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

type ServerError int
const (
	SE_None     ServerError = iota
	SE_QueueOverflow
)
func (s ServerError) Error() string {
	switch s {
	case SE_None: return "No Error!"
	case SE_QueueOverflow: return "Queue Overflow"
	}
	return "???"
}

type ServerHandler func(req Request,resp Response)
type ErrorHandler  func(e ServerError,req Request,resp Response)

type ServerCodec interface {
	Close() error
	
	// Threadsafe
	Send(id uint64,r Response) error
	
	// Not threadsafe
	RecvId() (uint64,error)
	Recv(r Request) error
}

type Server struct {
	Handle      ServerHandler
	Error       ErrorHandler
	Allocator   func() (req Request,resp Response)
	Concurrency int
	QueueMulti  int
	
	initialize  sync.Once
	queue       chan *serverWorkItem
	pool        sync.Pool
}
func (s *Server) initi() {
	conc := s.Concurrency
	qm := s.QueueMulti
	if conc<1 { conc=16 }
	if qm<1 { qm =16 }
	s.queue = make(chan *serverWorkItem,conc*qm)
	for i:=0 ; i<conc ; i++ {
		go s.worker()
	}
}
func (s *Server) worker() {
	for swi := range s.queue {
		swi.handle()
		swi.server.pool.Put(swi)
	}
}
func (s *Server) alloc() *serverWorkItem {
	rswi := s.pool.Get()
	if rswi!=nil { return rswi.(*serverWorkItem) }
	
	swi := new(serverWorkItem)
	swi.req,swi.resp = s.Allocator()
	swi.server = s
	return swi
}

type serverWorkItem struct {
	id     uint64
	req    Request
	resp   Response
	codec  ServerCodec
	server *Server
}
func (s *serverWorkItem) handle() {
	s.server.Handle(s.req,s.resp)
	s.codec.Send(s.id,s.resp)
}
func (s *serverWorkItem) doerror() {
	s.server.Error(SE_QueueOverflow,s.req,s.resp)
	s.codec.Send(s.id,s.resp)
}

func (s *Server) input(codec ServerCodec) error {
	for {
		id,err := codec.RecvId()
		if err!=nil { return err }
		swi := s.alloc()
		swi.id    = id
		swi.codec = codec
		err = codec.Recv(swi.req)
		if err!=nil { return err }
		select {
		case s.queue <- swi:
			// OK!
		default:
			swi.doerror()
			s.pool.Put(swi)
		}
	}
	panic("...")
}

func (s *Server) Serve(sc ServerCodec) error {
	s.initialize.Do(s.initi)
	return s.input(sc)
}


