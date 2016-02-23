package template

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/flexiant/digger"
	"github.com/flexiant/kdeploy/utils"
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
	utils.CheckError(err)

	return doc.String(), nil
}

func BuildAttributes(filePath string, defaults digger.Digger) digger.Digger {
	roleList, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("could not read file '%s' (%v)", filePath, err)
	}

	roleListDigger, err := digger.NewJSONDigger([]byte(roleList))
	utils.CheckError(err)

	attributes, err := digger.NewMultiDigger(
		roleListDigger,
		defaults,
	)
	utils.CheckError(err)

	return attributes
}
