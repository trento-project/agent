package gatherers

import "github.com/pkg/errors"

type Registry struct {
	gatherers map[string]FactGatherer
}

func (m *Registry) GetGatherer(name string) (FactGatherer, error) {
	if g, found := m.gatherers[name]; found {
		return g, nil
	}
	return nil, errors.Errorf("gatherer %s not found", name)
}

func (m *Registry) AvailableGatherers() []string {
	gatherersList := []string{}

	for gatherer := range m.gatherers {
		gatherersList = append(gatherersList, gatherer)
	}

	return gatherersList
}

// This is not safe, please not use concurrently.
func (m *Registry) AddGatherers(gatherers map[string]FactGatherer) {
	maps := []map[string]FactGatherer{m.gatherers, gatherers}
	result := make(map[string]FactGatherer)

	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	m.gatherers = result
}

func NewRegistry(gatherers map[string]FactGatherer) *Registry {
	return &Registry{
		gatherers: gatherers,
	}
}
