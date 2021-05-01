
# Debugging performance issues and profiling Go applications (CPU, IO, memory) hotspots

https://github.com/golang/go/wiki/Performance  
**Use tools in isolation to get more precise info.**  

Go runtime contains built-in CPU profiler, which shows what functions consume what percent of CPU time. There are 3 ways you can get access to it:
1. this option works only for tests/benchmarks:
```sh
go test -run=none -bench=ClientServerParallel4 -cpuprofile=cpu.out net/http
# or profile an iside package (note: only use one profile at a time e.g. cpuprofile):
go test -bench=. -benchmem -memprofile mem.out -cpuprofile cpu.out -blockprofile block.out -mutexprofile mutex.out ./internal/services/
```
will profile the given benchmark and write CPU profile into 'cprof' file. Then:

```sh
go tool pprof -text -output cpu.txt cpu.out
go tool pprof -png -output cpu.png cpu.out
go tool pprof -png -output mem.png mem.out
go tool pprof -png -output block.png block.out
go tool pprof -png -output mutex.png mutex.out
```
will print a list of the hottest functions.

2. [net/http/pprof](http://golang.org/pkg/net/http/pprof): This is the ideal solution for network servers:  
Add an import:
```go
import _ "net/http/pprof"
```
If your application is not already running an http server, you need to start one. Add "net/http" and "log" to your imports and the following code to your main function (change the port number if needed):
```go
	go func() {
		log.Println(http.ListenAndServe("localhost:8787", nil))
	}()
```

Get the results of CPU profiling:
```sh
curl -o cpu.prof "http://myserver:8787/debug/pprof/profile"
```

```sh
go tool pprof --text mybin  http://myserver:8787:/debug/pprof/profile
```
3. Manual profile collection. You need to import [runtime/pprof](http://golang.org/pkg/runtime/pprof/) and add the following code to main function:
```go
if *flagCpuprofile != "" {
    f, err := os.Create(*flagCpuprofile)
    if err != nil {
        log.Fatal(err)
    }
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()
}
```
The profile will be written to the specified file, visualize it the same way as in the first option.

---

## CPU Profiler
There are also 3 special entries that the profiler uses when it can't unwind stack: GC, System and ExternalCode. GC means time spent during garbage collection. System means time spent in goroutine scheduler, stack management code and other auxiliary runtime code. ExternalCode means time spent in native dynamic libraries.  
If you see lots of time spent in runtime.mallocgc function, the program potentially makes excessive amount of small memory allocations. The profile will tell you where the allocations are coming from.  

If lots of time is spent in channel operations, sync.Mutex code and other synchronization primitives or System component, the program probably suffers from contention. Consider to restructure program to eliminate frequently accessed shared resources. Common techniques for this include sharding/partitioning, local buffering/batching and copy-on-write technique.  

If lots of time is spent in syscall.Read/Write, the program potentially makes excessive amount of small reads and writes. Bufio wrappers around os.File or net.Conn can help in this case.

If lots of time is spent in GC component, the program either allocates too many transient objects or heap size is very small so garbage collections happen too frequently.  

## Memory Profiler
Memory profiler shows what functions allocate heap memory. You can collect it in similar ways as CPU profile: with go test --memprofile (http://golang.org/cmd/go/#hdr-Description_of_testing_flags), with net/http/pprof via http://myserver:6060:/debug/pprof/heap or by calling runtime/pprof.WriteHeapProfile.

You can visualize only allocations live at the time of profile collection (--inuse_space flag to pprof, default), or all allocations happened since program start (--alloc_space flag to pprof). The former is useful for profiles collected with net/http/pprof on live applications, the latter is useful for profiles collected at program end (otherwise you will see almost empty profile).

Note: the memory profiler is sampling, that is, it collects information only about some subset of memory allocations. Probability of sampling an object is proportional to its size. You can change the sampling rate with go test --memprofilerate flag, or by setting runtime.MemProfileRate variable at program startup. The rate of 1 will lead to collection of information about all allocations, but it can slow down execution. The default sampling rate is 1 sample per 512KB of allocated memory.

You can also visualize number of bytes allocated or number of objects allocated (--inuse/alloc_space and --inuse/alloc_objects flags, respectively). The profiler tends to sample larger objects during profiling more. But it's important to understand that large objects affect memory consumption and GC time, while large number of tiny allocations affects execution speed (and GC time to some degree as well). So it may be useful to look at both.

Objects can be persistent or transient. If you have several large persistent objects allocated at program start, they will be most likely sampled by the profiler (because they are large). Such objects do affect memory consumption and GC time, but they do not affect normal execution speed (no memory management operations happen on them). On the other hand if you have large number of objects with very short life durations, they can be barely represented in the profile (if you use the default --inuse_space mode). But they do significantly affect execution speed, because they are constantly allocated and freed. So, once again, it may be useful to look at both types of objects. So, generally, if you want to reduce memory consumption, you need to look at --inuse_space profile collected during normal program operation. If you want to improve execution speed, look at --alloc_objects profile collected after significant running time or at program end.

There are several flags that control reporting granularity. --functions makes pprof report on function level (default). --lines makes pprof report on source line level, which is useful if hot functions allocate on different lines. And there are also --addresses and --files for exact instruction address and file level, respectively.

There is a useful option for the memory profile -- you can look at it right in the browser (provided that you imported net/http/pprof).


---

https://blog.golang.org/pprof

https://medium.com/@tvii/continuous-profiling-and-go-6c0ab4d2504b


[Cloud Profiler](https://cloud.google.com/profiler/):  
Low-impact production profiling
While it’s possible to measure code performance in development environments, the results generally don’t map well to what’s happening in production. Many production profiling techniques either slow down code execution or can only inspect a small subset of a codebase. Cloud Profiler uses statistical techniques and extremely low-impact instrumentation that runs across all production application instances to provide a complete picture of an application’s performance without slowing it down.


[Google-Wide Profiling: A Continuous Profiling Infrastructure for Data Centers](https://research.google/pubs/pub36575/)

---


```sh
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative hello/hello.proto

go mod tidy

go run .

go test -v

go test -bench=. -benchmem -memprofile mem.out -cpuprofile cpu.out -blockprofile block.out -mutexprofile mutex.out

go tool pprof -http=":8787" cpu.out

go tool pprof -text -output cpu.txt cpu.out
go tool pprof -png -output cpu.png cpu.out
go tool pprof -png -output mem.png mem.out

go tool pprof -png -output block.png block.out
go tool pprof -png -output mutex.png mutex.out
```