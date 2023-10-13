package gatherers

import (
	"fmt"
	"sort"
	"strings"
)

type GathererNotFoundError struct {
	Name string
}

func (e *GathererNotFoundError) Error() string {
	return fmt.Sprintf("gatherer %s not found", e.Name)
}

// map[gathererName]map[GathererVersion]Gatherer
type Tree map[string]map[string]FactGatherer

type Registry struct {
	gatherers Tree
}

func NewRegistry(gatherers Tree) *Registry {
	return &Registry{
		gatherers: gatherers,
	}
}

func (m *Registry) GetGatherer(name string) (FactGatherer, error) {
	gathererName, version, err := extractVersionAndGathererName(name)
	if err != nil {
		return nil, err
	}
	if version == "" {
		latestVersion, err := m.getLatestVersionForGatherer(name)
		if err != nil {
			return nil, err
		}
		version = latestVersion
	}

	if g, found := m.gatherers[gathererName][version]; found {
		return g, nil
	}
	return nil, &GathererNotFoundError{Name: name}
}

func (m *Registry) AvailableGatherers() []string {
	gatherersList := []string{}

	for gatherer, versions := range m.gatherers {
		gathererVersions := []string{}
		for v := range versions {
			gathererVersions = append(gathererVersions, v)
		}
		gatherersList = append(
			gatherersList,
			fmt.Sprintf("%s - %s", gatherer, strings.Join(gathererVersions, "/")),
		)
	}

	return gatherersList
}

// This is not safe, please not use concurrently.
func (m *Registry) AddGatherers(gatherers Tree) {
	maps := []Tree{m.gatherers, gatherers}
	result := make(Tree)

	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	m.gatherers = result
}

func (m *Registry) getLatestVersionForGatherer(name string) (string, error) {
	availableGatherers, found := m.gatherers[name]
	if !found {
		return "", &GathererNotFoundError{Name: name}
	}
	versions := []string{}
	for v := range availableGatherers {
		versions = append(versions, v)
	}

	sort.Strings(versions)

	return versions[len(versions)-1], nil
}

func extractVersionAndGathererName(gathererName string) (string, string, error) {
	parts := strings.Split(gathererName, "@")
	if len(parts) == 1 {
		// no version found, just gatherer name
		return parts[0], "", nil
	}
	if len(parts) != 2 {
		return "", "", fmt.Errorf(
			"could not extract the gatherer version from %s, version should follow <gathererName>@<version> syntax",
			gathererName,
		)
	}
	return parts[0], parts[1], nil
}
