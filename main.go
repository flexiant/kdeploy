package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/mafraba/digger"
	"gopkg.in/yaml.v2"
)

func main() {
	// tempDir, _ := filepath.Abs("./tmp")
	// os.Mkdir(tempDir, 0666)

	metadata := parseMetadata()
	defaults, err := metadata.AttributeDefaults()
	if err != nil {
		panic(err)
	}

	roleList := `
	{
		"rc" : {
			"frontend" : {
				"number" : 3,
				"container": {
					"name": "contname",
					"image": "somerepo/someimage",
					"version": "latest",
					"port": 8080
				}
			}
		}
	}
	`
	roleListDigger, err := digger.NewJSONDigger([]byte(roleList))
	if err != nil {
		panic(err)
	}
	attributes, _ := digger.NewMultiDigger(
		roleListDigger,
		defaults,
	)

	tmpl := parseTemplate("./samples/frontend-controller.yaml", attributes)

	// Run the template to verify the output.
	err = tmpl.Execute(os.Stdout, nil)
	if err != nil {
		log.Fatalf("error executing template: %s", err)
	}

}

func parseTemplate(templatePath string, attributes digger.Digger) *template.Template {
	templateFile, _ := filepath.Abs(templatePath)
	templateContent, err := ioutil.ReadFile(templateFile)
	if err != nil {
		panic(err)
	}
	// First we create a FuncMap with which to register the function.
	funcMap := template.FuncMap{
		// Associating functions with aliases
		"getString": attributes.GetString,
		"getNumber": attributes.GetNumber,
		"getBool":   attributes.GetBool,
	}
	// Create a template, add the function map, and parse the text.
	tmpl, err := template.New("templateTest").Funcs(funcMap).Parse(string(templateContent))
	if err != nil {
		log.Fatalf("error parsing: %s", err)
	}
	return tmpl
}

func parseMetadata() Metadata {
	metadataFile, _ := filepath.Abs("./samples/metadata.yaml")
	metadataContent, err := ioutil.ReadFile(metadataFile)
	if err != nil {
		panic(err)
	}
	var metadata Metadata
	err = yaml.Unmarshal(metadataContent, &metadata)
	if err != nil {
		panic(err)
	}
	return metadata
}
