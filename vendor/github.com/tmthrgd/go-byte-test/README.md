# go-byte-test

[![GoDoc](https://godoc.org/github.com/tmthrgd/go-byte-test?status.svg)](https://godoc.org/github.com/tmthrgd/go-byte-test)
[![Build Status](https://travis-ci.org/tmthrgd/go-byte-test.svg?branch=master)](https://travis-ci.org/tmthrgd/go-byte-test)

An efficient byte test implementation for Golang.

It is SSE accelerated equivalent of the following function:
```
// Test returns true iff each byte in data is equal to value.
func Test(data []byte, value byte) bool {
	for _, v := range data {
		if v != value {
			return false
		}
	}

	return true
}
```

## Download

```
go get github.com/tmthrgd/go-byte-test
```

## Benchmark

```
BenchmarkTest/32-8         	200000000	         7.13 ns/op	4488.67 MB/s
BenchmarkTest/128-8        	200000000	         7.83 ns/op	16350.02 MB/s
BenchmarkTest/1K-8         	50000000	        25.6 ns/op	40050.28 MB/s
BenchmarkTest/16K-8        	 5000000	       325 ns/op	50334.53 MB/s
BenchmarkTest/128K-8       	  500000	      3021 ns/op	43373.70 MB/s
BenchmarkTest/1M-8         	   50000	     35621 ns/op	29436.64 MB/s
BenchmarkTest/16M-8        	    2000	    977792 ns/op	17158.26 MB/s
BenchmarkTest/128M-8       	     200	   7775128 ns/op	17262.44 MB/s
BenchmarkTest/512M-8       	      50	  31003763 ns/op	17316.31 MB/s
```

```
BenchmarkGoTest/32-8         	50000000	        35.8 ns/op	 893.80 MB/s
BenchmarkGoTest/128-8        	10000000	       120 ns/op	1058.03 MB/s
BenchmarkGoTest/1K-8         	 2000000	       869 ns/op	1177.11 MB/s
BenchmarkGoTest/16K-8        	  100000	     13760 ns/op	1190.62 MB/s
BenchmarkGoTest/128K-8       	   10000	    109813 ns/op	1193.59 MB/s
BenchmarkGoTest/1M-8         	    2000	    878439 ns/op	1193.68 MB/s
BenchmarkGoTest/16M-8        	     100	  14339512 ns/op	1170.00 MB/s
BenchmarkGoTest/128M-8       	      10	 114336485 ns/op	1173.88 MB/s
BenchmarkGoTest/512M-8       	       3	 457974138 ns/op	1172.27 MB/s
```

go -> go-byte-test:
```
benchmark                old ns/op     new ns/op     delta
BenchmarkTest/32-8       35.8          7.13          -80.08%
BenchmarkTest/128-8      120           7.83          -93.47%
BenchmarkTest/1K-8       869           25.6          -97.05%
BenchmarkTest/16K-8      13760         325           -97.64%
BenchmarkTest/128K-8     109813        3021          -97.25%
BenchmarkTest/1M-8       878439        35621         -95.94%
BenchmarkTest/16M-8      14339512      977792        -93.18%
BenchmarkTest/128M-8     114336485     7775128       -93.20%
BenchmarkTest/512M-8     457974138     31003763      -93.23%

benchmark                old MB/s     new MB/s     speedup
BenchmarkTest/32-8       893.80       4488.67      5.02x
BenchmarkTest/128-8      1058.03      16350.02     15.45x
BenchmarkTest/1K-8       1177.11      40050.28     34.02x
BenchmarkTest/16K-8      1190.62      50334.53     42.28x
BenchmarkTest/128K-8     1193.59      43373.70     36.34x
BenchmarkTest/1M-8       1193.68      29436.64     24.66x
BenchmarkTest/16M-8      1170.00      17158.26     14.67x
BenchmarkTest/128M-8     1173.88      17262.44     14.71x
BenchmarkTest/512M-8     1172.27      17316.31     14.77x
```

## License

Unless otherwise noted, the go-byte-test source files are distributed under the Modified BSD License
found in the LICENSE file.
