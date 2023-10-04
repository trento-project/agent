package gatherers

import "os/user"

type CredentialsFetcher struct{}

func (f *CredentialsFetcher) GetUsernameByID(userID string) (string, error) {
	u, err := user.LookupId(userID)
	if err != nil {
		return "", err
	}
	return u.Username, nil
}

func (f *CredentialsFetcher) GetGroupByID(groupID string) (string, error) {
	u, err := user.LookupGroupId(groupID)
	if err != nil {
		return "", err
	}

	return u.Name, nil
}
