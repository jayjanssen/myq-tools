package viewer

import (
	_ "embed"
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

var (
	ViewNames []string
	Views     map[string]View
)

//go:embed view_defaults.yaml
var defaultViewYaml string

func LoadDefaultViews() error {
	return ParseViews(defaultViewYaml)
}

// Get the name Viewer, or return an error
func GetViewer(name string) (StateViewer, error) {
	view, ok := Views[name]
	if !ok {
		return nil, fmt.Errorf("view %s not found", name)
	} else {
		return view, nil
	}
}

func ParseViews(yaml_str string) error {
	var views []View
	err := yaml.Unmarshal([]byte(yaml_str), &views)
	if err != nil {
		return err
	}

	Views = make(map[string]View)

	// construct the Views map
	for _, view := range views {
		ViewNames = append(ViewNames, view.Name)
		Views[view.Name] = view
	}
	sort.Strings(ViewNames)
	return nil
}
