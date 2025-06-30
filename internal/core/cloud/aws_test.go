package cloud_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/cloud"
	"github.com/trento-project/agent/internal/core/cloud/mocks"
	"github.com/trento-project/agent/test/helpers"
)

type AWSMetadataTestSuite struct {
	suite.Suite
	mockHTTPClient *mocks.MockHTTPClient
}

type metadataFixtureSpec struct {
	fixture     string
	requestPath string
}

const mockedIMDSToken string = "awsMetadataToken"

func fetchTokenRequest() any {
	return mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == http.MethodPut &&
			req.URL.Path == "/latest/api/token" &&
			req.Header.Get("X-aws-ec2-metadata-token-ttl-seconds") == "120"
	})
}

func matchesRequestPathAndToken(matchingPath string, matchingToken string) any {
	return mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == http.MethodGet &&
			req.URL.Path == matchingPath &&
			req.Header.Get("X-aws-ec2-metadata-token") == matchingToken
	})
}

func mockSuccessfulTokenResponse(mockHTTPClient *mocks.MockHTTPClient) *mock.Call {
	token := []byte(mockedIMDSToken)

	body := io.NopCloser(bytes.NewReader(token))
	successfulTokenResponse := &http.Response{
		StatusCode: 200,
		Body:       body,
	}

	return mockHTTPClient.On("Do", fetchTokenRequest()).Return(
		successfulTokenResponse, nil,
	)
}

func mockSuccessfulMetadataDiscoveryUntil(mockHTTPClient *mocks.MockHTTPClient, untilFixture string, failureResponse ...*http.Response) {
	fixtures := []metadataFixtureSpec{
		{
			fixture:     "meta-data",
			requestPath: "/latest/meta-data/",
		},
		{
			fixture:     "ami-id",
			requestPath: "/latest/meta-data/ami-id",
		},
		{
			fixture:     "block-device-mapping",
			requestPath: "/latest/meta-data/block-device-mapping/",
		},
		{
			fixture:     "ebs1",
			requestPath: "/latest/meta-data/block-device-mapping/ebs1",
		},
		{
			fixture:     "ebs2",
			requestPath: "/latest/meta-data/block-device-mapping/ebs2",
		},
		{
			fixture:     "root",
			requestPath: "/latest/meta-data/block-device-mapping/root",
		},
		{
			fixture:     "identity-credentials",
			requestPath: "/latest/meta-data/identity-credentials/",
		},
		{
			fixture:     "ec2",
			requestPath: "/latest/meta-data/identity-credentials/ec2/",
		},
		{
			fixture:     "info",
			requestPath: "/latest/meta-data/identity-credentials/ec2/info",
		},
		{
			fixture:     "instance-id",
			requestPath: "/latest/meta-data/instance-id",
		},
		{
			fixture:     "instance-type",
			requestPath: "/latest/meta-data/instance-type",
		},
		{
			fixture:     "network",
			requestPath: "/latest/meta-data/network/",
		},
		{
			fixture:     "interfaces",
			requestPath: "/latest/meta-data/network/interfaces/",
		},
		{
			fixture:     "macs",
			requestPath: "/latest/meta-data/network/interfaces/macs/",
		},
		{
			fixture:     "some-mac",
			requestPath: "/latest/meta-data/network/interfaces/macs/some-mac/",
		},
		{
			fixture:     "vpc-id",
			requestPath: "/latest/meta-data/network/interfaces/macs/some-mac/vpc-id",
		},
		{
			fixture:     "placement",
			requestPath: "/latest/meta-data/placement/",
		},
		{
			fixture:     "availability-zone",
			requestPath: "/latest/meta-data/placement/availability-zone",
		},
		{
			fixture:     "region",
			requestPath: "/latest/meta-data/placement/region",
		},
	}

	fixturesFolder := helpers.GetFixturePath("discovery/aws")

	matcher := func(fixtureSpec metadataFixtureSpec) any {
		return matchesRequestPathAndToken(fixtureSpec.requestPath, mockedIMDSToken)
	}

	for _, fixtureSpec := range fixtures {
		aFile, _ := os.Open(path.Join(fixturesFolder, fixtureSpec.fixture))
		bodyText, _ := io.ReadAll(aFile)
		body := io.NopCloser(bytes.NewReader(bodyText))

		response := &http.Response{
			StatusCode: 200,
			Body:       body,
		}

		if fixtureSpec.fixture == untilFixture {
			if len(failureResponse) > 0 {
				response = failureResponse[0]
			} else {
				response = &http.Response{
					StatusCode: 403,
					Status:     "403 Forbidden",
					Body:       io.NopCloser(bytes.NewReader([]byte(""))),
				}
			}
		}

		mockHTTPClient.On("Do", matcher(fixtureSpec)).Return(
			response, nil,
		).Once()
	}
}

func mockSuccessfulMetadataDiscovery(mockHTTPClient *mocks.MockHTTPClient) {
	mockSuccessfulMetadataDiscoveryUntil(mockHTTPClient, "the-end")
}

func withEmptyBody(responses ...*http.Response) []*http.Response {
	emptyBodies := [][]byte{
		[]byte(""),
		[]byte(`
		`),
		[]byte(`
	
		`),
	}

	resultingResponses := []*http.Response{}
	for _, errorResponse := range responses {
		for _, emptyBody := range emptyBodies {
			errorResponseWithEmptyBody := errorResponse
			errorResponseWithEmptyBody.Body = io.NopCloser(bytes.NewReader(emptyBody))
			resultingResponses = append(resultingResponses, errorResponseWithEmptyBody)
		}
	}
	return resultingResponses
}

func TestAWSMetadataTestSuite(t *testing.T) {
	suite.Run(t, new(AWSMetadataTestSuite))
}

func (suite *AWSMetadataTestSuite) SetupTest() {
	suite.mockHTTPClient = new(mocks.MockHTTPClient)
}

func (suite *AWSMetadataTestSuite) TestUnableToFetchMetadataToken() {
	responses := []*http.Response{
		{
			StatusCode: 400,
			Status:     "400 Bad Request",
		},
		{
			StatusCode: 401,
			Status:     "401 Unauthorized",
		},
		{
			StatusCode: 403,
			Status:     "403 Forbidden",
		},
		{
			StatusCode: 404,
			Status:     "404 Not Found",
		},
		{
			StatusCode: 500,
			Status:     "500 Internal Server Error",
		},
	}

	for _, response := range withEmptyBody(responses...) {
		ctx := context.TODO()

		suite.mockHTTPClient.On("Do", fetchTokenRequest()).Return(
			response, nil,
		).Once()

		extractedMetadata, err := cloud.NewAWSMetadata(ctx, suite.mockHTTPClient)

		suite.ErrorContains(err, fmt.Sprintf("failed to fetch metadata token: %s", response.Status))
		suite.Nil(extractedMetadata)
	}

}

func (suite *AWSMetadataTestSuite) TestUnableToFetchRootMetadata() {
	responses := []*http.Response{
		{
			StatusCode: 401,
			Status:     "401 Unauthorized",
		},
		{
			StatusCode: 403,
			Status:     "403 Forbidden",
		},
	}

	for _, response := range withEmptyBody(responses...) {
		mockSuccessfulTokenResponse(suite.mockHTTPClient).Once()
		mockSuccessfulMetadataDiscoveryUntil(suite.mockHTTPClient, "meta-data", response)

		ctx := context.TODO()

		extractedMetadata, err := cloud.NewAWSMetadata(ctx, suite.mockHTTPClient)

		suite.ErrorContains(err, fmt.Sprintf("failed to fetch AWS metadata: %s", response.Status))

		suite.Nil(extractedMetadata)
	}
}

func (suite *AWSMetadataTestSuite) TestUnableToFetchMetadata() {
	responses := []*http.Response{
		{
			StatusCode: 401,
			Status:     "401 Unauthorized",
		},
		{
			StatusCode: 403,
			Status:     "403 Forbidden",
		},
	}

	for _, response := range withEmptyBody(responses...) {
		mockSuccessfulTokenResponse(suite.mockHTTPClient).Once()
		mockSuccessfulMetadataDiscoveryUntil(suite.mockHTTPClient, "info", response)

		ctx := context.TODO()

		extractedMetadata, err := cloud.NewAWSMetadata(ctx, suite.mockHTTPClient)

		suite.ErrorContains(err, fmt.Sprintf("failed to fetch AWS metadata: %s", response.Status))

		suite.Nil(extractedMetadata)
	}
}

func (suite *AWSMetadataTestSuite) TestGracefullyHandlesEmptyResponsesUnableToFetchRootMetadata() {
	response := &http.Response{
		StatusCode: 200,
	}

	for _, responseWithEmptyBody := range withEmptyBody(response) {
		mockSuccessfulTokenResponse(suite.mockHTTPClient).Once()

		suite.mockHTTPClient.On("Do", matchesRequestPathAndToken("/latest/meta-data/", mockedIMDSToken)).Return(
			responseWithEmptyBody, nil,
		).Once()

		ctx := context.TODO()

		_, err := cloud.NewAWSMetadata(ctx, suite.mockHTTPClient)

		suite.NoError(err)
	}
}

func (suite *AWSMetadataTestSuite) TestNewAWSMetadata() {
	mockSuccessfulTokenResponse(suite.mockHTTPClient).Once()
	mockSuccessfulMetadataDiscovery(suite.mockHTTPClient)

	ctx := context.TODO()
	m, err := cloud.NewAWSMetadata(ctx, suite.mockHTTPClient)

	suite.NoError(err)

	suite.Equal("some-ami-id", m.AmiID)
	suite.Equal(map[string]string{
		"root": "/dev/sda",
		"ebs1": "/dev/sdb1",
		"ebs2": "/dev/sdb2",
	}, m.BlockDeviceMapping)
	suite.Equal("some-instance", m.InstanceID)
	suite.Equal("some-instance-type", m.InstanceType)
	suite.Equal("some-account-id", m.IdentityCredentials.EC2.Info.AccountID)
	suite.Equal("some-vpc-id", m.Network.Interfaces.Macs["some-mac"].VpcID)
	suite.Equal("some-availability-zone", m.Placement.AvailabilityZone)
	suite.Equal("some-region", m.Placement.Region)
}
