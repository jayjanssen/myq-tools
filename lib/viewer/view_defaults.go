package viewer

import (
	"embed"
	"fmt"
	"io/fs"

	"gopkg.in/yaml.v3"
)

var (
	viewNames []string
	views     map[string]View
)

//go:embed views/*.yaml
var viewFiles embed.FS

// Load the default views from the embedded files
func LoadDefaultViews() error {
	// get the list of files
	fileNames, err := fs.Glob(viewFiles, "views/*.yaml")
	if err != nil {
		return err
	}

	// read and parse each file and add it to the Views map
	views = make(map[string]View)
	for _, fileName := range fileNames {
		bytes, err := fs.ReadFile(viewFiles, fileName)
		if err != nil {
			return err
		}

		// Each file could have multiple views
		var parsedViews []View
		err = yaml.Unmarshal(bytes, &parsedViews)
		if err != nil {
			return err
		}

		// Add the parsed views to the global map
		for _, view := range parsedViews {
			viewNames = append(viewNames, view.Name)
			views[view.Name] = view
		}
	}

	return nil
}

// List the names of all the Views
func ListViews() []string {
	return viewNames
}

// Get the named Viewer, or return an error
func GetViewer(name string) (Viewer, error) {
	view, ok := views[name]
	if !ok {
		return nil, fmt.Errorf("view %s not found", name)
	} else {
		return view, nil
	}
}
