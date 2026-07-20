// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package gatherers

import (
	"errors"
	"fmt"
	"time"

	"github.com/clbanning/mxj/v2"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

//nolint:gochecknoinits
func init() {
	mxj.PrependAttrWithHyphen(false)
}

func parseXMLToFactValueMap(
	xmlContent []byte,
	elementsToList map[string]bool,
	factValueOpts ...entities.FactValueOption) (*entities.FactValueMap, error) {
	mv, err := mxj.NewMapXml(xmlContent)
	if err != nil {
		return nil, err
	}

	mapValue := map[string]any(mv)
	updatedMap := convertListElements(mapValue, elementsToList)

	factValue, err := entities.NewFactValue(updatedMap, factValueOpts...)
	if err != nil {
		return nil, err
	}

	factValueMap, ok := factValue.(*entities.FactValueMap)
	if !ok {
		return nil, errors.New("error converting to FactValueMap")
	}

	return factValueMap, nil
}

// convertListElements returns and updated version of the current xml map, where the keys in elementsToList
// are converted to lists if they are unique entries. This is required due how xml works, as the initial
// xml to map conversion cannot know if the coming entry is an unique element or not.
func convertListElements(currentMap map[string]any, elementsToList map[string]bool) map[string]any {
	convertedMap := make(map[string]any)

	for key, value := range currentMap {
		switch assertedValue := value.(type) {
		case map[string]any:
			{
				convertedMap[key] = convertListElements(assertedValue, elementsToList)
			}
		// Item is a list, so each element must be treated to see if its children need a conversion.
		// The items could be maps itself, so they must recurse yet again
		case []any:
			{
				newList := []any{}

				for _, item := range assertedValue {
					if toMap, ok := item.(map[string]any); ok {
						newList = append(newList, convertListElements(toMap, elementsToList))
					} else {
						newList = append(newList, item)
					}
				}

				convertedMap[key] = newList
			}
		default:
			{
				// If the current key is a duration, convert it to seconds float
				duration, err := time.ParseDuration(fmt.Sprint(assertedValue))
				if err == nil {
					convertedMap[key] = duration.Seconds()
				} else {
					convertedMap[key] = assertedValue
				}
			}
		}

		// If the current key is not a list and it is included in the elementsToList, initialize as list
		_, isList := convertedMap[key].([]any)
		if elementsToList[key] && !isList {
			convertedMap[key] = []any{convertedMap[key]}
		}
	}

	return convertedMap
}
