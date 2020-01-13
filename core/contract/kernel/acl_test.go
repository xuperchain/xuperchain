package kernel

import (
	"io/ioutil"
	"os"
	"testing"

	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperchain/core/contract"
	crypto_client "github.com/xuperchain/xuperchain/core/crypto/client"
	"github.com/xuperchain/xuperchain/core/kv/kvdb"
	"github.com/xuperchain/xuperchain/core/ledger"
	"github.com/xuperchain/xuperchain/core/pluginmgr"
	"github.com/xuperchain/xuperchain/core/xmodel"
)

const DefaultKvEngine = "default"

var logger log.Logger

func openDB(dbPath string, logger log.Logger) (kvdb.Database, error) {
	plgMgr, plgErr := pluginmgr.GetPluginMgr()
	if plgErr != nil {
		logger.Warn("fail to get plugin manager")
		return nil, plgErr
	}
	var baseDB kvdb.Database
	soInst, err := plgMgr.PluginMgr.CreatePluginInstance("kv", "default")
	if err != nil {
		logger.Warn("fail to create plugin instance", "kvtype", "default")
		return nil, err
	}
	baseDB = soInst.(kvdb.Database)
	err = baseDB.Open(dbPath, map[string]interface{}{
		"cache":     128,
		"fds":       512,
		"dataPaths": []string{},
	})
	if err != nil {
		logger.Warn("xmodel::openDB failed to open db", "dbPath", dbPath)
		return nil, err
	}
	return baseDB, nil
}

func getXmodel(t *testing.T) *xmodel.XModel {
	logger = log.New("module", "xmodel")
	logger.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	path, pathErr := ioutil.TempDir("", "acl-test")
	if pathErr != nil {
		t.Fatal(pathErr)
	}
	ledger, err := ledger.NewLedger(path, nil, nil, DefaultKvEngine, crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Fatal(err)
	}
	stateDB, errDB := openDB(path+"/utxo_vm", logger)
	if errDB != nil {
		t.Fatal(err)
	}
	xModel, errModel := xmodel.NewXuperModel(ledger, stateDB, logger)
	if errModel != nil {
		t.Fatal(errModel)
	}
	return xModel
}

func TestNewAccountMethod(t *testing.T) {
	xModel := getXmodel(t)

	na := &NewAccountMethod{}
	saa := &SetAccountACLMethod{}
	testCase := map[string]struct {
		in          string
		accountName string
		expect      error
	}{
		"1": {
			in: `
            {
                "pm": {
                    "rule": 1,
                    "acceptValue": 1.0
                },
                "aksWeight": {
                    "AK1": 1.0,
                    "AK2": 1.0
                }
            }
            `,
			accountName: "1111111111111111",
			expect:      nil,
		},
		"2": {
			in: `
            {
                "pm": {
                    "rule": 1,
                    "acceptValue": 1.0
                },
                "aksWeight": {
                    "AK1": 1.0,
                    "AK2": 1.0,
                    "AK3": 1.0
                }
            }
            `,
			accountName: "XC1111111111111111@xuper",
			expect:      nil,
		},
	}
	modelCache, err := xmodel.NewXModelCache(xModel, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := &KContext{
		ModelCache:    modelCache,
		ResourceLimit: contract.MaxLimits,
		ContextConfig: &contract.ContextConfig{
			BCName: "xuper",
		},
	}
	arr := [2]string{"1", "2"}
	for _, value := range arr {
		if value == "1" {
			args := map[string][]byte{
				"account_name": []byte(testCase[value].accountName),
				"acl":          []byte(testCase[value].in),
			}
			_, err := na.Invoke(ctx, args)
			if err != testCase[value].expect {
				t.Fatal(err)
			}
		}
		if value == "2" {
			args := map[string][]byte{
				"account_name": []byte(testCase[value].accountName),
				"acl":          []byte(testCase[value].in),
			}
			_, err := saa.Invoke(ctx, args)
			if err != testCase[value].expect {
				t.Fatal(err)
			}
		}
	}
}

func TestSetContractMethodAclMethod(t *testing.T) {
	xModel := getXmodel(t)

	sma := &SetMethodACLMethod{}
	testCases := map[string]struct {
		in           string
		contractName string
		methodName   string
		expect       error
	}{
		"1": {
			in: `
            {
                "pm": {
                    "rule": 1,
                    "acceptValue": 1.0
                },
                "aksWeight": {
                    "AK1": 1.0,
                    "AK2": 1.0
                }
            }
            `,
			expect: nil,
		},
	}

	modelCache, err := xmodel.NewXModelCache(xModel, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := &KContext{
		ModelCache:    modelCache,
		ResourceLimit: contract.MaxLimits,
	}
	for _, testCase := range testCases {
		args := map[string][]byte{
			"contract_name": []byte(testCase.contractName),
			"method_name":   []byte(testCase.methodName),
			"acl":           []byte(testCase.in),
		}
		_, err := sma.Invoke(ctx, args)
		if err != testCase.expect {
			t.Fatal(err)
		}
	}

}
