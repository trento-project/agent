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
	unsafePasswords            = "linux"
)

// nolint:gochecknoglobals
var (
	VerifyPasswordInvalidUsername = entities.FactGatheringError{
		Type:    "verify-password-invalid-username",
		Message: "requested user is not whitelisted for password check",
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
This gatherer expects only the username for which the password will be verified
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

		salt, hash, err := g.getSalt(username)
		if err != nil {
			log.Error(err)
		}
		log.Debugf("Obtained salt using user %s: %s", username, salt)

		crypter := sha512crypt.New()
		isPasswordWeak := false
		for _, password := range strings.Split(unsafePasswords, ",") {
			passwordBytes := []byte(password)
			match := crypter.Verify(hash, passwordBytes)
			if !errors.Is(match, crypt.ErrKeyMismatch) {
				isPasswordWeak = true
				break
			}
		}

		fact := entities.NewFactGatheredWithRequest(factReq,
			&entities.FactValueBool{Value: isPasswordWeak})
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
