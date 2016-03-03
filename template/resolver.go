package template

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/cbroglie/mustache"
	"github.com/flexiant/kdeploy/utils"
	"github.com/mafraba/deeply"
)

// ResolveTemplate resolves a template according to the attributes passed
func ResolveTemplate(templatePath string, attributes map[string]interface{}) (string, error) {
	templateFile, err := filepath.Abs(templatePath)
	if err != nil {
		return "", err
	}
	output, err := mustache.RenderFile(templateFile, attributes)
	if err != nil {
		return "", fmt.Errorf("error parsing template %s: %v", templatePath, err)
	}
	return output, nil
}

func BuildAttributes(filePath string, defaults map[string]interface{}) map[string]interface{} {
	// just defaults if no attributes given
	if filePath == "" {
		return defaults
	}

	attJSON, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("could not read file '%s' (%v)", filePath, err)
	}

	var attribs map[string]interface{}
	err = json.Unmarshal(attJSON, &attribs)
	utils.CheckError(err)

	attributes := deeply.Merge(defaults, attribs)

	return attributes
}
