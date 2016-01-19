package kdeploy_metadata

import (
	"fmt"

	"github.com/mafraba/digger"
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
	Name        string
	Maintainer  string
	Email       string
	Description string
	Version     string
	Attributes  AttributesMetadata
	Rc          map[string]string
	Svc         map[string]string
}

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
