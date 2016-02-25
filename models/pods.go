package models

import "encoding/json"

type Pod struct {
	Metadata struct {
		CreationTimestamp string
		Name              string
		Labels            map[string]string
		Namespace         string
		ResourceVersion   string
	}
	Status struct {
		Conditions []struct {
			Status string
			Type   string
		}
	}
}

// TODO: we should probably just return a slice instead of a pointer, since we are already
// signaling errors with the error object returned, no need to return nil
func NewPodsJSON(jsonStr string) (*[]Pod, error) {
	type PodsList struct {
		Items []Pod
	}
	var pl PodsList
	err := json.Unmarshal([]byte(jsonStr), &pl)
	if err != nil {
		return nil, err
	}
	return &pl.Items, nil
}
