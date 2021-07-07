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
48 core cpu.
```
goos: linux
goarch: amd64
pkg: github.com/cocotyty/ringbuffer
cpu: Intel(R) Xeon(R) Gold 6271C CPU @ 2.60GHz
BenchmarkConsume
BenchmarkConsume-48     	 2074874	       557.1 ns/op
BenchmarkPConsume
BenchmarkPConsume-48    	 1000000	      1138 ns/op
BenchmarkChannel
BenchmarkChannel-48     	  975303	      1232 ns/op
BenchmarkPChannel
BenchmarkPChannel-48    	  780142	      1659 ns/op
PASS
ok  	github.com/cocotyty/ringbuffer	6.889s
```