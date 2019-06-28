# About
This depository is a fork of log15. Add time-based log rotation and some other improvements.

For the original [log15](https://github.com/inconshreveable/log15) depository, please click the link.



# log15 

Package log15 provides an opinionated, simple toolkit for best-practice logging in Go (golang) that is both human and machine readable. It is modeled after the Go standard library's [`io`](http://golang.org/pkg/io/) and [`net/http`](http://golang.org/pkg/net/http/) packages and is an alternative to the standard library's [`log`](http://golang.org/pkg/log/) package.

## Features
- A simple, easy-to-understand API
- Promotes structured logging by encouraging use of key/value pairs
- Child loggers which inherit and add their own private context
- Lazy evaluation of expensive operations
- Simple Handler interface allowing for construction of flexible, custom logging configurations with a tiny API.
- Color terminal support
- Built-in support for logging to files, streams, syslog, and the network
- Support for forking records to multiple handlers, buffering records for output, failing over from failed handler writes, + more
- **Time based log rotation**. Support log rotation interval in minutes, and keep specifed numbers of old log files.
- **Add a BoundLvlFilterHandler**, which accepts two log verbosal level as a boundary. Only write the log between minLvl and maxLvl.
- **Add a new Trace log level**,which is between Info and Debug.

## Versioning
The API of the master branch of log15 should always be considered unstable. If you want to rely on a stable API,
you must vendor the library.

## Importing

```go
import "github.com/xuperchain/log15"
```

## Log Levels
```go
// List of predefined log Levels
const (
	LvlCrit Lvl = iota
	LvlError
	LvlWarn
	LvlInfo
	LvlTrace
	LvlDebug
)
```

## Examples

```go
// all loggers can have key/value context
srvlog := log.New("module", "app/server")

// all log messages can have key/value context
srvlog.Warn("abnormal conn rate", "rate", curRate, "low", lowRate, "high", highRate)

// child loggers with inherited context
connlog := srvlog.New("raddr", c.RemoteAddr())
connlog.Info("connection open")

// lazy evaluation
connlog.Debug("ping remote", "latency", log.Lazy{pingRemote})

// flexible configuration
srvlog.SetHandler(log.MultiHandler(
    log.StreamHandler(os.Stderr, log.LogfmtFormat()),
    log.LvlFilterHandler(
        log.LvlError,
        log.Must.FileHandler("errors.json", log.JsonFormat()))))
```

Will result in output that looks like this:

```
WARN[06-17|21:58:10] abnormal conn rate                       module=app/server rate=0.500 low=0.100 high=0.800
INFO[06-17|21:58:10] connection open                          module=app/server raddr=10.0.0.1
```

## Log Rotate Example
Using RotateFileHandler to create time-based rotation strategy.

```go
l := New()
fmtr := LogfmtFormat()
// rotate every 1 minutes and keep 2 backup log files
l.SetHandler(MultiHandler(
    BoundLvlFilterHandler(
        LvlDebug, 
        LvlInfo, 
        Must.RotateFileHandler("./log/test.log", fmtr, 1, 2)),
    LvlFilterHandler(
        LvlWarn, 
        Must.RotateFileHandler("./log/test.log.wf", fmtr, 1, 2))))

times := 150
for i := 0; i < times; i++ {
    l.Info("this is a info", "index", i)
    l.Warn("this is a warn", "index", i)
    time.Sleep(time.Second * 1)
}
```

The log file look like:

```
-rw-r--r-- 1 work work 990 Dec 18 13:48 test.log
-rw-r--r-- 1 work work   0 Dec 18 13:48 test.log.201812181347
-rw-r--r-- 1 work work 990 Dec 18 13:48 test.log.wf
-rw-r--r-- 1 work work   0 Dec 18 13:48 test.log.wf.201812181347
```

## Breaking API Changes
The following commits broke API stability. This reference is intended to help you understand the consequences of updating to a newer version
of log15.

- 57a084d014d4150152b19e4e531399a7145d1540 - Added a `Get()` method to the `Logger` interface to retrieve the current handler
- 93404652ee366648fa622b64d1e2b67d75a3094a - `Record` field `Call` changed to `stack.Call` with switch to `github.com/go-stack/stack`
- a5e7613673c73281f58e15a87d2cf0cf111e8152 - Restored `syslog.Priority` argument to the `SyslogXxx` handler constructors

## FAQ

### The varargs style is brittle and error prone! Can I have type safety please?
Yes. Use `log.Ctx`:

```go
srvlog := log.New(log.Ctx{"module": "app/server"})
srvlog.Warn("abnormal conn rate", log.Ctx{"rate": curRate, "low": lowRate, "high": highRate})
```

### Regenerating the CONTRIBUTORS file

```
go get -u github.com/kevinburke/write_mailmap
write_mailmap > CONTRIBUTORS
```

## License
Apache
