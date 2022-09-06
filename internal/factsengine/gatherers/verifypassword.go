package gatherers

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	crypt "github.com/tredoe/osutil/user/crypt"
	sha512crypt "github.com/tredoe/osutil/user/crypt/sha512_crypt"
	"github.com/trento-project/agent/internal/factsengine/entities"
)

const (
	VerifyPasswordGathererName = "verify_password"
)

var (
	VerifyPasswordInvalidArgument = entities.FactGatheringError{ // nolint
		Type:    "verify-password-invalid-argument",
		Message: "the provided argument should follow the \"username:password\" format",
	}

	VerifyPasswordSaltError = entities.FactGatheringError{ // nolint
		Type:    "verify-password-salt-error",
		Message: "error getting password salt",
	}
)

type VerifyPasswordGatherer struct {
	executor CommandExecutor
}

func NewDefaultPasswordGatherer() *VerifyPasswordGatherer {
	return NewVerifyPasswordGatherer(Executor{})
}

func NewVerifyPasswordGatherer(executor CommandExecutor) *VerifyPasswordGatherer {
	return &VerifyPasswordGatherer{
		executor,
	}
}

/*
This gatherer expects the next format for the argument: "username:password"
Where:
- username: The user which the password is verified
- password: The password to verify
*/

func (g *VerifyPasswordGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.FactsGatheredItem, error) {
	facts := []entities.FactsGatheredItem{}
	log.Infof("Starting password verifying facts gathering process")

	for _, factReq := range factsRequests {
		arguments := strings.Split(factReq.Argument, ":")
		if len(arguments) != 2 {
			gatheringError := VerifyPasswordInvalidArgument
			log.Errorf(gatheringError.Error())
			fact := entities.NewFactGatheredWithError(factReq, &gatheringError)
			facts = append(facts, fact)
			continue
		}

		username := arguments[0]
		password := []byte(arguments[1])

		salt, hash, err := g.getSalt(username)
		if err != nil {
			gatheringError := VerifyPasswordSaltError
			log.Errorf(gatheringError.Error())
			fact := entities.NewFactGatheredWithError(factReq, &gatheringError)
			facts = append(facts, fact)
			continue
		}
		log.Debugf("Obtained salt using user %s and password %s: %s", username, password, salt)

		crypter := sha512crypt.New()
		match := crypter.Verify(hash, password)

		fact := entities.NewFactGatheredWithRequest(factReq, !errors.Is(match, crypt.ErrKeyMismatch))
		facts = append(facts, fact)
	}

	log.Infof("Requested password verifying facts gathered")
	return facts, nil
}

func (g *VerifyPasswordGatherer) getSalt(user string) ([]byte, string, error) {
	shadow, err := g.executor.Exec("getent", "shadow", user)
	if err != nil {
		return nil, "", errors.Wrap(err, "Error getting salt")
	}
	salt := strings.Split(string(shadow), "$")[2]
	hash := strings.Split(string(shadow), ":")[1]

	return []byte(fmt.Sprintf("$6$%s", salt)), hash, nil
}
