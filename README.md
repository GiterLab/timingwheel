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
BenchmarkStandardTimer_StartStop/N-5m-4       	 2526444	       12165 ns/op	     855 B/op	       2 allocs/op
BenchmarkStandardTimer_StartStop/N-10m-4      	 1497739	       737.2 ns/op	      80 B/op	       1 allocs/op
PASS
ok  	github.com/GiterLab/timingwheel	256.826s
```

```
MacBook Pro (16-inch, 2019), 2.6 GHz 六核Intel Core i7, 16 GB 2667 MHz DDR4

$ go test -bench=. -benchmem
goos: darwin
goarch: amd64
pkg: github.com/GiterLab/timingwheel
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
BenchmarkTimingWheel_StartStop/N-1m-12 	         4592186	       259.3 ns/op	      84 B/op	       2 allocs/op
BenchmarkTimingWheel_StartStop/N-5m-12 	         4344758	       278.9 ns/op	     105 B/op	       2 allocs/op
BenchmarkTimingWheel_StartStop/N-10m-12          4952068	       272.2 ns/op	     141 B/op	       2 allocs/op
BenchmarkStandardTimer_StartStop/N-1m-12         6451620	       191.3 ns/op	      82 B/op	       1 allocs/op
BenchmarkStandardTimer_StartStop/N-5m-12         5815448	       191.9 ns/op	      85 B/op	       1 allocs/op
BenchmarkStandardTimer_StartStop/N-10m-12        5460996	       285.6 ns/op	      85 B/op	       1 allocs/op
PASS
ok  	github.com/GiterLab/timingwheel	59.867s

$ go test -bench=. -benchmem
goos: darwin
goarch: amd64
pkg: github.com/GiterLab/timingwheel
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
BenchmarkTimingWheel_StartStop/N-1m-12 	         4784413	       242.3 ns/op	      84 B/op	       2 allocs/op
BenchmarkTimingWheel_StartStop/N-5m-12 	         4522057	       269.3 ns/op	     104 B/op	       2 allocs/op
BenchmarkTimingWheel_StartStop/N-10m-12          4596720	       341.2 ns/op	      45 B/op	       1 allocs/op
BenchmarkStandardTimer_StartStop/N-1m-12         6490748	       172.0 ns/op	      80 B/op	       1 allocs/op
BenchmarkStandardTimer_StartStop/N-5m-12         7075614	       193.2 ns/op	      81 B/op	       1 allocs/op
BenchmarkStandardTimer_StartStop/N-10m-12        4448599	       275.0 ns/op	      82 B/op	       1 allocs/op
PASS
ok  	github.com/GiterLab/timingwheel	57.599s
```

## License

[MIT][5]

[1]: https://www.confluent.io/blog/apache-kafka-purgatory-hierarchical-timing-wheels/
[2]: http://www.cs.columbia.edu/~nahum/w6998/papers/ton97-timing-wheels.pdf
[3]: http://russellluo.com/2018/10/golang-implementation-of-hierarchical-timing-wheels.html
[4]: https://godoc.org/github.com/GiterLab/timingwheel
[5]: http://opensource.org/licenses/MIT
