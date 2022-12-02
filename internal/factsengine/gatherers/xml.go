package gatherers

import (
	"fmt"

	"github.com/clbanning/mxj/v2"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

func init() {
	mxj.PrependAttrWithHyphen(false)
}

func parseXMLToFactValueMap(xmlContent []byte, elementsToList []string) (*entities.FactValueMap, error) {
	mv, err := mxj.NewMapXml(xmlContent)
	if err != nil {
		return nil, err
	}

	for _, element := range elementsToList {
		err = convertList(&mv, element)
		if err != nil {
			return nil, fmt.Errorf("error converting %s to list", element)
		}
	}

	mapValue := map[string]interface{}(mv)
	factValue := entities.ParseInterfaceFactValue(mapValue)
	factValueMap, ok := factValue.(*entities.FactValueMap)
	if !ok {
		return nil, fmt.Errorf("error converting to FactValueMap")
	}

	return factValueMap, nil
}

// convertList converts given keys to list if only one value was present
// this is needed as many fields are lists even though they might have
// one element
func convertList(mv *mxj.Map, key string) error {
	paths := mv.PathsForKey(key)
	for _, path := range paths {
		value, err := mv.ValuesForPath(path)
		if err != nil {
			return err
		}

		values := map[string]interface{}{
			key: value,
		}

		_, err = mv.UpdateValuesForPath(values, path)
		if err != nil {
			return err
		}
	}

	return nil
}
