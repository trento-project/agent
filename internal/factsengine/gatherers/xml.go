package gatherers

import (
	"fmt"

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
	factValue, err := entities.NewFactValue(updatedMap)
	if err != nil {
		return nil, err
	}

	factValueMap, ok := factValue.(*entities.FactValueMap)
	if !ok {
		return nil, fmt.Errorf("error converting to FactValueMap")
	}

	return factValueMap, nil
}

func convertListElements(currentMap map[string]interface{}, elementsToList map[string]bool) map[string]interface{} {
	convertedMap := make(map[string]interface{})
	for key, value := range currentMap {
		switch assertedValue := value.(type) {
		case map[string]interface{}:
			convertedMap[key] = convertListElements(assertedValue, elementsToList)
		case []interface{}:
			newList := []interface{}{}
			for _, item := range assertedValue {
				if toMap, ok := item.(map[string]interface{}); ok {
					newList = append(newList, convertListElements(toMap, elementsToList))
				} else {
					newList = append(newList, item)
				}
			}
			convertedMap[key] = newList
		default:
			convertedMap[key] = assertedValue
		}

		_, isList := convertedMap[key].([]interface{})
		if elementsToList[key] && !isList {
			convertedMap[key] = []interface{}{convertedMap[key]}
		}
	}

	return convertedMap
}
