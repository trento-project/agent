package utils

import (
	"regexp"
	"strings"
)

// FindMatches finds regular expression matches in a key/value based
// text (ini files, for example), and returns a map with them.
// If the matched key has spaces, they will be replaced with underscores
// If the same keys is found multiple times, the entry of the map will
// have a list as value with all of the matched values
// The pattern must have 2 groups. For example: `(.+)=(.*)`
func FindMatches(pattern string, text []byte) map[string]interface{} {
	configMap := make(map[string]interface{})

	r := regexp.MustCompile(pattern)
	values := r.FindAllStringSubmatch(string(text), -1)
	for _, match := range values {
		key := strings.Replace(match[1], " ", "_", -1) //nolint
		if _, ok := configMap[key]; ok {
			switch configMap[key].(type) { //nolint
			case string:
				configMap[key] = []interface{}{configMap[key]}
			}
			configMap[key] = append(configMap[key].([]interface{}), match[2]) //nolint
		} else {
			configMap[key] = match[2]
		}
	}
	return configMap
}

func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
