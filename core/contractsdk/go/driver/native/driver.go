package native

import (
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
	pbrpc "github.com/xuperchain/xuperchain/core/contractsdk/go/pbrpc"
	"google.golang.org/grpc"
)

const (
	xchainPingTimeout = "XCHAIN_PING_TIMEOUT"
	xchainCodePort    = "XCHAIN_CODE_PORT"
	xchainChainAddr   = "XCHAIN_CHAIN_ADDR"
)

type driver struct {
}

// New returns a native driver
func New() code.Driver {
	return new(driver)
}

func (d *driver) Serve(contract code.Contract) {
	chainAddr := os.Getenv(xchainChainAddr)
	codePort := os.Getenv(xchainCodePort)

	if chainAddr == "" {
		panic("empty XCHAIN_CHAIN_ADDR env")
	}

	if codePort == "" {
		panic("empty XCHAIN_CODE_PORT env")
	}

	nativeCodeService := newNativeCodeService(chainAddr, contract)
	rpcServer := grpc.NewServer()
	pbrpc.RegisterNativeCodeServer(rpcServer, nativeCodeService)

	var listener net.Listener
	listener, err := net.Listen("tcp", "127.0.0.1:"+codePort)
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
