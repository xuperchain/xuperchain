package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	"github.com/xuperchain/xuperunion/contractsdk/go/code"
	"github.com/xuperchain/xuperunion/contractsdk/go/driver"
)

type math struct{}

func (m *math) Initialize(nci code.Context) code.Response {
	body := ""
	for key, value := range nci.Args() {
		err := nci.PutObject([]byte(key), value)
		if err != nil {
			return code.Error(err)
		}

		body += fmt.Sprintf("[%s]=[%s]", key, value)
	}

	return code.OK([]byte(body))
}

func (m *math) Invoke(nci code.Context) code.Response {
	var resp code.Response
	args := nci.Args()
	body := map[string]string{}
	action := string(args["action"])
	if action == "query" {
		for key := range args {
			if key == "action" {
				continue
			}
			res, err := nci.GetObject([]byte(key))
			if err != nil {
				return code.Error(err)
			}
			body[key] = string(res)
		}
	} else if action == "querytx" {
		id := string(args["id"])
		rawid, _ := hex.DecodeString(id)
		tx, err := nci.QueryTx(string(rawid))
		if err != nil {
			return code.Error(err)
		}
		out, _ := json.MarshalIndent(tx, "", "  ")
		os.Stderr.Write(out)
	} else if action == "queryblock" {
		id := string(args["id"])
		rawid, _ := hex.DecodeString(id)
		block, err := nci.QueryBlock(string(rawid))
		if err != nil {
		}
		out, _ := json.MarshalIndent(block, "", "  ")
		os.Stderr.Write(out)
	} else {
		for key, value := range args {
			if key == "action" {
				continue
			}
			res, err := nci.GetObject([]byte(key))
			if err != nil {
				return code.Error(err)
			}

			tmp, ok := new(big.Int).SetString(string(res), 10)
			if !ok {
				return code.Error(fmt.Errorf("tmp is %s, cann't convert to int", string(res)))
			}
			tmpb, ok := new(big.Int).SetString(string(value), 10)
			if !ok {
				return code.Error(fmt.Errorf("tmp is %s, cann't convert to int", string(res)))
			}
			tmp.Add(tmp, tmpb)
			/*
				fmt.Println(value)
				tmp := new(big.Int).SetUint64(10)
			*/

			err = nci.PutObject([]byte(key), []byte(tmp.String()))
			if err != nil {
				return code.Error(err)
			}

			res, err = nci.GetObject([]byte(key))
			body[key] = string(res)
		}
	}
	bodyStr, _ := json.Marshal(body)
	resp = code.Response{
		Status:  200,
		Body:    []byte(nil),
		Message: string(bodyStr),
	}
	return resp
}

func (m *math) Query(nci code.Context) code.Response {
	key, ok := nci.Args()["key"]
	if !ok {
		return code.Errors("argument keys not found")
	}
	value, _ := nci.GetObject(key)
	return code.OK(value)
}

func main() {
	driver.Serve(new(math))
}
