// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"fmt"
	"strings"

	. "github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/tmthrgd/go-hex"
)

const logNTextTopicCutset = "\x00"
const LogNKeyPrefix = "Log"

func LogNKey(topic int) string {
	return fmt.Sprintf("%s%d", LogNKeyPrefix, topic)
}

func LogNTextKey(topic int) string {
	return fmt.Sprintf("%s%dText", LogNKeyPrefix, topic)
}

var logTagKeys []string
var logNTopicIndex = make(map[string]int, 5)
var logNTextTopicIndex = make(map[string]int, 5)

func init() {
	for i := 0; i <= 4; i++ {
		logN := LogNKey(i)
		logTagKeys = append(logTagKeys, LogNKey(i))
		logNText := LogNTextKey(i)
		logTagKeys = append(logTagKeys, logNText)
		logNTopicIndex[logN] = i
		logNTextTopicIndex[logNText] = i
	}
	logTagKeys = append(logTagKeys, event.AddressKey)
}

func (log *LogEvent) Get(key string) (interface{}, bool) {
	if log == nil {
		return "", false
	}
	var value interface{}
	switch key {
	case event.AddressKey:
		value = log.Address
	default:
		if i, ok := logNTopicIndex[key]; ok {
			return hex.EncodeUpperToString(log.GetTopic(i).Bytes()), true
		}
		if i, ok := logNTextTopicIndex[key]; ok {
			return strings.Trim(string(log.GetTopic(i).Bytes()), logNTextTopicCutset), true
		}
		return "", false
	}
	return value, true
}

func (log *LogEvent) GetTopic(i int) Word256 {
	if i < len(log.Topics) {
		return log.Topics[i]
	}
	return Word256{}
}

func (log *LogEvent) SolidityEventID() abi.EventID {
	var eventID abi.EventID
	copy(eventID[:], log.Topics[0].Bytes())
	return eventID
}
