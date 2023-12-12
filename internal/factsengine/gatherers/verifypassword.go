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
)

// nolint:gochecknoglobals
var (
	checkableUsernames   = []string{"hacluster"}
	unsafePasswords      = []string{"linux"}
	passwordNotSetValues = "!*:;\\" // Get more info with "man 3 crypt"
)

// nolint:gochecknoglobals
var (
	VerifyPasswordInvalidUsername = entities.FactGatheringError{
		Type:    "verify-password-invalid-username",
		Message: "requested user is not whitelisted for password check",
	}

	VerifyPasswordShadowError = entities.FactGatheringError{
		Type:    "verify-password-shadow-error",
		Message: "error getting shadow output",
	}

	VerifyPasswordPasswordBlocked = entities.FactGatheringError{
		Type:    "verify-password-password-blocked",
		Message: "password authentication blocked for user",
	}

	VerifyPasswordPasswordNotSet = entities.FactGatheringError{
		Type:    "verify-password-password-not-set",
		Message: "password not set for user",
	}

	VerifyPasswordCryptError = entities.FactGatheringError{
		Type:    "verify-password-crypt-error",
		Message: "error while verifying the password for user",
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
		if !utils.Contains(checkableUsernames, factReq.Argument) {
			gatheringError := VerifyPasswordInvalidUsername.Wrap(factReq.Argument)
			log.Error(gatheringError)
			facts = append(facts, entities.NewFactGatheredWithError(factReq, gatheringError))
			continue
		}
		username := factReq.Argument

		hash, err := g.getHash(username)

		switch {
		case err != nil:
			{
				gatheringError := VerifyPasswordShadowError.Wrap(err.Error())
				log.Error(gatheringError)
				facts = append(facts, entities.NewFactGatheredWithError(factReq, gatheringError))
				continue
			}

		case len(hash) == 0:
			{
				gatheringError := VerifyPasswordPasswordNotSet.Wrap(username)
				log.Error(gatheringError)
				facts = append(facts, entities.NewFactGatheredWithError(factReq, gatheringError))
				continue
			}

		case strings.ContainsAny(hash, passwordNotSetValues):
			{
				gatheringError := VerifyPasswordPasswordBlocked.Wrap(username)
				log.Error(gatheringError)
				facts = append(facts, entities.NewFactGatheredWithError(factReq, gatheringError))
				continue
			}
		}

		log.Debugf("Obtained hash using user %s: %s", username, hash)

		crypter := sha512crypt.New()
		isPasswordWeak := false
		var gatheringError *entities.FactGatheringError
		for _, password := range unsafePasswords {
			passwordBytes := []byte(password)

			matchErr := crypter.Verify(hash, passwordBytes)
			if matchErr == nil {
				isPasswordWeak = true
				break
			}
			if !errors.Is(matchErr, crypt.ErrKeyMismatch) {
				gatheringError = VerifyPasswordCryptError.Wrap(username).Wrap(matchErr.Error())
				break
			}
		}

		if gatheringError != nil {
			fact := entities.NewFactGatheredWithError(factReq, gatheringError)
			facts = append(facts, fact)
			continue
		}

		fact := entities.NewFactGatheredWithRequest(factReq,
			&entities.FactValueBool{Value: isPasswordWeak})
		facts = append(facts, fact)
	}

	log.Infof("Requested password verifying facts gathered")
	return facts, nil
}

func (g *VerifyPasswordGatherer) getHash(user string) (string, error) {
	shadow, err := g.executor.Exec("getent", "shadow", user)
	if err != nil {
		return "", errors.Wrap(err, "Error getting hash")
	}

	fields := strings.Split(string(shadow), ":")
	if len(fields) != 9 {
		return "", fmt.Errorf("shadow output does not have 9 fields: %s", string(shadow))
	}

	return fields[1], nil
}
