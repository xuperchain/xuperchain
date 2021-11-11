/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperchain/service/pb"
	"github.com/xuperchain/xupercore/lib/utils"
)

// StatusCommand status cmd
type StatusCommand struct {
	cli *Cli
	cmd *cobra.Command

	ledger bool
	utxo   bool
	branch bool
	peers  bool
}

// NewStatusCommand new status cmd
func NewStatusCommand(cli *Cli) *cobra.Command {
	s := new(StatusCommand)
	s.cli = cli
	s.cmd = &cobra.Command{
		Use:   "status",
		Short: "Operate a command to get status of current xchain server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return s.printXchainStatus(ctx)
		},
	}
	s.addFlags()
	return s.cmd
}

func (s *StatusCommand) addFlags() {
	s.cmd.Flags().BoolVarP(&s.ledger, "ledger", "L", false, "Get ledger info")
	s.cmd.Flags().BoolVarP(&s.utxo, "utxo", "U", false, "Get utxo info")
	s.cmd.Flags().BoolVarP(&s.branch, "branch", "B", false, "Get branch info")
	s.cmd.Flags().BoolVarP(&s.peers, "peers", "P", false, "Get peers info")
}

func (s *StatusCommand) printXchainStatus(ctx context.Context) error {
	client := s.cli.XchainClient()
	req := &pb.CommonIn{
		Header: &pb.Header{
			Logid: utils.GenLogId(),
		},
		ViewOption: s.convertToFlag(),
	}
	reply, err := client.GetSystemStatus(ctx, req)
	if err != nil {
		return err
	}
	if reply.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return errors.New(reply.Header.Error.String())
	}
	status := FromSystemStatusPB(reply.GetSystemsStatus(), s.cli.RootOptions.Name)
	if s.extractSpecificInfo(status) {
		return nil
	}
	output, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(output))
	return nil
}

// convert to flag(viewOption)
func (s *StatusCommand) convertToFlag() pb.ViewOption {
	if s.ledger {
		return pb.ViewOption_LEDGER
	} else if s.utxo {
		return pb.ViewOption_UTXOINFO
	} else if s.branch {
		return pb.ViewOption_BRANCHINFO
	} else if s.peers {
		return pb.ViewOption_PEERS
	} else {
		return pb.ViewOption_NONE
	}
}

func (s *StatusCommand) extractSpecificInfo(status *SystemStatus) bool {
	handled := false
	if s.ledger {
		type LedgerInfo struct {
			Name       string     `json:"name"`
			LedgerMeta LedgerMeta `json:"ledger"`
		}
		var ledgerInfos []LedgerInfo
		for _, chainStatus := range status.ChainStatus {
			ledgerInfos = append(ledgerInfos, LedgerInfo{
				Name:       chainStatus.Name,
				LedgerMeta: chainStatus.LedgerMeta,
			})
		}
		output, err := json.MarshalIndent(ledgerInfos, "", "  ")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(output))
		handled = true
	} else if s.utxo {
		type UtxoMetaInfo struct {
			Name     string   `json:"name"`
			UtxoMeta UtxoMeta `json:"utxo"`
		}
		var utxoMetaInfos []UtxoMetaInfo
		for _, chainStatus := range status.ChainStatus {
			utxoMetaInfos = append(utxoMetaInfos, UtxoMetaInfo{
				Name:     chainStatus.Name,
				UtxoMeta: chainStatus.UtxoMeta,
			})
		}
		output, err := json.MarshalIndent(utxoMetaInfos, "", "  ")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(output))
		handled = true
	} else if s.branch {
		type BranchInfo struct {
			Name             string   `json:"name"`
			BifurcationRatio float64  `json:"bifurcationRatio"`
			BranchBlockid    []string `json:"branchBlockid"`
		}
		var branchInfos []BranchInfo
		for _, chainStatus := range status.ChainStatus {
			branchInfos = append(branchInfos, BranchInfo{
				Name:             chainStatus.Name,
				BifurcationRatio: float64(len(chainStatus.BranchBlockid)) / float64(chainStatus.LedgerMeta.TrunkHeight),
				BranchBlockid:    chainStatus.BranchBlockid,
			})
		}
		output, err := json.MarshalIndent(branchInfos, "", "  ")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(output))
		handled = true
	} else if s.peers {
		peers := status.Peers
		output, err := json.MarshalIndent(peers, "", "  ")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(output))
		handled = true
	}
	return handled
}

func init() {
	AddCommand(NewStatusCommand)
}
