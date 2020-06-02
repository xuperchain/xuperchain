package native

import (
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/docker/go-connections/sockets"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
	pbrpc "github.com/xuperchain/xuperchain/core/contractsdk/go/pbrpc"
	"google.golang.org/grpc"
)

const (
	xchainUnixSocketGid = "XCHAIN_UNIXSOCK_GID"
	xchainPingTimeout   = "XCHAIN_PING_TIMEOUT"
	xchainCodeSock      = "XCHAIN_CODE_SOCK"
	xchainChainSock     = "XCHAIN_CHAIN_SOCK"
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
	chainSockPath := os.Getenv(xchainChainSock)
	codeSockPath := os.Getenv(xchainCodeSock)

	nativeCodeService := newNativeCodeService(chainSockPath, contract)
	rpcServer := grpc.NewServer()
	pbrpc.RegisterNativeCodeServer(rpcServer, nativeCodeService)

	var listener net.Listener
	uid := os.Getuid()
	gid := getUnixSocketGroupid()
	relpath := normalizeSockPath(codeSockPath)
	listener, err := sockets.NewUnixSocketWithOpts(relpath, sockets.WithChown(uid, gid), sockets.WithChmod(0660))
	if err != nil {
		panic(err)
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

// normalizeSockPath make unix socket path as shorter as possiable
func normalizeSockPath(s string) string {
	if !filepath.IsAbs(s) {
		return s
	}

	wd, _ := os.Getwd()
	if !strings.HasPrefix(s, wd) {
		return s
	}
	relpath, _ := filepath.Rel(wd, s)
	return relpath
}
