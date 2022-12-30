package gatherers

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	crypt "github.com/tredoe/osutil/user/crypt"
	sha512crypt "github.com/tredoe/osutil/user/crypt/sha512_crypt"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	VerifyPasswordGathererName = "verify_password"
	checkableUsernames         = "hacluster"
)

// nolint:gochecknoglobals
var (
	VerifyPasswordInvalidUsername = entities.FactGatheringError{
		Type:    "verify-password-invalid-username",
		Message: "unknown username or not allowed to check",
	}
)

type VerifyPasswordGatherer struct {
	executor utils.CommandExecutor
}

func NewDefaultPasswordGatherer() *VerifyPasswordGatherer {
	return NewVerifyPasswordGatherer(utils.Executor{})
}

func NewVerifyPasswordGatherer(executor utils.CommandExecutor) *VerifyPasswordGatherer {
	return &VerifyPasswordGatherer{
		executor,
	}
}

/*
This gatherer expects the only the user which password is verified
*/
func (g *VerifyPasswordGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting password verifying facts gathering process")

	for _, factReq := range factsRequests {
		if !strings.Contains(checkableUsernames, factReq.Argument) {
			gatheringError := VerifyPasswordInvalidUsername.Wrap(factReq.Argument)
			fact := entities.NewFactGatheredWithError(factReq, gatheringError)
			facts = append(facts, fact)
			continue
		}
		username := factReq.Argument
		password := []byte("linux")

		salt, hash, err := g.getSalt(username)
		if err != nil {
			log.Error(err)
		}
		log.Debugf("Obtained salt using user %s and password %s: %s", username, password, salt)

		crypter := sha512crypt.New()
		match := crypter.Verify(hash, password)

		fact := entities.NewFactGatheredWithRequest(factReq,
			&entities.FactValueBool{Value: !errors.Is(match, crypt.ErrKeyMismatch)})
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
