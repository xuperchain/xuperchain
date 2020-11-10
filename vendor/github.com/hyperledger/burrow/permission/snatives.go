// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package permission

import (
	"fmt"
	"strings"

	"github.com/hyperledger/burrow/crypto"
)

//---------------------------------------------------------------------------------------------------
// PermissionsTx.PermArgs interface and argument encoding

func (pa PermArgs) String() string {
	body := make([]string, 0, 5)
	body = append(body, fmt.Sprintf("PermFlag: %v", String(pa.Action)))
	if pa.Target != nil {
		body = append(body, fmt.Sprintf("Address: %s", *pa.Target))
	}
	if pa.Permission != nil {
		body = append(body, fmt.Sprintf("Permission: %v", String(*pa.Permission)))
	}
	if pa.Role != nil {
		body = append(body, fmt.Sprintf("Role: %s", *pa.Role))
	}
	if pa.Value != nil {
		body = append(body, fmt.Sprintf("Value: %v", *pa.Value))
	}
	return fmt.Sprintf("PermArgs{%s}", strings.Join(body, ", "))
}

func (pa PermArgs) EnsureValid() error {
	pf := pa.Action
	// Address
	if pa.Target == nil && pf != SetGlobal {
		return fmt.Errorf("PermArgs for PermFlag %v requires Address to be provided but was nil", pf)
	}
	if pf == HasRole || pf == AddRole || pf == RemoveRole {
		// Role
		if pa.Role == nil {
			return fmt.Errorf("PermArgs for PermFlag %v requires Role to be provided but was nil", pf)
		}
		// Permission
	} else if pa.Permission == nil {
		return fmt.Errorf("PermArgs for PermFlag %v requires Permission to be provided but was nil", pf)
		// Value
	} else if (pf == SetBase || pf == SetGlobal) && pa.Value == nil {
		return fmt.Errorf("PermArgs for PermFlag %v requires Value to be provided but was nil", pf)
	}
	return nil
}

func HasBaseArgs(address crypto.Address, permFlag PermFlag) PermArgs {
	return PermArgs{
		Action:     HasBase,
		Target:     &address,
		Permission: &permFlag,
	}
}

func SetBaseArgs(address crypto.Address, permFlag PermFlag, value bool) PermArgs {
	return PermArgs{
		Action:     SetBase,
		Target:     &address,
		Permission: &permFlag,
		Value:      &value,
	}
}

func UnsetBaseArgs(address crypto.Address, permFlag PermFlag) PermArgs {
	return PermArgs{
		Action:     UnsetBase,
		Target:     &address,
		Permission: &permFlag,
	}
}

func SetGlobalArgs(permFlag PermFlag, value bool) PermArgs {
	return PermArgs{
		Action:     SetGlobal,
		Permission: &permFlag,
		Value:      &value,
	}
}

func HasRoleArgs(address crypto.Address, role string) PermArgs {
	return PermArgs{
		Action: HasRole,
		Target: &address,
		Role:   &role,
	}
}

func AddRoleArgs(address crypto.Address, role string) PermArgs {
	return PermArgs{
		Action: AddRole,
		Target: &address,
		Role:   &role,
	}
}

func RemoveRoleArgs(address crypto.Address, role string) PermArgs {
	return PermArgs{
		Action: RemoveRole,
		Target: &address,
		Role:   &role,
	}
}
