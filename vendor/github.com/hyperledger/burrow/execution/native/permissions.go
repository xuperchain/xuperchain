package native

import (
	"fmt"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/permission"
)

var Permissions = New().MustContract("Permissions",
	`* Interface for managing Secure Native authorizations.
		* @dev This interface describes the functions exposed by the native permissions layer in burrow.
		`,
	Function{
		Comment: `
			* @notice Adds a role to an account
			* @param _account account address
			* @param _role role name
			* @return _result whether role was added
			`,
		PermFlag: permission.AddRole,
		F:        addRole,
	},
	Function{
		Comment: `
			* @notice Removes a role from an account
			* @param _account account address
			* @param _role role name
			* @return _result whether role was removed
			`,
		PermFlag: permission.RemoveRole,
		F:        removeRole,
	},
	Function{
		Comment: `
			* @notice Indicates whether an account has a role
			* @param _account account address
			* @param _role role name
			* @return _result whether account has role
			`,
		PermFlag: permission.HasRole,
		F:        hasRole,
	},
	Function{
		Comment: `
			* @notice Sets the permission flags for an account. Makes them explicitly set (on or off).
			* @param _account account address
			* @param _permission the base permissions flags to set for the account
			* @param _set whether to set or unset the permissions flags at the account level
			* @return _result is the permission flag that was set as uint64
			`,
		PermFlag: permission.SetBase,
		F:        setBase,
	},
	Function{
		Comment: `
			* @notice Unsets the permissions flags for an account. Causes permissions being unset to fall through to global permissions.
      		* @param _account account address
      		* @param _permission the permissions flags to unset for the account
			* @return _result is the permission flag that was unset as uint64
      `,
		PermFlag: permission.UnsetBase,
		F:        unsetBase,
	},
	Function{
		Comment: `
			* @notice Indicates whether an account has a subset of permissions set
			* @param _account account address
			* @param _permission the permissions flags (mask) to check whether enabled against base permissions for the account
			* @return _result is whether account has the passed permissions flags set
			`,
		PermFlag: permission.HasBase,
		F:        hasBase,
	},
	Function{Comment: `
			* @notice Sets the global (default) permissions flags for the entire chain
			* @param _permission the permissions flags to set
			* @param _set whether to set (or unset) the permissions flags
			* @return _result is the permission flag that was set as uint64
			`,
		PermFlag: permission.SetGlobal,
		F:        setGlobal,
	},
)

// CONTRACT: it is the duty of the contract writer to call known permissions
// we do not convey if a permission is not set
// (unlike in state/execution, where we guarantee HasPermission is called
// on known permissions and panics else)
// If the perm is not defined in the acc nor set by default in GlobalPermissions,
// this function returns false.
func HasPermission(st acmstate.Reader, address crypto.Address, perm permission.PermFlag) (bool, error) {
	acc, err := st.GetAccount(address)
	if err != nil {
		return false, err
	}
	if acc == nil {
		return false, fmt.Errorf("account %v does not exist", address)
	}
	globalPerms, err := acmstate.GlobalAccountPermissions(st)
	if err != nil {
		return false, err
	}
	perms := acc.Permissions.Base.Compose(globalPerms.Base)
	value, err := perms.Get(perm)
	if err != nil {
		return false, err
	}
	return value, nil
}

type hasBaseArgs struct {
	Account    crypto.Address
	Permission uint64
}

type hasBaseRets struct {
	Result bool
}

func hasBase(ctx Context, args hasBaseArgs) (hasBaseRets, error) {
	permN := permission.PermFlag(args.Permission) // already shifted
	if !permN.IsValid() {
		return hasBaseRets{}, permission.ErrInvalidPermission(permN)
	}
	hasPermission, err := HasPermission(ctx.State, args.Account, permN)
	if err != nil {
		return hasBaseRets{}, err
	}
	ctx.Logger.Trace.Log("function", "hasBase",
		"address", args.Account.String(),
		"perm_flag", fmt.Sprintf("%b", permN),
		"has_permission", hasPermission)
	return hasBaseRets{Result: hasPermission}, nil
}

type setBaseArgs struct {
	Account    crypto.Address
	Permission uint64
	Set        bool
}

type setBaseRets struct {
	Result uint64
}

func setBase(ctx Context, args setBaseArgs) (setBaseRets, error) {
	permFlag := permission.PermFlag(args.Permission)
	if !permFlag.IsValid() {
		return setBaseRets{}, permission.ErrInvalidPermission(permFlag)
	}
	err := UpdateAccount(ctx.State, args.Account, func(acc *acm.Account) error {
		err := acc.Permissions.Base.Set(permFlag, args.Set)
		return err
	})
	if err != nil {
		return setBaseRets{}, err
	}
	ctx.Logger.Trace.Log("function", "setBase", "address", args.Account.String(),
		"permission_flag", fmt.Sprintf("%b", permFlag),
		"permission_value", args.Permission)
	return setBaseRets{Result: uint64(permFlag)}, nil
}

type unsetBaseArgs struct {
	Account    crypto.Address
	Permission uint64
}

type unsetBaseRets struct {
	Result uint64
}

func unsetBase(ctx Context, args unsetBaseArgs) (unsetBaseRets, error) {
	permFlag := permission.PermFlag(args.Permission)
	if !permFlag.IsValid() {
		return unsetBaseRets{}, permission.ErrInvalidPermission(permFlag)
	}
	err := UpdateAccount(ctx.State, args.Account, func(acc *acm.Account) error {
		return acc.Permissions.Base.Unset(permFlag)
	})
	if err != nil {
		return unsetBaseRets{}, err
	}
	ctx.Logger.Trace.Log("function", "unsetBase", "address", args.Account.String(),
		"perm_flag", fmt.Sprintf("%b", permFlag),
		"permission_flag", fmt.Sprintf("%b", permFlag))

	return unsetBaseRets{Result: uint64(permFlag)}, nil
}

type setGlobalArgs struct {
	Permission uint64
	Set        bool
}

type setGlobalRets struct {
	Result uint64
}

func setGlobal(ctx Context, args setGlobalArgs) (setGlobalRets, error) {
	permFlag := permission.PermFlag(args.Permission)
	if !permFlag.IsValid() {
		return setGlobalRets{}, permission.ErrInvalidPermission(permFlag)
	}
	err := UpdateAccount(ctx.State, acm.GlobalPermissionsAddress, func(acc *acm.Account) error {
		return acc.Permissions.Base.Set(permFlag, args.Set)
	})
	if err != nil {
		return setGlobalRets{}, err
	}
	ctx.Logger.Trace.Log("function", "setGlobal",
		"permission_flag", fmt.Sprintf("%b", permFlag),
		"permission_value", args.Set)
	return setGlobalRets{Result: uint64(permFlag)}, nil
}

type hasRoleArgs struct {
	Account crypto.Address
	Role    string
}

type hasRoleRets struct {
	Result bool
}

func hasRole(ctx Context, args hasRoleArgs) (hasRoleRets, error) {
	acc, err := mustAccount(ctx.State, args.Account)
	if err != nil {
		return hasRoleRets{}, err
	}
	hasRole := acc.Permissions.HasRole(args.Role)
	ctx.Logger.Trace.Log("function", "hasRole", "address", args.Account.String(),
		"role", args.Role,
		"has_role", hasRole)
	return hasRoleRets{Result: hasRole}, nil
}

type addRoleArgs struct {
	Account crypto.Address
	Role    string
}

type addRoleRets struct {
	Result bool
}

func addRole(ctx Context, args addRoleArgs) (addRoleRets, error) {
	ret := addRoleRets{}
	err := UpdateAccount(ctx.State, args.Account, func(account *acm.Account) error {
		ret.Result = account.Permissions.AddRole(args.Role)
		return nil
	})
	if err != nil {
		return ret, err
	}
	ctx.Logger.Trace.Log("function", "addRole", "address", args.Account.String(),
		"role", args.Role,
		"role_added", ret.Result)
	return ret, nil
}

type removeRoleArgs struct {
	Account crypto.Address
	Role    string
}

type removeRoleRets struct {
	Result bool
}

func removeRole(ctx Context, args removeRoleArgs) (removeRoleRets, error) {
	ret := removeRoleRets{}
	err := UpdateAccount(ctx.State, args.Account, func(account *acm.Account) error {
		ret.Result = account.Permissions.RemoveRole(args.Role)
		return nil
	})
	if err != nil {
		return ret, err
	}
	ctx.Logger.Trace.Log("function", "removeRole", "address", args.Account.String(),
		"role", args.Role,
		"role_removed", ret.Result)
	return ret, nil
}
