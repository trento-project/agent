/*
Based on
- https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instancedata-data-retrieval.html
- https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instancedata-data-categories.html
*/

package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	caching "github.com/trento-project/agent/pkg/cache"
)

const (
	awsMetadataURL                = "http://169.254.169.254/latest/"
	awsMetadataResource           = "meta-data"
	metadataTokenTTLHeader string = "X-aws-ec2-metadata-token-ttl-seconds" //nolint
	metadataTokenHeader    string = "X-aws-ec2-metadata-token"             //nolint
	maxRefreshAttempts     byte   = 3

	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/configuring-instance-metadata-service.html
	// (TTL) for the token, in seconds, up to a maximum of six hours (21,600 seconds).
	awsEC2MetadataTokenTTL     int    = 60 * 60 * 6 // 6 hours
	metadataFetchTokenCacheKey string = "awsMetadataToken"
)

type MetadataFetchContext struct {
	Token          string
	RefreshAttempt byte
	URL            string
	cache          *caching.Cache
}

func (context MetadataFetchContext) HasToken() bool {
	return context.Token != ""
}

func (context MetadataFetchContext) WithToken(token string) MetadataFetchContext {
	entry := context.cache.Update(metadataFetchTokenCacheKey, caching.DefaultUpdateFn(token))
	context.Token = entry.Content.(string)
	return context
}

func (context MetadataFetchContext) IncreaseTokenRefreshAttempt() MetadataFetchContext {
	context.cache.Delete(metadataFetchTokenCacheKey)
	context.Token = ""
	context.RefreshAttempt++
	return context
}

func (context MetadataFetchContext) WithNextURL(part string) MetadataFetchContext {
	context.URL += part
	return context
}

func NewMetadataFetchContext(cache *caching.Cache) MetadataFetchContext {
	return MetadataFetchContext{
		URL:   awsMetadataURL,
		Token: cache.GetOrDefault(metadataFetchTokenCacheKey, "").Content.(string),
		cache: cache,
	}
}

type AWSMetadata struct {
	AmiID               string              `json:"ami-id"`
	BlockDeviceMapping  map[string]string   `json:"block-device-mapping"`
	IdentityCredentials IdentityCredentials `json:"identity-credentials"`
	InstanceID          string              `json:"instance-id"`
	InstanceType        string              `json:"instance-type,omitempty"`
	Network             AWSNetwork          `json:"network"`
	Placement           Placement           `json:"placement"`
}

type IdentityCredentials struct {
	EC2 struct {
		Info struct {
			AccountID string `json:"AccountId"`
		} `json:"info"`
	} `json:"ec2"`
}

type AWSNetwork struct {
	Interfaces struct {
		Macs map[string]MacEntry `json:"macs"`
	} `json:"interfaces"`
}

type MacEntry struct {
	VpcID string `json:"vpc-id"`
}

type Placement struct {
	AvailabilityZone string `json:"availability-zone"`
	Region           string `json:"region"`
}

func NewAWSMetadata(ctx context.Context, client HTTPClient, cache *caching.Cache) (*AWSMetadata, error) {
	return NewAWSMetadataWithFetchContext(ctx, client, NewMetadataFetchContext(cache))
}

func NewAWSMetadataWithFetchContext(
	ctx context.Context,
	client HTTPClient,
	metadataFetchContext MetadataFetchContext,
) (*AWSMetadata, error) {
	var err error
	awsMetadata := &AWSMetadata{
		AmiID:              "",
		BlockDeviceMapping: map[string]string{},
		IdentityCredentials: IdentityCredentials{
			EC2: struct {
				Info struct {
					AccountID string "json:\"AccountId\""
				} "json:\"info\""
			}{
				Info: struct {
					AccountID string "json:\"AccountId\""
				}{
					AccountID: "",
				},
			},
		},
		InstanceID:   "",
		InstanceType: "",
		Network: AWSNetwork{
			Interfaces: struct {
				Macs map[string]MacEntry "json:\"macs\""
			}{
				Macs: map[string]MacEntry{},
			},
		},
		Placement: Placement{
			AvailabilityZone: "",
			Region:           "",
		},
	}

	firstElementsList := []string{fmt.Sprintf("%s/", awsMetadataResource)}
	metadata, err := buildAWSMetadata(ctx, client, firstElementsList, metadataFetchContext)
	if err != nil {
		return nil, err
	}

	jsonMetadata, err := json.Marshal(metadata[awsMetadataResource])
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonMetadata, awsMetadata)
	if err != nil {
		return nil, err
	}

	return awsMetadata, err
}

func buildAWSMetadata(
	ctx context.Context,
	client HTTPClient,
	elements []string,
	metadataFetchContext MetadataFetchContext,
) (map[string]any, error) {
	metadata := make(map[string]any)

	metadataFetchContext, err := requestMetadataToken(ctx, client, metadataFetchContext)
	if err != nil {
		return nil, err
	}

	for _, element := range elements {
		response, metadataFetchContext, err := requestMetadata(ctx, client, metadataFetchContext.WithNextURL(element))
		if err != nil {
			return metadata, err
		}

		if element[len(element)-1:] == "/" {
			currentElement := strings.Trim(element, "/")
			newElements := strings.Split(fmt.Sprintf("%v", response), "\n")

			metadata[currentElement], err = buildAWSMetadata(ctx, client, newElements, metadataFetchContext)
			if err != nil {
				return nil, err
			}
		} else {
			metadata[element] = response
		}
	}

	return metadata, nil
}

func requestMetadata(
	ctx context.Context,
	client HTTPClient,
	metadataFetchContext MetadataFetchContext,
) (any, MetadataFetchContext, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, metadataFetchContext.URL, nil)
	req.Header.Add(metadataTokenHeader, metadataFetchContext.Token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, metadataFetchContext, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		metadataFetchContext, err := refreshToken(ctx, client, metadataFetchContext)
		if err != nil {
			return nil, metadataFetchContext, err
		}

		return requestMetadata(ctx, client, metadataFetchContext)
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, metadataFetchContext, errors.Errorf("failed to fetch metadata. IMDS might be disabled: %s", resp.Status)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, metadataFetchContext, err
	}

	// The metadata endpoint may return json elements
	if json.Valid(body) {
		var jsonData any
		err := json.Unmarshal(body, &jsonData)
		return jsonData, metadataFetchContext, err
	}
	return string(body), metadataFetchContext, nil
}

func refreshToken(
	ctx context.Context,
	client HTTPClient,
	metadataFetchContext MetadataFetchContext,
) (MetadataFetchContext, error) {
	metadataFetchContext = metadataFetchContext.IncreaseTokenRefreshAttempt()
	log.Debugf("Refreshing metadata token... attempt %d", metadataFetchContext.RefreshAttempt)

	if metadataFetchContext.RefreshAttempt > maxRefreshAttempts {
		return metadataFetchContext, errors.Errorf("failed to fetch metadata: reached maximum token refresh attempts")
	}

	metadataFetchContext, err := requestMetadataToken(
		ctx,
		client,
		metadataFetchContext,
	)

	if err != nil {
		log.Debugf("failed to refresh metadata token: %s", err)
		return refreshToken(ctx, client, metadataFetchContext)
	}

	log.Debug("Metadata token refreshed successfully")
	return metadataFetchContext, nil
}

func requestMetadataToken(
	ctx context.Context,
	client HTTPClient,
	metadataFetchContext MetadataFetchContext,
) (MetadataFetchContext, error) {
	log.Debug("Fetching IMDS token...")

	if metadataFetchContext.HasToken() {
		log.Debug("Using previously fetched IMDS token")
		return metadataFetchContext, nil
	}

	url := fmt.Sprintf("%sapi/token", awsMetadataURL)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, url, nil)

	req.Header.Add(metadataTokenTTLHeader, strconv.Itoa(awsEC2MetadataTokenTTL))

	resp, err := client.Do(req)
	if err != nil {
		log.Debugf("An error occurred while fetching IMDS token: %s", err)
		return metadataFetchContext, err
	}

	if resp.StatusCode != http.StatusOK {
		return metadataFetchContext, errors.Errorf("failed to fetch metadata token: %s", resp.Status)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return metadataFetchContext, err
	}

	log.Debug("Metadata token fetched successfully")

	return metadataFetchContext.WithToken(string(body)), nil
}
