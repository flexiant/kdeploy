package resolver

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"path/filepath"

	"github.com/flexiant/digger"
)

// ResolveTemplate resolves a template according to the attributes passed
func ResolveTemplate(templatePath string, attributes digger.Digger) (string, error) {
	templateFile, err := filepath.Abs(templatePath)
	if err != nil {
		return "", err
	}
	templateContent, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return "", err
	}
	// associate functions with aliases
	funcMap := template.FuncMap{
		"attr": attributes.Get,
	}
	// create a template, add the function map, and parse the text.
	tmpl, err := template.New("templateTest").Funcs(funcMap).Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("error parsing template %s: %s", templatePath, err)
	}
	// execute the template
	var doc bytes.Buffer
	err = tmpl.Execute(&doc, nil)
	if err != nil {
		panic("error executing template: " + err.Error())
	}
	return doc.String(), nil
}
