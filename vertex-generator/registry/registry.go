package registry

import "github.com/EverythingMe/vertex/swagger"

type Generator interface {
	Generate(*swagger.API) ([]byte, error)
}

var registry = map[string]Generator{}

func RegisterGenerator(name string, g Generator) {
	registry[name] = g
}

func Get(name string) (Generator, bool) {
	g, f := registry[name]
	return g, f
}
