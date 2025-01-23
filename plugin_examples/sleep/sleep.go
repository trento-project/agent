package main

// go build -o /usr/etc/trento/sleep ./plugin_examples/sleep/sleep.go

import (
	"context"
	"fmt"
	"os/exec"
	"sync"

	"github.com/hashicorp/go-plugin"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/factsengine/plugininterface"
)

type sleepGatherer struct {
}

func (s sleepGatherer) Gather(ctx context.Context, factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}

	log.Infof("Starting sleep plugin facts gathering process")

	wg := sync.WaitGroup{}

	for _, factReq := range factsRequests {
		log.Infof("Sleeping for %s", factReq.Argument)
		fact := entities.NewFactGatheredWithRequest(factReq, &entities.FactValueString{Value: fmt.Sprint(factReq.Argument)})
		facts = append(facts, fact)

		time := fmt.Sprint(factReq.Argument)
		wg.Add(1)
		go func(time string) {
			defer wg.Done()
			cmd := exec.CommandContext(ctx, "sleep", time)
			err := cmd.Run()
			if err != nil {
				log.Errorf("Error running sleep command: %s", err)
			}
		}(time)

	}

	wg.Wait()

	log.Infof("Requested sleep plugin facts gathered")
	return facts, nil
}

func main() {
	d := &sleepGatherer{}

	handshakeConfig := plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "TRENTO_PLUGIN",
		MagicCookieValue: "gatherer",
	}

	var pluginMap = map[string]plugin.Plugin{
		"gatherer": &plugininterface.GathererPlugin{Impl: d},
	}

	plugin.Serve(&plugin.ServeConfig{ // nolint
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
