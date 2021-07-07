package ringbuffer

import (
	"github.com/cocotyty/oneshot"
	"runtime"
	"sync/atomic"
)

type Element struct {
	oneshot.Shot
	Req      interface{}
	Response interface{}
}

type RingBuffer struct {
	pointer
	mask      uint32
	ring      []*Element
	committed []uint64
}

// New create a ring buffer.
// Buffer size is 2 to the power exp.
// If exp is zero , buffer size will be round(cpu * 32)
func New(exp uint32) *RingBuffer {
	size := uint32(2 << exp)
	if exp == 0 {
		size = round(uint32(runtime.NumCPU() * 32))
	}
	rb := &RingBuffer{ring: make([]*Element, size), committed: make([]uint64, size)}
	rb.committed[0] -= 1
	rb.mask = size - 1
	return rb
}

func round(v uint32) uint32 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return v
}

type pointer struct {
	head uint32
	tail uint32
}

func (r *RingBuffer) SubmitAndWait(request interface{}) (response interface{}) {
	tail := atomic.AddUint32(&r.tail, 1)
	for tail-atomic.LoadUint32(&r.head) >= uint32(len(r.ring)) {
		runtime.Gosched()
	}
	idx := tail - 1
	elm := &Element{
		Req:      request,
		Response: nil,
	}
	r.ring[idx&r.mask] = elm
	atomic.StoreUint64(&r.committed[idx&r.mask], uint64(idx))
	elm.Wait()
	return elm.Response
}

func (r *RingBuffer) SafeConsume() (elem *Element) {
	for {
		head := atomic.LoadUint32(&r.head)
		idx := head
		if atomic.LoadUint64(&r.committed[idx&r.mask]) != uint64(idx) {
			runtime.Gosched()
			continue
		}
		if !atomic.CompareAndSwapUint32(&r.head, head, head+1) {
			runtime.Gosched()
			continue
		}
		elem = r.ring[idx&r.mask]
		break
	}
	return
}

func (r *RingBuffer) Consume() (elems []*Element) {
	size := uint32(0)
	head := r.head
	idx := head
	for {
		if atomic.LoadUint64(&r.committed[idx&r.mask]) != uint64(idx) {
			break
		}
		idx++
		size++
	}
	elems = make([]*Element, size)
	for i := uint32(0); i < size; i++ {
		elems[i] = r.ring[(head+i)&r.mask]
	}
	atomic.StoreUint32(&r.head, head+size)
	return elems
}

func (r *RingBuffer) Commit(elems []*Element) {
	for _, elem := range elems {
		elem.WakeUp()
	}
}
