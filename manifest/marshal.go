package manifest

import (
	"encoding/json"
	"github.com/pkg/errors"
)

// Marshals the given manifest with correct indentation and also removes empty `"apply"` data.
func Marshal(manifest Manifest) ([]byte, error) {
	// First marshal the given manifest
	rawBuf, err := json.Marshal(manifest)
	if err != nil {
		return nil, errors.Wrap(err, "unable to serialize manifest")
	}

	// Now run the cleanup to remove unwanted elements
	cleanedUpBuf, err := marshalCleanup(rawBuf)
	if err != nil {
		return nil, errors.Wrap(err, "unable to cleanup JSON data")
	}

	// Marshal the cleaned data again with the correct indentation
	buf, err := json.MarshalIndent(cleanedUpBuf, "", "  ")
	if err != nil {
		return nil, errors.Wrap(err, "unable to deserialize manifest")
	}

	return buf, nil
}

func marshalCleanup(buf []byte) (map[string]interface{}, error) {
	// Unmarshal the given data into a map structure so we can remove unwanted stuff
	var tempValue map[string]interface{}

	if err := json.Unmarshal(buf, &tempValue); err != nil {
		return tempValue, errors.Wrap(err, "unable to read serialized manifest")
	}

	for key, value := range tempValue {
		if key == "modules" {
			marshalCleanupModuleData(value.([]interface{}))
		}
	}

	return tempValue, nil
}

func marshalCleanupModuleData(modules []interface{}) {
	// Cleanup the modules and submodules so that there are no `"apply": {}` left in the JSON structure
	for _, m := range modules {
		module := m.(map[string]interface{})
		for k, v := range module {
			if k == "apply" {
				if len(v.(map[string]interface{})) == 0 {
					delete(module, "apply")
				}
			}
			// Recurse into submodules
			if k == "submodules" {
				marshalCleanupModuleData(v.([]interface{}))
			}
		}
	}
}
