package builtins

import "github.com/ddliu/motto"

// RegisterModule register a builtin module using name as module name
func RegisterModule(name string, source string) {
	loader := motto.CreateLoaderFromSource(source, name+".js")
	motto.AddModule(name, loader)
}
