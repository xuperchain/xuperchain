# go-popcount

[![GoDoc](https://godoc.org/github.com/tmthrgd/go-popcount?status.svg)](https://godoc.org/github.com/tmthrgd/go-popcount)
[![Build Status](https://travis-ci.org/tmthrgd/go-popcount.svg?branch=master)](https://travis-ci.org/tmthrgd/go-popcount)

A population count implementation for Golang.

An x86-64 implementation is provided that uses the POPCNT instruction.

## Download

```
go get github.com/tmthrgd/go-popcount
```

## Benchmark

The following benchmarks relate to the assembly implementation on an
AMD64 CPU with POPCOUNT support.
```
BenchmarkCountBytes/32-8        300000000                5.47 ns/op     5854.04 MB/s
BenchmarkCountBytes/128-8       100000000               10.2 ns/op      12609.36 MB/s
BenchmarkCountBytes/1K-8        30000000                49.2 ns/op      20804.48 MB/s
BenchmarkCountBytes/16K-8        3000000               572 ns/op        28642.76 MB/s
BenchmarkCountBytes/128K-8        300000              4948 ns/op        26486.47 MB/s
BenchmarkCountBytes/1M-8           30000             50728 ns/op        20670.19 MB/s
BenchmarkCountBytes/16M-8           1000           1412299 ns/op        11879.36 MB/s
BenchmarkCountBytes/128M-8           100          11388799 ns/op        11785.06 MB/s
BenchmarkCountBytes/512M-8            30          45068056 ns/op        11912.45 MB/s
BenchmarkCountSlice64/32-8      300000000                5.51 ns/op     5804.87 MB/s
BenchmarkCountSlice64/128-8     100000000               10.6 ns/op      12047.47 MB/s
BenchmarkCountSlice64/1K-8      20000000                51.4 ns/op      19932.68 MB/s
BenchmarkCountSlice64/16K-8      2000000               597 ns/op        27414.74 MB/s
BenchmarkCountSlice64/128k-8  	  300000              4960 ns/op        26425.47 MB/s
BenchmarkCountSlice64/1M-8    	   30000             50861 ns/op        20616.24 MB/s
BenchmarkCountSlice64/16M-8   	    1000           1419479 ns/op        11819.28 MB/s
BenchmarkCountSlice64/128M-8  	     100          11287323 ns/op        11891.01 MB/s
BenchmarkCountSlice64/512M-8  	      30          45210522 ns/op        11874.91 MB/s
```

The following benchmarks relate to the Golang implementation using
[math/bits.OnesCount64](https://golang.org/pkg/math/bits/#OnesCount64) on
an AMD64 CPU with POPCOUNT support.
```
BenchmarkCountBytesGo/32-8    	100000000               11.1 ns/op      2883.25 MB/s
BenchmarkCountBytesGo/128-8    100000000               20.6 ns/op       6204.80 MB/s
BenchmarkCountBytesGo/1k-8    	10000000               115 ns/op        8896.25 MB/s
BenchmarkCountBytesGo/16k-8   	 1000000              1640 ns/op        9986.94 MB/s
BenchmarkCountBytesGo/128k-8  	  100000             13017 ns/op        10068.65 MB/s
BenchmarkCountBytesGo/1M-8    	   10000            105315 ns/op        9956.50 MB/s
BenchmarkCountBytesGo/16M-8   	    1000           2140396 ns/op        7838.37 MB/s
BenchmarkCountBytesGo/128M-8  	     100          17149248 ns/op        7826.45 MB/s
BenchmarkCountBytesGo/512M-8  	      20          68345879 ns/op        7855.21 MB/s
BenchmarkCountSlice64Go/32-8  	200000000                6.61 ns/op     4840.05 MB/s
BenchmarkCountSlice64Go/128-8 	100000000               16.1 ns/op      7936.33 MB/s
BenchmarkCountSlice64Go/1k-8  	20000000               111 ns/op        9184.79 MB/s
BenchmarkCountSlice64Go/16k-8 	 1000000              1636 ns/op        10012.94 MB/s
BenchmarkCountSlice64Go/128k-8	  100000             13053 ns/op        10041.31 MB/s
BenchmarkCountSlice64Go/1M-8  	   10000            105796 ns/op        9911.24 MB/s
BenchmarkCountSlice64Go/16M-8 	    1000           2145359 ns/op        7820.24 MB/s
BenchmarkCountSlice64Go/128M-8	     100          17232666 ns/op        7788.56 MB/s
BenchmarkCountSlice64Go/512M-8	      20          68713386 ns/op        7813.19 MB/s
```

## License

Unless otherwise noted, the go-popcount source files are distributed under the Modified BSD License
found in the LICENSE file.
