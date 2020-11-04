package spec

import (
	"sort"

	"github.com/hyperledger/burrow/acm/balance"
	"github.com/hyperledger/burrow/permission"
)

// Files here can be used as starting points for building various 'chain types' but are otherwise
// a fairly unprincipled collection of GenesisSpecs that we find useful in testing and development

func FullAccount(name string) GenesisSpec {
	// Inheriting from the arbitrary figures used by monax tool for now
	amount := uint64(99999999999999)
	Power := uint64(9999999999)
	return GenesisSpec{
		Accounts: []TemplateAccount{{
			Name:        name,
			Amounts:     balance.New().Native(amount).Power(Power),
			Permissions: []string{permission.AllString},
		},
		},
	}
}

func RootAccount(name string) GenesisSpec {
	// Inheriting from the arbitrary figures used by monax tool for now
	amount := uint64(99999999999999)
	return GenesisSpec{
		Accounts: []TemplateAccount{{
			Name:        name,
			Amounts:     balance.New().Native(amount),
			Permissions: []string{permission.AllString},
		},
		},
	}
}

func ParticipantAccount(name string) GenesisSpec {
	// Inheriting from the arbitrary figures used by monax tool for now
	amount := uint64(9999999999)
	return GenesisSpec{
		Accounts: []TemplateAccount{{
			Name:    name,
			Amounts: balance.New().Native(amount),
			Permissions: []string{permission.SendString, permission.CallString, permission.NameString,
				permission.HasRoleString, permission.ProposalString, permission.InputString},
		}},
	}
}

func DeveloperAccount(name string) GenesisSpec {
	// Inheriting from the arbitrary figures used by monax tool for now
	amount := uint64(9999999999)
	return GenesisSpec{
		Accounts: []TemplateAccount{{
			Name:    name,
			Amounts: balance.New().Native(amount),
			Permissions: []string{permission.SendString, permission.CallString, permission.CreateContractString,
				permission.CreateAccountString, permission.NameString, permission.HasRoleString,
				permission.RemoveRoleString, permission.ProposalString, permission.InputString},
		}},
	}
}

func ValidatorAccount(name string) GenesisSpec {
	// Inheriting from the arbitrary figures used by monax tool for now
	amount := uint64(9999999999)
	Power := amount - 1
	return GenesisSpec{
		Accounts: []TemplateAccount{{
			Name:        name,
			Amounts:     balance.New().Native(amount).Power(Power),
			Permissions: []string{permission.BondString},
		}},
	}
}

func MergeGenesisSpecs(genesisSpecs ...GenesisSpec) GenesisSpec {
	mergedGenesisSpec := GenesisSpec{}
	// We will deduplicate and merge global permissions flags
	permSet := make(map[string]struct{})

	for _, genesisSpec := range genesisSpecs {
		// We'll overwrite chain name for later specs
		if genesisSpec.ChainName != "" {
			mergedGenesisSpec.ChainName = genesisSpec.ChainName
		}
		// Take the max genesis time
		if mergedGenesisSpec.GenesisTime == nil ||
			(genesisSpec.GenesisTime != nil && genesisSpec.GenesisTime.After(*mergedGenesisSpec.GenesisTime)) {
			mergedGenesisSpec.GenesisTime = genesisSpec.GenesisTime
		}

		for _, permString := range genesisSpec.GlobalPermissions {
			permSet[permString] = struct{}{}
		}

		mergedGenesisSpec.Salt = append(mergedGenesisSpec.Salt, genesisSpec.Salt...)
		mergedGenesisSpec.Accounts = mergeAccounts(mergedGenesisSpec.Accounts, genesisSpec.Accounts)
	}

	mergedGenesisSpec.GlobalPermissions = make([]string, 0, len(permSet))

	for permString := range permSet {
		mergedGenesisSpec.GlobalPermissions = append(mergedGenesisSpec.GlobalPermissions, permString)
	}

	// Make sure merged GenesisSpec is deterministic on inputs
	sort.Strings(mergedGenesisSpec.GlobalPermissions)

	return mergedGenesisSpec
}

// Merge accounts by adding to base list or updating previously named account
func mergeAccounts(bases, overrides []TemplateAccount) []TemplateAccount {
	indexOfBase := make(map[string]int, len(bases))
	for i, ta := range bases {
		if ta.Name != "" {
			indexOfBase[ta.Name] = i
		}
	}

	for _, override := range overrides {
		if override.Name != "" {
			if i, ok := indexOfBase[override.Name]; ok {
				bases[i] = mergeAccount(bases[i], override)
				continue
			}
		}
		bases = append(bases, override)
	}
	return bases
}

func mergeAccount(base, override TemplateAccount) TemplateAccount {
	if override.Address != nil {
		base.Address = override.Address
	}
	if override.PublicKey != nil {
		base.PublicKey = override.PublicKey
	}
	if override.Name != "" {
		base.Name = override.Name
	}

	base.Amounts = base.Balances().Sum(override.Balances())

	base.Permissions = mergeStrings(base.Permissions, override.Permissions)
	base.Roles = mergeStrings(base.Roles, override.Roles)
	return base
}

func mergeStrings(as, bs []string) []string {
	var strs []string
	strSet := make(map[string]struct{})
	for _, a := range as {
		strSet[a] = struct{}{}
	}
	for _, b := range bs {
		strSet[b] = struct{}{}
	}
	for str := range strSet {
		strs = append(strs, str)
	}
	sort.Strings(strs)
	return strs
}
