package template

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/flexiant/digger"
	"github.com/flexiant/kdeploy/utils"
	"gopkg.in/yaml.v2"
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
	ReplicationControllers map[string]string `yaml:"rc"`
	Services               map[string]string `yaml:"svc"`
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
func (m Metadata) CheckRequiredAttributes(attributes map[string]interface{}) error {
	reqs, err := m.RequiredAttributes()
	if err != nil {
		return fmt.Errorf("could not calculate required attributes: %v", err)
	}
	digger, err := digger.NewMapDigger(attributes)
	if err != nil {
		return fmt.Errorf("could not build digger: %v", err)
	}
	for _, att := range reqs {
		_, err := digger.Get(att)
		if err != nil {
			return fmt.Errorf("required attribute not present: '%s'", att)
		}
	}
	return nil
}

// AttributeDefaults returns default values for attributes
func (m Metadata) AttributeDefaults() (map[string]interface{}, error) {
	defaults, err := defaultsMapFromMetadata(m.Attributes)
	if err != nil {
		return nil, fmt.Errorf("could not build defaults: %v", err)
	}
	return defaults, nil
}

func defaultsMapFromMetadata(md AttributesMetadata) (map[string]interface{}, error) {
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
func (m Metadata) ParseServices(attributes map[string]interface{}) (map[string]string, error) {
	err := m.CheckRequiredAttributes(attributes)
	if err != nil {
		return nil, err
	}
	specMap, err := m.parseTemplates(m.Services, attributes)
	if err != nil {
		return nil, err
	}
	return marshalMapValues(specMap)
}

func marshalMapValues(m map[string]interface{}) (map[string]string, error) {
	specs := map[string]string{}
	for k, v := range m {
		specJSON, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("error marshalling into json (%s): %v", k, err)
		}
		specs[k] = string(specJSON)
	}
	return specs, nil
}

// ParseControllers parses the replication controllers in the kube and returns their JSON representations
func (m Metadata) ParseControllers(attributes map[string]interface{}) (map[string]string, error) {
	err := m.CheckRequiredAttributes(attributes)
	if err != nil {
		return nil, err
	}
	specMap, err := m.parseTemplates(m.ReplicationControllers, attributes)
	if err != nil {
		return nil, err
	}
	err = setKubeVersionOnRCs(specMap)
	if err != nil {
		return nil, err
	}
	return marshalMapValues(specMap)
}

func (m Metadata) parseTemplates(templates map[string]string, attributes map[string]interface{}) (map[string]interface{}, error) {
	var specs = map[string]interface{}{}
	for specName, templateFile := range templates {
		log.Debugf("Going to parse %s/%s", m.path, templateFile)
		specMap, err := parseTemplate(fmt.Sprintf("%s/%s", m.path, templateFile), attributes)
		if err != nil {
			return nil, fmt.Errorf("error parsing template %s: %v", templateFile, err)
		}
		s, err := digger.NewMapDigger(specMap)
		if err != nil {
			return nil, fmt.Errorf("error parsing template %s: %v", templateFile, err)
		}
		name, err := s.GetString("metadata/name")
		if err != nil {
			return nil, fmt.Errorf("error parsing template %s: %v", templateFile, err)
		}
		if name != specName {
			return nil, fmt.Errorf("non matching resource name in %s", templateFile)
		}
		err = addKubewareLabel(m.Name, m.Version, specMap)
		if err != nil {
			return nil, fmt.Errorf("error adding kubeware labels to %s: %v", templateFile, err)
		}
		specs[specName] = specMap
	}
	return specs, nil
}

func parseTemplate(templateFile string, attributes map[string]interface{}) (map[string]interface{}, error) {
	specYaml, err := ResolveTemplate(templateFile, attributes)
	if err != nil {
		return nil, fmt.Errorf("error resolving template %s: %v", templateFile, err)
	}
	var specMap map[interface{}]interface{}
	err = yaml.Unmarshal([]byte(specYaml), &specMap)
	if err != nil {
		return nil, fmt.Errorf("error parsing yaml for %s: %v", templateFile, err)
	}
	normalizedMap, err := normalizeValue(specMap)
	if err != nil {
		return nil, fmt.Errorf("error normalizing yaml for %s: %v", templateFile, err)
	}
	return normalizedMap.(map[string]interface{}), nil
}

func addKubewareLabel(name, version string, specmap map[string]interface{}) error {
	metadata := specmap["metadata"].(map[string]interface{})
	if metadata["labels"] != nil {
		labels := metadata["labels"].(map[string]interface{})
		labels["kubeware"] = name
		labels["kubeware-version"] = version
	} else {
		metadata["labels"] = map[string]string{
			"kubeware":         name,
			"kubeware-version": version,
		}
	}

	return nil
}

func normalizeValue(value interface{}) (interface{}, error) {
	switch value := value.(type) {
	case map[interface{}]interface{}:
		node := make(map[string]interface{}, len(value))
		for k, v := range value {
			key, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("Unsupported map key: %#v", k)
			}
			item, err := normalizeValue(v)
			if err != nil {
				return nil, fmt.Errorf("Unsupported map value: %#v (%v)", v, err)
			}
			node[key] = item
		}
		return node, nil
	case map[string]interface{}:
		node := make(map[string]interface{}, len(value))
		for key, v := range value {
			item, err := normalizeValue(v)
			if err != nil {
				return nil, fmt.Errorf("Unsupported map value: %#v", v)
			}
			node[key] = item
		}
		return node, nil
	case []interface{}:
		node := make([]interface{}, len(value))
		for key, v := range value {
			item, err := normalizeValue(v)
			if err != nil {
				return nil, fmt.Errorf("Unsupported list item: %#v", v)
			}
			node[key] = item
		}
		return node, nil
	case bool, float64, int, string:
		return value, nil
	}
	return nil, fmt.Errorf("Unsupported type: %T", value)
}

func setKubeVersionOnRCs(rcs map[string]interface{}) error {
	for name, rc := range rcs {
		err := setKubeVersionOnRC(rc.(map[string]interface{}))
		if err != nil {
			return err
		}
		rcs[name] = rc
	}
	return nil
}

func setKubeVersionOnRC(rc map[string]interface{}) error {
	kv, err := extractKubeVersion(rc)
	if err != nil {
		return err
	}
	// set at pod template
	path := []string{"spec", "template", "metadata", "labels"}
	m := rc
	for _, s := range path {
		if m[s] == nil {
			m[s] = map[string]interface{}{}
		}
		m = m[s].(map[string]interface{})
	}
	m["kubeware"] = kv
	// set at label selector
	path = []string{"spec", "selector"}
	m = rc
	for _, s := range path {
		if m[s] == nil {
			m[s] = map[string]interface{}{}
		}
		m = m[s].(map[string]interface{})
	}
	m["kubeware"] = kv

	return nil
}

func extractKubeVersion(rc map[string]interface{}) (string, error) {
	d, err := digger.NewMapDigger(rc)
	if err != nil {
		return "", err
	}
	k, err := d.GetString("metadata/labels/kubeware")
	if err != nil {
		return "", err
	}
	v, err := d.GetString("metadata/labels/kubeware-version")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s", k, v), nil
}
