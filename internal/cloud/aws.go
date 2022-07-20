/*
Based on
- https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instancedata-data-retrieval.html
- https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instancedata-data-categories.html
*/

package cloud

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	awsMetadataURL      = "http://169.254.169.254/latest/"
	awsMetadataResource = "meta-data"
)

type AwsMetadata struct {
	AmiID               string              `json:"ami-id"` // nolint
	BlockDeviceMapping  map[string]string   `json:"block-device-mapping"`
	IdentityCredentials IdentityCredentials `json:"identity-credentials"`
	InstanceID          string              `json:"instance-id"`             //nolint
	InstanceType        string              `json:"instance-type,omitempty"` //nolint
	Network             AwsNetwork          `json:"network"`
	Placement           Placement           `json:"placement"`
}

type IdentityCredentials struct {
	EC2 struct {
		Info struct {
			AccountID string `json:"AccountId"`
		} `json:"info"`
	} `json:"ec2"`
}

type AwsNetwork struct {
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

func NewAwsMetadata() (*AwsMetadata, error) {
	var err error
	awsMetadata := &AwsMetadata{
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
		Network: AwsNetwork{
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
	metadata, err := buildAwsMetadata(awsMetadataURL, firstElementsList)
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

func buildAwsMetadata(url string, elements []string) (map[string]interface{}, error) {
	metadata := make(map[string]interface{})

	for _, element := range elements {
		newURL := url + element

		response, err := requestMetadata(newURL)
		if err != nil {
			return metadata, err
		}

		if element[len(element)-1:] == "/" {
			currentElement := strings.Trim(element, "/")
			newElements := strings.Split(fmt.Sprintf("%v", response), "\n")

			metadata[currentElement], err = buildAwsMetadata(newURL, newElements)
			if err != nil {
				return nil, err
			}
		} else {
			metadata[element] = response
		}
	}

	return metadata, nil
}

func requestMetadata(url string) (interface{}, error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// The metadata endpoint may return json elements
	if json.Valid(body) {
		var jsonData interface{}
		err := json.Unmarshal(body, &jsonData)
		return jsonData, err
	}
	return string(body), nil
}
