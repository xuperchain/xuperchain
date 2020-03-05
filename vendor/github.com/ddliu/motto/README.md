# motto

[![Build Status](https://travis-ci.org/ddliu/motto.png)](https://travis-ci.org/ddliu/motto)

Modular [otto](https://github.com/robertkrimen/otto)

Motto provide a Nodejs like module environment to run javascript files in golang.

## Installation

```bash
go get github.com/ddliu/motto
```

## Usage

```js
var _ = require('underscore');
var data = require('./data.json'); // [3, 2, 1, 4, 6]
module.exports = _.min(data);
```

```go
package main

import (
    "github.com/ddliu/motto"
    _ "github.com/ddliu/motto/underscore"
)

func main() {
    motto.Run("path/to/index.js")
}
```

You can also install the motto command line tool to run it directly:

```bash
go install github.com/ddliu/motto/motto
motto path/to/index.js
```

## Modules

The module environment is almost capable with Nodejs [spec](http://nodejs.org/api/modules.html).

Some Nodejs modules(without core module dependencies) can be used directly in Motto.

## Addons

Motto can be extended with addons, below is an example addon which implement part of the "fs" module of Nodejs:

```go
package fs

import (
    "github.com/ddliu/motto"
    "github.com/robertkrimen/otto"
)

func fsModuleLoader(vm *motto.Motto) (otto.Value, error) {
    fs, _ := vm.Object(`({})`)
    fs.Set("readFileSync", func(call otto.FunctionCall) otto.Value {
        filename, _ := call.Argument(0).ToString()
        bytes, err := ioutil.ReadFile(filename)
        if err != nil {
            return otto.UndefinedValue()
        }

        v, _ := call.Otto.ToValue(string(bytes))
        return v
    })

    return vm.ToValue(fs)
}

func init() {
    motto.AddModule("fs", fsModuleLoader)
}
```

After import this package, you can `require` it directly in your js file: 

```js
var fs = require('fs');
var content = fs.readFileSync('/path/to/data');
```

## Nodejs in Golang?

[nodego](https://github.com/ddliu/nodego) implements some features and core modules
of Nodejs.

## Performance

Simple benchmark shows below for furthur performance optimize:

```bash
strace -c -Ttt motto tests/index.js
strace -c -Ttt node tests/index.js
```

Motto:

```
% time     seconds  usecs/call     calls    errors syscall
------ ----------- ----------- --------- --------- ----------------
 20.20    0.000144           2        59           rt_sigaction
 15.71    0.000112           7        15           mmap
 11.92    0.000085          11         8           futex
 10.10    0.000072           6        13         4 stat
  7.43    0.000053           7         8           read
  5.89    0.000042          21         2           clone
  5.75    0.000041          10         4           open
  4.77    0.000034           7         5           rt_sigprocmask
  4.63    0.000033          33         1           execve
  3.23    0.000023          12         2           write
  2.24    0.000016           4         4           fstat
  1.82    0.000013           3         4           close
  1.82    0.000013          13         1           sched_getaffinity
  1.68    0.000012          12         1           sched_yield
  0.98    0.000007           7         1           munmap
  0.98    0.000007           7         1           arch_prctl
  0.42    0.000003           3         1           getcwd
  0.42    0.000003           3         1           sigaltstack
------ ----------- ----------- --------- --------- ----------------
100.00    0.000713                   131         4 total
```

Nodejs:

```
% time     seconds  usecs/call     calls    errors syscall
------ ----------- ----------- --------- --------- ----------------
 20.15    0.000636           7        92           mmap
 17.78    0.000561          16        36           munmap
 13.97    0.000441          18        24           read
  7.73    0.000244           7        35           mprotect
  7.70    0.000243          15        16           brk
  7.32    0.000231           7        34         1 futex
  4.56    0.000144           7        22           open
  3.61    0.000114           5        22        15 ioctl
  2.15    0.000068           3        21           close
  2.03    0.000064           3        21           fstat
  2.00    0.000063           5        14        14 access
  1.58    0.000050           4        12           lstat
  1.24    0.000039           8         5           write
  1.24    0.000039           6         7           rt_sigaction
  1.05    0.000033           6         6         2 stat
  0.89    0.000028          28         1           readlink
  0.76    0.000024           5         5           rt_sigprocmask
  0.70    0.000022           6         4           getcwd
  0.67    0.000021          11         2           pipe2
  0.63    0.000020          20         1           clone
  0.38    0.000012           6         2           getrlimit
  0.29    0.000009           9         1           epoll_create1
  0.25    0.000008           8         1           eventfd2
  0.22    0.000007           7         1           clock_gettime
  0.22    0.000007           4         2           epoll_ctl
  0.19    0.000006           6         1           gettid
  0.19    0.000006           6         1           set_tid_address
  0.19    0.000006           6         1           set_robust_list
  0.13    0.000004           4         1           execve
  0.10    0.000003           3         1           arch_prctl
  0.10    0.000003           3         1           epoll_wait
------ ----------- ----------- --------- --------- ----------------
100.00    0.003156                   393        32 total
```

## Changelog

### v0.1.0 (2014-06-22)

Initial release

### v0.2.0 (2014-06-24)

Make module capable with Nodejs

### v0.3.0 (2014-06-26)

Rewrite module.

Make it easier to write addons.

Add underscore addon as an example.