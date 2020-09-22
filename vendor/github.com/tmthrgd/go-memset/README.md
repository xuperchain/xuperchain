# go-memset

[![GoDoc](https://godoc.org/github.com/tmthrgd/go-memset?status.svg)](https://godoc.org/github.com/tmthrgd/go-memset)
[![Build Status](https://travis-ci.org/tmthrgd/go-memset.svg?branch=master)](https://travis-ci.org/tmthrgd/go-memset)

An efficient memset implementation for Golang.

In Golang the following loop is optimised with an assembly implementation in src/runtime/memclr_$GOARCH.s
(since [137880043](https://golang.org/cl/137880043)):
```
for i := range data {
	data[i] = 0
}
```
but the following loop is *not* optimised:
```
for i := range data {
	data[i] = 0xff
}
```
and neither is:
```
for i := range data {
	data[i] = value
}
```

go-memset provides a Memset function which uses an assembly implementation on x86-64 and can provide
performance equivalent to the optimised first loop.

## Download

```
go get github.com/tmthrgd/go-memset
```

## Benchmark

```
BenchmarkMemset/32-8  	200000000	         6.32 ns/op	5060.69 MB/s
BenchmarkMemset/128-8 	200000000	         6.55 ns/op	19527.77 MB/s
BenchmarkMemset/1k-8  	50000000	        22.9 ns/op	44788.18 MB/s
BenchmarkMemset/16k-8 	 5000000	       278 ns/op	58868.04 MB/s
BenchmarkMemset/128k-8	  500000	      3726 ns/op	35171.08 MB/s
BenchmarkMemset/1M-8  	   30000	     40219 ns/op	26071.10 MB/s
BenchmarkMemset/16M-8 	    2000	   1120266 ns/op	14976.09 MB/s
BenchmarkMemset/128M-8	     200	   8749141 ns/op	15340.67 MB/s
BenchmarkMemset/512M-8	      50	  35078079 ns/op	15305.03 MB/s
```

```
BenchmarkGoZero/32-8  	500000000	         3.43 ns/op	9326.33 MB/s
BenchmarkGoZero/128-8 	300000000	         4.50 ns/op	28414.80 MB/s
BenchmarkGoZero/1k-8  	100000000	        19.1 ns/op	53557.45 MB/s
BenchmarkGoZero/16k-8 	 5000000	       278 ns/op	58854.20 MB/s
BenchmarkGoZero/128k-8	  500000	      3733 ns/op	35102.85 MB/s
BenchmarkGoZero/1M-8  	   50000	     39968 ns/op	26234.86 MB/s
BenchmarkGoZero/16M-8 	    1000	   1319397 ns/op	12715.81 MB/s
BenchmarkGoZero/128M-8	     100	  10682865 ns/op	12563.83 MB/s
BenchmarkGoZero/512M-8	      30	  42689135 ns/op	12576.29 MB/s
BenchmarkGoSet/32-8   	100000000	        17.4 ns/op	1840.05 MB/s
BenchmarkGoSet/128-8  	20000000	        73.0 ns/op	1754.20 MB/s
BenchmarkGoSet/1k-8   	 3000000	       545 ns/op	1878.82 MB/s
BenchmarkGoSet/16k-8  	  200000	      8638 ns/op	1896.63 MB/s
BenchmarkGoSet/128k-8 	   20000	     69077 ns/op	1897.47 MB/s
BenchmarkGoSet/1M-8   	    3000	    552612 ns/op	1897.49 MB/s
BenchmarkGoSet/16M-8  	     200	   8867019 ns/op	1892.09 MB/s
BenchmarkGoSet/128M-8 	      20	  70937303 ns/op	1892.06 MB/s
BenchmarkGoSet/512M-8 	       5	 283412563 ns/op	1894.31 MB/s
```

```
benchmark                old ns/op     new ns/op     delta
BenchmarkZero/32-8       3.43          6.32          +84.26%
BenchmarkZero/128-8      4.50          6.55          +45.56%
BenchmarkZero/1k-8       19.1          22.9          +19.90%
BenchmarkZero/16k-8      278           278           +0.00%
BenchmarkZero/128k-8     3733          3726          -0.19%
BenchmarkZero/1M-8       39968         40219         +0.63%
BenchmarkZero/16M-8      1319397       1120266       -15.09%
BenchmarkZero/128M-8     10682865      8749141       -18.10%
BenchmarkZero/512M-8     42689135      35078079      -17.83%
BenchmarkSet/32-8        17.4          6.32          -63.68%
BenchmarkSet/128-8       73.0          6.55          -91.03%
BenchmarkSet/1k-8        545           22.9          -95.80%
BenchmarkSet/16k-8       8638          278           -96.78%
BenchmarkSet/128k-8      69077         3726          -94.61%
BenchmarkSet/1M-8        552612        40219         -92.72%
BenchmarkSet/16M-8       8867019       1120266       -87.37%
BenchmarkSet/128M-8      70937303      8749141       -87.67%
BenchmarkSet/512M-8      283412563     35078079      -87.62%

benchmark                old MB/s     new MB/s     speedup
BenchmarkZero/32-8       9326.33      5060.69      0.54x
BenchmarkZero/128-8      28414.80     19527.77     0.69x
BenchmarkZero/1k-8       53557.45     44788.18     0.84x
BenchmarkZero/16k-8      58854.20     58868.04     1.00x
BenchmarkZero/128k-8     35102.85     35171.08     1.00x
BenchmarkZero/1M-8       26234.86     26071.10     0.99x
BenchmarkZero/16M-8      12715.81     14976.09     1.18x
BenchmarkZero/128M-8     12563.83     15340.67     1.22x
BenchmarkZero/512M-8     12576.29     15305.03     1.22x
BenchmarkSet/32-8        1840.05      5060.69      2.75x
BenchmarkSet/128-8       1754.20      19527.77     11.13x
BenchmarkSet/1k-8        1878.82      44788.18     23.84x
BenchmarkSet/16k-8       1896.63      58868.04     31.04x
BenchmarkSet/128k-8      1897.47      35171.08     18.54x
BenchmarkSet/1M-8        1897.49      26071.10     13.74x
BenchmarkSet/16M-8       1892.09      14976.09     7.92x
BenchmarkSet/128M-8      1892.06      15340.67     8.11x
BenchmarkSet/512M-8      1894.31      15305.03     8.08x
```

## License

Unless otherwise noted, the go-memset source files are distributed under the Modified BSD License
found in the LICENSE file.
