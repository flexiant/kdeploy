package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

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

	roleList := `{"rc" : { "frontend" : { "number" : 3 }}}`
	roleListDigger, _ := digger.NewJSONDigger([]byte(roleList))
	attributes, _ := digger.NewMultiDigger(
		roleListDigger,
		defaults,
	)

	path := "rc/frontend/number"
	val, _ := attributes.GetNumber(path)
	fmt.Printf("%s -> %v", path, val)
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
