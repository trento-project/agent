package gatherers

import (
	"fmt"
	"time"

	"github.com/clbanning/mxj/v2"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

func init() {
	mxj.PrependAttrWithHyphen(false)
}

func parseXMLToFactValueMap(xmlContent []byte, elementsToList map[string]bool) (*entities.FactValueMap, error) {
	mv, err := mxj.NewMapXml(xmlContent)
	if err != nil {
		return nil, err
	}

	mapValue := map[string]interface{}(mv)
	updatedMap := convertListElements(mapValue, elementsToList)
	factValue, err := entities.NewFactValue(updatedMap, entities.WithStringConversion())
	if err != nil {
		return nil, err
	}

	factValueMap, ok := factValue.(*entities.FactValueMap)
	if !ok {
		return nil, fmt.Errorf("error converting to FactValueMap")
	}

	return factValueMap, nil
}

// convertListElements returns and updated version of the current xml map, where the keys in elementsToList
// are converted to lists if they are unique entries. This is required due how xml works, as the initial
// xml to map conversion cannot know if the coming entry is an unique element or not
func convertListElements(currentMap map[string]interface{}, elementsToList map[string]bool) map[string]interface{} {
	convertedMap := make(map[string]interface{})
	for key, value := range currentMap {
		switch assertedValue := value.(type) {
		case map[string]interface{}:
			{
				convertedMap[key] = convertListElements(assertedValue, elementsToList)
			}
		// Item is a list, so each element must be treated to see if its children need a conversion.
		// The items could be maps itself, so they must recurse yet again
		case []interface{}:
			{
				newList := []interface{}{}
				for _, item := range assertedValue {
					if toMap, ok := item.(map[string]interface{}); ok {
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
				if duration, err := time.ParseDuration(fmt.Sprint(assertedValue)); err == nil {
					convertedMap[key] = duration.Seconds()
				} else {
					convertedMap[key] = assertedValue
				}
			}
		}

		// If the current key is not a list and it is included in the elementsToList, initialize as list
		_, isList := convertedMap[key].([]interface{})
		if elementsToList[key] && !isList {
			convertedMap[key] = []interface{}{convertedMap[key]}
		}
	}

	return convertedMap
}
