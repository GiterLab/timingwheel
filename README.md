# timingwheel

Golang implementation of Hierarchical Timing Wheels.

kafka实现: https://github.com/apache/kafka/tree/trunk/core/src/main/scala/kafka/utils/timer

## Installation

```bash
$ go get -u github.com/GiterLab/timingwheel
```


## Design

`timingwheel` is ported from Kafka's [purgatory][1], which is designed based on [Hierarchical Timing Wheels][2].

中文博客：[层级时间轮的 Golang 实现][3]。

## Documentation

For usage and examples see the [Godoc][4].

## Benchmark

    ENV: go 1.17.1


```
MacBook Pro (Retina, 13-inch, Early 2015), 2.7 GHz 双核Intel Core i5, 8 GB 1867 MHz DDR3

$ go test -bench=. -benchmem
goos: darwin
goarch: amd64
pkg: github.com/GiterLab/timingwheel
cpu: Intel(R) Core(TM) i5-5257U CPU @ 2.70GHz
BenchmarkTimingWheel_StartStop/N-1m-4         	 2707137	       406.4 ns/op	      89 B/op	       2 allocs/op
BenchmarkTimingWheel_StartStop/N-5m-4         	 3171666	       598.3 ns/op	      46 B/op	       1 allocs/op
BenchmarkTimingWheel_StartStop/N-10m-4        	 2399018	       510.7 ns/op	      51 B/op	       1 allocs/op
BenchmarkStandardTimer_StartStop/N-1m-4       	 4140207	       305.0 ns/op	      82 B/op	       1 allocs/op
BenchmarkStandardTimer_StartStop/N-5m-4       	 1253014	     25412 ns/op	    1516 B/op	       4 allocs/op
BenchmarkStandardTimer_StartStop/N-10m-4      	 2584293	     17166 ns/op	      80 B/op	       1 allocs/op
PASS
ok  	github.com/GiterLab/timingwheel	356.214s

$ go test -bench=. -benchmem
goos: darwin
goarch: amd64
pkg: github.com/GiterLab/timingwheel
cpu: Intel(R) Core(TM) i5-5257U CPU @ 2.70GHz
BenchmarkTimingWheel_StartStop/N-1m-4         	 2649186	       398.2 ns/op	      87 B/op	       2 allocs/op
BenchmarkTimingWheel_StartStop/N-5m-4         	 2382670	       673.6 ns/op	      71 B/op	       1 allocs/op
BenchmarkTimingWheel_StartStop/N-10m-4        	 2251845	       510.3 ns/op	      51 B/op	       1 allocs/op
BenchmarkStandardTimer_StartStop/N-1m-4       	 3329818	       316.2 ns/op	      83 B/op	       1 allocs/op
BenchmarkStandardTimer_StartStop/N-5m-4       	 2526444	     12165 ns/op	     855 B/op	       2 allocs/op
BenchmarkStandardTimer_StartStop/N-10m-4      	 1497739	       737.2 ns/op	      80 B/op	       1 allocs/op
PASS
ok  	github.com/GiterLab/timingwheel	256.826s
```


## License

[MIT][5]

[1]: https://www.confluent.io/blog/apache-kafka-purgatory-hierarchical-timing-wheels/
[2]: http://www.cs.columbia.edu/~nahum/w6998/papers/ton97-timing-wheels.pdf
[3]: http://russellluo.com/2018/10/golang-implementation-of-hierarchical-timing-wheels.html
[4]: https://godoc.org/github.com/GiterLab/timingwheel
[5]: http://opensource.org/licenses/MIT
