// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"github.com/go-kit/kit/log/term"
	"github.com/hyperledger/burrow/logging/structure"
)

func Colors(keyvals ...interface{}) term.FgBgColor {
	for i := 0; i < len(keyvals)-1; i += 2 {
		if keyvals[i] != structure.LevelKey {
			continue
		}
		switch keyvals[i+1] {
		case "debug":
			return term.FgBgColor{Fg: term.DarkGray}
		case "info":
			return term.FgBgColor{Fg: term.Gray}
		case "warn":
			return term.FgBgColor{Fg: term.Yellow}
		case "error":
			return term.FgBgColor{Fg: term.Red}
		case "crit":
			return term.FgBgColor{Fg: term.Gray, Bg: term.DarkRed}
		default:
			return term.FgBgColor{}
		}
	}
	return term.FgBgColor{}
}
