// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package permission

var (
	ZeroBasePermissions = BasePermissions{
		Perms:  0,
		SetBit: 0,
	}
	ZeroAccountPermissions = AccountPermissions{
		Base: ZeroBasePermissions,
	}
	DefaultAccountPermissions = AccountPermissions{
		Base: BasePermissions{
			Perms:  DefaultPermFlags,
			SetBit: AllPermFlags,
		},
		Roles: []string{},
	}
	AllAccountPermissions = AccountPermissions{
		Base: BasePermissions{
			Perms:  AllPermFlags,
			SetBit: AllPermFlags,
		},
		Roles: []string{},
	}
)
