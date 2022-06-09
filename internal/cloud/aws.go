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
	awsMetadataUrl      = "http://169.254.169.254/latest/"
	awsMetadataResource = "meta-data"
)

type AwsMetadata struct {
	AmiId               string              `json:"ami-id,omitempty"`
	BlockDeviceMapping  map[string]string   `json:"block-device-mapping,omitempty"`
	IdentityCredentials IdentityCredentials `json:"identity-credentials,omitempty"`
	InstanceId          string              `json:"instance-id,omitempty"`
	InstanceType        string              `json:"instance-type,omitemtpy"`
	Network             AwsNetwork          `json:"network,omitempty"`
	Placement           Placement           `json:"placement,omitempty"`
}

type IdentityCredentials struct {
	EC2 struct {
		Info struct {
			AccountId string `json:"AccountId,omitempty"`
		} `json:"info,omitempty"`
	} `json:"ec2,omitempty"`
}

type AwsNetwork struct {
	Interfaces struct {
		Macs map[string]struct {
			VpcId string `json:"vpc-id,omitempty"`
		} `json:"macs,omitempty"`
	} `json:"interfaces,omitempty"`
}

type Placement struct {
	AvailabilityZone string `json:"availability-zone,omitempty"`
	Region           string `json:"region,omitempty"`
}

func NewAwsMetadata() (*AwsMetadata, error) {
	var err error
	awsMetadata := &AwsMetadata{}

	firstElementsList := []string{fmt.Sprintf("%s/", awsMetadataResource)}
	metadata, err := buildAwsMetadata(awsMetadataUrl, firstElementsList)
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
		new_url := url + element

		response, err := requestMetadata(new_url)
		if err != nil {
			return metadata, err
		}

		if element[len(element)-1:] == "/" {
			current_element := strings.Trim(element, "/")
			new_elements := strings.Split(fmt.Sprintf("%v", response), "\n")
			metadata[current_element], err = buildAwsMetadata(new_url, new_elements)
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
	} else {
		return string(body), nil
	}
}
