package template

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/flexiant/digger"
	"github.com/flexiant/kdeploy/utils"
	ymlutil "github.com/ghodss/yaml"
)

// SingleAttributeMetadata holds metadata for a configuration attribute
type SingleAttributeMetadata struct {
	Description string      // Description of the attribute
	Default     interface{} // Default value for the attribute: could be string, number, or bool
	Required    bool        // Required or not
}

// AttributesMetadata holds a whole collection of MetadataAttribute organized into three levels:
// <resource-type>/<resource-name>/<attribute-name> (e.g. "svc/frontend/balancer")
type AttributesMetadata map[string]map[string]map[string]SingleAttributeMetadata

// Metadata holds generic info about the deployment's resources and attributes
type Metadata struct {
	Name                   string
	Maintainer             string
	Email                  string
	Description            string
	Version                string
	Attributes             AttributesMetadata
	ReplicationControllers map[string]string
	Services               map[string]string
	path                   string
}

// ParseMetadata parses the metadata file in the kube dir
func ParseMetadata(path string) Metadata {
	absPath, err := filepath.Abs(path)
	utils.CheckError(err)

	metadataFile := fmt.Sprintf("%s/metadata.yaml", filepath.Clean(absPath))
	utils.CheckError(err)

	metadataContent, err := ioutil.ReadFile(metadataFile)
	utils.CheckError(err)

	var metadata Metadata
	err = yaml.Unmarshal(metadataContent, &metadata)
	utils.CheckError(err)

	metadata.path = filepath.Dir(metadataFile)
	return metadata
}

// RequiredAttributes returns default values for attributes
func (m Metadata) RequiredAttributes() ([]string, error) {
	var reqs = []string{}
	for resourceType, resources := range m.Attributes {
		for resourceName, attributes := range resources {
			for attributeName, attrMetadata := range attributes {
				if attrMetadata.Required {
					reqs = append(reqs, fmt.Sprintf("%s/%s/%s", resourceType, resourceName, attributeName))
				}
			}
		}
	}
	return reqs, nil
}

// CheckRequiredAttributes returns an error if some required attribute is missing
func (m Metadata) CheckRequiredAttributes(attributes digger.Digger) error {
	reqs, err := m.RequiredAttributes()
	if err != nil {
		return fmt.Errorf("could not calculate required attributes: %v", err)
	}
	for _, att := range reqs {
		_, err := attributes.Get(att)
		if err != nil {
			return fmt.Errorf("required attribute not present: '%s'", att)
		}
	}
	return nil
}

// AttributeDefaults returns default values for attributes
func (m Metadata) AttributeDefaults() (digger.Digger, error) {
	defaults, err := defaultsMapFromMetadata(m.Attributes)
	if err != nil {
		return nil, fmt.Errorf("could not build defaults: %v", err)
	}
	digger, err := digger.NewMapDigger(defaults)
	if err != nil {
		return nil, fmt.Errorf("could not build digger: %v", err)
	}
	return digger, nil
}

func defaultsMapFromMetadata(md AttributesMetadata) (interface{}, error) {
	defaults := make(map[string]interface{})

	for resourceType, resources := range md {
		for resourceName, attributes := range resources {
			for attributeName, attrMetadata := range attributes {
				// take default value for the attribute
				val := attrMetadata.Default
				// check if resourceType already in map
				if defaults[resourceType] == nil {
					defaults[resourceType] = make(map[string]interface{})
				}
				// check if resource already in map
				if defaults[resourceType].(map[string]interface{})[resourceName] == nil {
					defaults[resourceType].(map[string]interface{})[resourceName] = make(map[string]interface{})
				}
				defaults[resourceType].(map[string]interface{})[resourceName].(map[string]interface{})[attributeName] = val
			}
		}
	}

	return defaults, nil
}

// ParseServices parses the service templates in the kube and returns their JSON representations
func (m Metadata) ParseServices(attributes digger.Digger) ([]string, error) {
	err := m.CheckRequiredAttributes(attributes)
	if err != nil {
		return nil, err
	}
	return parseTemplates(m.path, m.Services, attributes)
}

// ParseControllers parses the replication controllers in the kube and returns their JSON representations
func (m Metadata) ParseControllers(attributes digger.Digger) ([]string, error) {
	err := m.CheckRequiredAttributes(attributes)
	if err != nil {
		return nil, err
	}
	return parseTemplates(m.path, m.ReplicationControllers, attributes)
}

func parseTemplates(path string, templates map[string]string, attributes digger.Digger) ([]string, error) {
	var specs = []string{}
	for _, templateFile := range templates {
		specYaml, err := ResolveTemplate(fmt.Sprintf("%s/%s", path, templateFile), attributes)
		if err != nil {
			return nil, fmt.Errorf("error resolving template %s: %v", templateFile, err)
		}
		specJSON, err := ymlutil.YAMLToJSON([]byte(specYaml))
		if err != nil {
			log.Printf("yaml:\n%s", specYaml)
			return nil, fmt.Errorf("error converting template %s: %v", templateFile, err)
		}
		specs = append(specs, string(specJSON))
	}
	return specs, nil
}
