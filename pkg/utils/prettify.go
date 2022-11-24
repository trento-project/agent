package utils

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"
)

func PrettifyInterfaceToJSON(data interface{}) (string, error) {
	jsonResult, err := json.Marshal(data)
	if err != nil {
		return "", errors.Wrap(err, "Error building the response")
	}

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, jsonResult, "", "  "); err != nil {
		return "", errors.Wrap(err, "Error indenting the json data")
	}

	return prettyJSON.String(), nil
}
