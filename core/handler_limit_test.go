package xchaincore

import (
	"fmt"
	"github.com/xuperchain/xuperunion/p2pv2/pb"
	"testing"
	"time"
)

func handleMsg(msg *xuperp2p.XuperMessage) {
	// TODO
	if msg != nil {
		fmt.Println("xuperMessage ", msg)
	}
}

func TestHandleLimitBasic(t *testing.T) {
	bcname := "xuper"
	logid := "123456789"
	res, _ := xuperp2p.NewXuperMessage(xuperp2p.XuperMsgVersion2, bcname, logid,
		xuperp2p.XuperMessage_GET_BLOCK_RES, nil, xuperp2p.XuperMessage_CHECK_SUM_ERROR)
	msgDisp := &MessageDispatcher{}
	msgDisp.Init(10)
	<-time.After(2 * time.Second)
	msgDisp.Register(handleMsg)
	go msgDisp.Start()
	if msgDisp.NearFull() == false {
		msgDisp.msgDispChan <- res
	}
	<-time.After(3 * time.Second)
	msgDisp.Stop()
}
