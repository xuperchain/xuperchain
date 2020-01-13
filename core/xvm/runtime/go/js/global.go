package js

// Global simulates js'global object
type Global struct {
	properties map[string]interface{}
}

// NewGlobal instances a global object
func NewGlobal() *Global {
	return &Global{
		properties: make(map[string]interface{}),
	}
}

// Register property to global object
func (g *Global) Register(name string, prop interface{}) {
	g.properties[name] = prop
}

// GetProperty implements js.PropertyGetter interface
func (g *Global) GetProperty(name string) (interface{}, bool) {
	v, ok := g.properties[name]
	return v, ok
}
