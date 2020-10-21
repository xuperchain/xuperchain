// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package permission

import (
	"strings"
)

// ConvertMapStringIntToPermissions converts a map of string-bool pairs and a slice of
// strings for the roles to an AccountPermissions type. If the value in the
// permissions map is true for a particular permission string then the permission
// will be set in the AccountsPermissions. For all unmentioned permissions the
// ZeroBasePermissions is defaulted to.
func ConvertPermissionsMapAndRolesToAccountPermissions(permissions map[string]bool,
	roles []string) (*AccountPermissions, error) {
	var err error
	accountPermissions := &AccountPermissions{}
	accountPermissions.Base, err = convertPermissionsMapStringIntToBasePermissions(permissions)
	if err != nil {
		return nil, err
	}
	accountPermissions.Roles = roles
	return accountPermissions, nil
}

// convertPermissionsMapStringIntToBasePermissions converts a map of string-bool
// pairs to BasePermissions.
func convertPermissionsMapStringIntToBasePermissions(permissions map[string]bool) (BasePermissions, error) {
	// initialise basePermissions as ZeroBasePermissions
	basePermissions := ZeroBasePermissions

	for permissionName, value := range permissions {
		permissionsFlag, err := PermStringToFlag(permissionName)
		if err != nil {
			return basePermissions, err
		}
		// sets the permissions bitflag and the setbit flag for the permission.
		basePermissions.Set(permissionsFlag, value)
	}

	return basePermissions, nil
}

// Builds a composite BasePermission by creating a PermFlag from permissions strings and
// setting them all
func BasePermissionsFromStringList(permissions []string) (BasePermissions, error) {
	permFlag, err := PermFlagFromStringList(permissions)
	if err != nil {
		return ZeroBasePermissions, err
	}
	return BasePermissions{
		Perms:  permFlag,
		SetBit: permFlag,
	}, nil
}

// Builds a composite PermFlag by mapping each permission string in permissions to its
// flag and composing them with binary or
func PermFlagFromStringList(permissions []string) (PermFlag, error) {
	var permFlag PermFlag
	for _, perm := range permissions {
		s := strings.TrimSpace(perm)
		if s == "" {
			continue
		}
		flag, err := PermStringToFlag(s)
		if err != nil {
			return permFlag, err
		}
		permFlag |= flag
	}
	return permFlag, nil
}

// Builds a list of set permissions from a BasePermission by creating a list of permissions strings
// from the resultant permissions of basePermissions
func BasePermissionsToStringList(basePermissions BasePermissions) []string {
	return PermFlagToStringList(basePermissions.ResultantPerms())
}

// Creates a list of individual permission flag strings from a possibly composite PermFlag
// by projecting out each bit and adding its permission string if it is set
func PermFlagToStringList(permFlag PermFlag) []string {
	permStrings := make([]string, 0, NumPermissions)
	for i := uint(0); i < NumPermissions; i++ {
		permFlag := permFlag & (1 << i)
		if permFlag > 0 {
			permStrings = append(permStrings, permFlag.String())
		}
	}
	return permStrings
}

// Generates a human readable string from the resultant permissions of basePermission
func BasePermissionsString(basePermissions BasePermissions) string {
	return strings.Join(BasePermissionsToStringList(basePermissions), " | ")
}

func String(permFlag PermFlag) string {
	return strings.Join(PermFlagToStringList(permFlag), " | ")
}

func (pf PermFlag) MarshalText() ([]byte, error) {
	return []byte(String(pf)), nil
}

func (pf *PermFlag) UnmarshalText(s []byte) (err error) {
	*pf, err = PermFlagFromStringList(strings.Split(string(s), "|"))
	return
}
