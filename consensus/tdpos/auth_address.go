package tdpos

func (tp *TDpos) isAuthAddress(candidate string, initiator string, authRequire []string) bool {
	if candidate == initiator {
		return true
	}
	for _, value := range authRequire {
		if candidate == value {
			return true
		}
	}
	return false
}
