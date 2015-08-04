package schema

import (
	"bytes"
	"fmt"
	"text/template"
)

func render(templateString string, sc string) ([]byte, error) {
	//	tpl, err := template.New("schema").Funcs(template.FuncMap{
	//		"getDefault": preprocessDefault,
	//	}).Parse(templateString)

	tpl, err := template.New("schema").Parse(templateString)
	if err != nil {
		return nil, fmt.Errorf("Could not parse template: %s", err)
	}

	buf := bytes.NewBuffer(nil)

	if err = tpl.Execute(buf, sc); err != nil {
		return nil, fmt.Errorf("Could not execute template: %s", err)
	}

	return buf.Bytes(), nil

}
