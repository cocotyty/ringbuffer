High Performance RingBuffer
---
Lock-free and sync-submitting RingBuffer.

无锁化，同步提交的，多核芯友好的高性能环形队列。

## Example

### M producers and one consumer

```go
package main
import "runtime"
import "github.com/cocotyty/ringbuffer"

func main() {
	rb := ringbuffer.New(0)
    go func() {
    	// xxx
    	job:= 1 
    	resp:=rb.SubmitAndWait(job)
    	// xxx
    }()
	for {
		elems := rb.Consume()
		if len(elems) == 0 {
			runtime.Gosched()
			continue
		}
		// do jobs
		handle(elems)
		// resume waiting goroutines
		rb.Commit(elems)
    }
}
```

## Performance
```
oos: linux
goarch: amd64
pkg: github.com/cocotyty/ringbuffer
cpu: Intel(R) Xeon(R) Gold 6240 CPU @ 2.60GHz
BenchmarkConsume
BenchmarkConsume-36     	 2507324	       503.2 ns/op
BenchmarkPConsume
BenchmarkPConsume-36    	 1000000	      1060 ns/op
BenchmarkChannel
BenchmarkChannel-36     	 1019973	      1320 ns/op
BenchmarkPChannel
BenchmarkPChannel-36    	  699348	      2011 ns/op
PASS
ok  	github.com/cocotyty/ringbuffer	7.720s
```