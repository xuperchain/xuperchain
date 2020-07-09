package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"sync"
	"time"

	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"

	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/server/xendorser"
)

// ProxyXEndorser is a proxy for XEndorser service
type ProxyXEndorser struct {
	clientCache sync.Map
	mutex       sync.Mutex
	conf        *Config
}

// Config the config of endorser
type Config struct {
	Hosts []string `yaml:"hosts"`
}

// make sure this plugin implemented the interface
var _ xendorser.XEndorser = (*ProxyXEndorser)(nil)

// GetInstance returns the an instance of DefaultXEndorser
func GetInstance() interface{} {
	return NewProxyXEndorser()
}

// NewProxyXEndorser create instance of DefaultXEndorser
func NewProxyXEndorser() *ProxyXEndorser {
	return &ProxyXEndorser{}
}

// Init initialize the plugin instance with params
func (pxe *ProxyXEndorser) Init(confPath string, params map[string]interface{}) error {
	if err := pxe.getConf(confPath); err != nil {
		return err
	}
	return nil
}

// EndorserCall process endorser call
func (pxe *ProxyXEndorser) EndorserCall(ctx context.Context, req *pb.EndorserRequest) (*pb.EndorserResponse, error) {
	xendc, err := pxe.getClient(pxe.getHost())
	if err != nil {
		return nil, err
	}
	return xendc.EndorserCall(ctx, req)
}

func (pxe *ProxyXEndorser) getHost() string {
	host := ""
	hostCnt := len(pxe.conf.Hosts)
	if hostCnt > 0 {
		rand.Seed(time.Now().Unix())
		index := rand.Intn(hostCnt)
		host = pxe.conf.Hosts[index]
	}
	return host
}

func (pxe *ProxyXEndorser) getClient(host string) (pb.XendorserClient, error) {
	if host == "" {
		return nil, fmt.Errorf("empty host")
	}
	if c, ok := pxe.clientCache.Load(host); ok {
		return c.(pb.XendorserClient), nil
	}

	pxe.mutex.Lock()
	defer pxe.mutex.Unlock()
	if c, ok := pxe.clientCache.Load(host); ok {
		return c.(pb.XendorserClient), nil
	}
	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	c := pb.NewXendorserClient(conn)
	pxe.clientCache.Store(host, c)
	return c, nil
}

func (pxe *ProxyXEndorser) getConf(path string) error {
	c := &Config{}
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return err
	}
	pxe.conf = c
	return nil
}
