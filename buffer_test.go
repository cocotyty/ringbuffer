package ringbuffer

import (
	"runtime"
	"testing"
)

const Parallelism = 200
const BufferSize = 65536

func BenchmarkConsume(b *testing.B) {
	rb := NewBySize(BufferSize)
	b.SetParallelism(Parallelism)
	total := 0
	go func() {
		for {
			elems := rb.Consume()
			if len(elems) == 0 {
				runtime.Gosched()
				continue
			}
			for _, elem := range elems {
				total += elem.Req.(int)
				elem.Response = total
			}
			rb.Commit(elems)
		}
	}()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rb.SubmitAndWait(1)
		}
	})
}

func BenchmarkPConsume(b *testing.B) {
	rb := NewBySize(BufferSize)
	total := 0
	for i := 0; i < runtime.NumCPU()*Parallelism; i++ {
		go func() {
			for {
				rb.SubmitAndWait(1)
				runtime.Gosched()
			}
		}()
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			elm := rb.SafeConsume()
			total += elm.Req.(int)
			elm.Response = total
			elm.WakeUp()
		}
	})
}

func BenchmarkChannel(b *testing.B) {
	type reqAndResp struct {
		req  interface{}
		resp interface{}
		ch   chan struct{}
	}
	ch := make(chan reqAndResp, BufferSize)
	b.SetParallelism(Parallelism)
	total := 0
	go func() {
		for rr := range ch {
			total += rr.req.(int)
			rr.resp = total
			rr.ch <- struct{}{}
		}
	}()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mc := make(chan struct{})
			ch <- reqAndResp{req: 1, ch: mc}
			<-mc
		}
	})
}

func BenchmarkPChannel(b *testing.B) {
	type reqAndResp struct {
		req  interface{}
		resp interface{}
		ch   chan struct{}
	}
	ch := make(chan reqAndResp, BufferSize)
	total := 0

	for i := 0; i < runtime.NumCPU()*Parallelism; i++ {
		go func() {
			for {
				mc := make(chan struct{})
				ch <- reqAndResp{req: 1, ch: mc}
				<-mc
			}
		}()
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rr := <-ch
			total += rr.req.(int)
			rr.resp = total
			rr.ch <- struct{}{}
		}
	})
}
