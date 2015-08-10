package java

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"

	"github.com/alecthomas/jsonschema"
	"gitlab.doit9.com/server/vertex/swagger"
	"gitlab.doit9.com/server/vertex/vertex-generator/registry"
)

// Generator is a code geneator that renders a complete Java client API code from a swagger definition of
// our API
type Generator struct {
	substitutions map[string]string
}

// NewGenerator creates a new generator and parses the command line arguments
func NewGenerator() *Generator {
	return &Generator{
		substitutions: map[string]string{},
	}
}

var extendstring string
var extendfile string
var pkg string = "com.example.foo"

func init() {

	flag.StringVar(&extendstring, "java.extend", "", "A comma separated list of class:extends. for extending instead of generating classes")
	flag.StringVar(&extendfile, "java.extendfile", "", "Path to a config YAML containing extention rules")
	flag.StringVar(&pkg, "java.package", "com.example.foo", "The package the generated API belongs to")

	g := NewGenerator()
	registry.RegisterGenerator("java", g)
}

func (g *Generator) readExtendFile(extendfile string) error {
	fp, err := os.Open(extendfile)
	if err != nil {
		return err
	}
	defer fp.Close()

	b, err := ioutil.ReadAll(fp)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, &g.substitutions)
}
func (g *Generator) readSubstitutions() error {

	if len(g.substitutions) > 0 {
		return nil
	}

	if extendfile != "" {
		return g.readExtendFile(extendfile)

	} else if extendstring != "" {

		kvs := strings.Split(extendstring, ",")
		for _, kv := range kvs {
			parts := strings.Split(kv, ":")
			if len(parts) != 2 {
				log.Println("Error parsing replacement, invalid format '%s'", kv)
				continue
			}
			fmt.Printf("Substituion: '%s' => '%s'\n", parts[0], parts[1])
			g.substitutions[parts[0]] = parts[1]
		}

	}

	return nil

}

// newJavaClass creates a new java client sub-class definition, for a method's return type
func (g *Generator) newJavaClass(name string, t *jsonschema.Type) Class {

	ret := Class{
		Name:    name,
		Members: []Member{},
	}

	if sub, found := g.substitutions[name]; found {
		ret.Extends = sub
	}

	for k, prop := range t.Properties {
		ret.Members = append(ret.Members, Member{
			Name: k,
			Type: newTypeRef(prop),
		})
	}
	return ret
}

// formatMethodName converts a path and verb of a method to a legal java method name.
//
// e.g. /User/byId with GET will be converted to getUserById
func formatMethodName(path, verb string) string {

	parts := strings.Split(path, "/")
	for i := range parts {
		parts[i] = strings.Title(cleanRe.ReplaceAllString(strings.TrimSpace(parts[i]), ""))
	}

	return strings.ToLower(verb) + strings.Join(parts, "")
}

// newJavaMathod creates a new method definition based on a swagger method definition and a return value
func (g *Generator) newJavaMethod(pth, verb string, method swagger.Method) Method {

	ret := Method{
		Name:     formatMethodName(pth, verb),
		Returns:  newTypeRef(method.Responses["default"].Schema.Type),
		Params:   make([]Param, 0, len(method.Parameters)),
		HttpVerb: strings.ToUpper(verb),
		Path:     pth,
		Doc:      method.Description,
	}

	for _, param := range method.Parameters {
		var jparm Param
		if param.Ref == "" {
			jparm = Param{
				Name:     param.Name,
				Type:     newTypeRefSwagger(param.Type, param.Items),
				Doc:      param.Description,
				In:       param.In,
				Required: param.Required,
			}
		} else {
			_, ref := path.Split(param.Ref)
			jparm = Param{
				Name:   ref,
				Global: true,
			}

		}

		ret.Params = append(ret.Params, jparm)

	}
	return ret
}

// Generate takes a swagger API and compiles a java client from it
func (g *Generator) Generate(swapi *swagger.API) ([]byte, error) {
	if err := g.readSubstitutions(); err != nil {
		return nil, err
	}
	japi := g.newJavaAPI(swapi)

	return g.render(tpl, &japi)

}

// render takes a java API definition and the template string and renders the complete generated code
func (g *Generator) render(templateString string, api *API) ([]byte, error) {

	templateString = strings.Replace(
		strings.Replace(
			strings.Replace(
				strings.Replace(templateString, "\\\n", "", -1),
				"\\n", "\n", -1),
			"~", "`", -1),
		"\\ ", "", -1)

	tpl, err := template.New("schema").Funcs(template.FuncMap{
		"renderArguments": renderArguments,
	}).Parse(templateString)
	if err != nil {
		return nil, fmt.Errorf("Could not parse template: %s", err)
	}

	buf := bytes.NewBuffer(nil)

	if err = tpl.Execute(buf, api); err != nil {
		return nil, fmt.Errorf("Could not execute template: %s", err)
	}

	return buf.Bytes(), nil

}

var cleanRe = regexp.MustCompile("[^[:alnum:]]")

// newJavaAPI creates the entire JavaAPI object from a swagger API definition
func (g *Generator) newJavaAPI(swapi *swagger.API) API {

	api := API{
		Name:    cleanRe.ReplaceAllString(swapi.Info.Title, ""),
		Package: pkg,
		Doc:     swapi.Info.Description,
		Root:    swapi.Basepath,
		Types:   make([]Class, 0, len(swapi.Definitions)),
		Methods: make([]Method, 0, len(swapi.Paths)),
		Globals: make([]Param, 0),
	}

	for name, tp := range swapi.Definitions {
		api.Types = append(api.Types, g.newJavaClass(name, tp.Type))
	}

	for path, methods := range swapi.Paths {
		for verb, method := range methods {

			api.Methods = append(api.Methods, g.newJavaMethod(path, verb, method))
		}
	}

	for _, param := range swapi.Parameters {

		jparm := Param{
			Name: param.Name,
			Type: newTypeRefSwagger(param.Type, param.Items),
			Doc:  param.Description,
			In:   param.In,
		}
		api.Globals = append(api.Globals, jparm)
	}

	return api
}

// renderArguments formats the input arguments of a java method based on its definition
func renderArguments(args []Param) string {
	if args == nil || len(args) == 0 {
		return ""
	}

	argstrs := make([]string, 0, len(args))
	for _, arg := range args {
		if !arg.Global && arg.In != "header" {
			argstrs = append(argstrs, fmt.Sprintf("%s %s", arg.Type, arg.Name))
		}
	}

	return strings.Join(argstrs, ", ")
}
