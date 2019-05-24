package exec

// MapResolver is the Resolver which stores symbols in map
type MapResolver map[string]interface{}

// ResolveFunc implements Resolver interface
func (m MapResolver) ResolveFunc(module, name string) (interface{}, bool) {
	v, ok := m[module+"."+name]
	if !ok {
		return nil, false
	}
	_, ok = v.(float64)
	if !ok {
		return v, true
	}
	return nil, false
}

// ResolveGlobal implements Resolver interface
func (m MapResolver) ResolveGlobal(module, name string) (float64, bool) {
	v, ok := m[module+"."+name]
	if !ok {
		return 0, false
	}
	ret, ok := v.(float64)
	if !ok {
		return 0, false
	}
	return ret, true
}
