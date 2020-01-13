package native

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/docker/go-connections/sockets"
	"github.com/xuperchain/xuperunion/contractsdk/go/code"
	pbrpc "github.com/xuperchain/xuperunion/contractsdk/go/pbrpc"
	"google.golang.org/grpc"
)

const (
	xchainUnixSocketGid = "XCHAIN_UNIXSOCK_GID"
	xchainPingTimeout   = "XCHAIN_PING_TIMEOUT"
)

func redirectStderr() {
	f, err := os.OpenFile("stderr.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err == nil {
		syscall.Dup2(int(f.Fd()), 2)
		f.Close()
	}
}

type driver struct {
}

// New returns a native driver
func New() code.Driver {
	return new(driver)
}

func (d *driver) Serve(contract code.Contract) {
	redirectStderr()

	var (
		flagset       = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		sockpath      = flagset.String("sock", "", "the path of unix socket file(if use unix socket)")
		chainSockpath = flagset.String("chain-sock", "", "the path of block chain service unix socket file(if use unix socket)")
		listenport    = flagset.String("port", "", "the listen port(if use tcp)")
	)
	flagset.Parse(os.Args[1:])

	nativeCodeService := newNativeCodeService(*chainSockpath, contract)
	rpcServer := grpc.NewServer()
	pbrpc.RegisterNativeCodeServer(rpcServer, nativeCodeService)

	var err error
	var listener net.Listener
	if *sockpath != "" {
		uid := os.Getuid()
		gid := getUnixSocketGroupid()
		relpath, err := relPathOfCWD(*sockpath)
		if err != nil {
			panic(err)
		}
		listener, err = sockets.NewUnixSocketWithOpts(relpath, sockets.WithChown(uid, gid), sockets.WithChmod(0660))
		if err != nil {
			panic(err)
		}
	} else if *listenport != "" {
		listener, err = sockets.NewTCPSocket(":"+*listenport, nil)
		if err != nil {
			panic(err)
		}
	} else {
		panic("empty --sock and --port")
	}

	go rpcServer.Serve(listener)

	sigch := make(chan os.Signal, 2)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM, syscall.SIGPIPE)

	timer := time.NewTicker(1 * time.Second)
	running := true
	pingTimeout := getPingTimeout()
	for running {
		select {
		case sig := <-sigch:
			running = false
			log.Print("receive signal ", sig)
		case <-timer.C:
			lastping := nativeCodeService.LastpingTime()
			if time.Since(lastping) > pingTimeout {
				log.Print("ping timeout")
				running = false
			}
		}
	}
	rpcServer.GracefulStop()
	nativeCodeService.Close()
	log.Print("native code ended")
}

func getUnixSocketGroupid() int {
	envgid := os.Getenv(xchainUnixSocketGid)
	if envgid == "" {
		return os.Getgid()
	}
	gid, err := strconv.Atoi(envgid)
	if err != nil {
		return os.Getgid()
	}
	return gid
}

func getPingTimeout() time.Duration {
	envtimeout := os.Getenv(xchainPingTimeout)
	if envtimeout == "" {
		return 3 * time.Second
	}
	timeout, err := strconv.Atoi(envtimeout)
	if err != nil {
		return 3 * time.Second
	}
	return time.Duration(timeout) * time.Second
}

//RelPathOfCWD 返回工作目录的相对路径
func relPathOfCWD(rootpath string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	socketPath, err := filepath.Rel(cwd, rootpath)
	if err != nil {
		return "", err
	}
	return socketPath, nil
}
