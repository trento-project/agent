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
	caching "github.com/trento-project/agent/pkg/cache"
	"github.com/trento-project/agent/test/helpers"
)

type AWSMetadataTestSuite struct {
	suite.Suite
	mockHTTPClient *mocks.HTTPClient
	cache          *caching.Cache
}

type metadataFixtureSpec struct {
	fixture     string
	requestPath string
}

const mockedIMDSToken string = "awsMetadataToken"

func matchesMetadataPath(req *http.Request, path string) bool {
	return req.Method == http.MethodGet && req.URL.Path == path
}

func matchesMetadataToken(req *http.Request, token string) bool {
	return req.Header.Get("X-aws-ec2-metadata-token") == token
}

func fetchTokenRequest() any {
	return mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == http.MethodPut &&
			req.URL.Path == "/latest/api/token" &&
			req.Header.Get("X-aws-ec2-metadata-token-ttl-seconds") == "21600"
	})
}

func matchesRequestPathAndMetadata(matchingPath string, matchingToken string) any {
	return mock.MatchedBy(func(req *http.Request) bool {
		return matchesMetadataPath(req, matchingPath) && matchesMetadataToken(req, matchingToken)
	})
}

func rootMetadataRequest(matchingToken string) any {
	return matchesRequestPathAndMetadata("/latest/meta-data/", matchingToken)
}

func requestWithToken(matchingToken string) any {
	return mock.MatchedBy(func(req *http.Request) bool {
		return matchesMetadataToken(req, matchingToken)
	})
}

func rootMetadataRequestWithInitialToken() any {
	return rootMetadataRequest(mockedIMDSToken)
}

func mockSuccessfulTokenResponse(mockHTTPClient *mocks.HTTPClient, metadataToken string) *mock.Call {
	token := []byte(metadataToken)

	body := io.NopCloser(bytes.NewReader(token))
	successfulTokenResponse := &http.Response{
		StatusCode: 200,
		Body:       body,
	}

	return mockHTTPClient.On("Do", fetchTokenRequest()).Return(
		successfulTokenResponse, nil,
	)
}

func mockInitialTokenResponse(mockHTTPClient *mocks.HTTPClient) *mock.Call {
	return mockSuccessfulTokenResponse(mockHTTPClient, mockedIMDSToken)
}

func mockSuccessfulMetadataDiscovery(mockHTTPClient *mocks.HTTPClient, matcher func(metadataFixtureSpec) any) {
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

	for _, fixtureSpec := range fixtures {
		aFile, _ := os.Open(path.Join(fixturesFolder, fixtureSpec.fixture))
		bodyText, _ := io.ReadAll(aFile)
		body := io.NopCloser(bytes.NewReader(bodyText))

		response := &http.Response{
			StatusCode: 200,
			Body:       body,
		}

		mockHTTPClient.On("Do", matcher(fixtureSpec)).Return(
			response, nil,
		).Once()
	}
}

func TestAWSMetadataTestSuite(t *testing.T) {
	suite.Run(t, new(AWSMetadataTestSuite))
}

func (suite *AWSMetadataTestSuite) SetupTest() {
	suite.mockHTTPClient = new(mocks.HTTPClient)
	suite.cache = caching.NewCache()
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

		extractedMetadata, err := cloud.NewAWSMetadata(ctx, suite.mockHTTPClient, suite.cache)

		suite.ErrorContains(err, fmt.Sprintf("failed to fetch metadata token: %s", response.Status))
		suite.Nil(extractedMetadata)
	}

}

func (suite *AWSMetadataTestSuite) TestUnableToFetchMetadataBecauseDisabledIMDS() {
	response := &http.Response{
		StatusCode: 403,
		Status:     "403 Forbidden",
	}

	for _, emptyResponse := range withEmptyBody(response) {
		mockInitialTokenResponse(suite.mockHTTPClient).Once()

		ctx := context.TODO()

		suite.mockHTTPClient.On("Do", rootMetadataRequestWithInitialToken()).Return(
			emptyResponse, nil,
		)

		extractedMetadata, err := cloud.NewAWSMetadata(ctx, suite.mockHTTPClient, suite.cache)

		suite.ErrorContains(err, "failed to fetch metadata. IMDS might be disabled: 403 Forbidden")
		suite.Nil(extractedMetadata)
	}
}

func (suite *AWSMetadataTestSuite) TestRefreshIMDSToken() {
	unauthorizedResponse := &http.Response{
		StatusCode: 401,
	}
	for _, response := range withEmptyBody(unauthorizedResponse) {
		ctx := context.TODO()

		suite.mockHTTPClient.On("Do", rootMetadataRequestWithInitialToken()).Return(
			response, nil,
		)

		refreshedToken := "refreshedToken"

		mockSuccessfulTokenResponse(suite.mockHTTPClient, refreshedToken).Once()

		mockSuccessfulMetadataDiscovery(suite.mockHTTPClient, func(_ metadataFixtureSpec) any {
			return requestWithToken(refreshedToken)
		})

		metadataFetchContext := cloud.NewMetadataFetchContext(suite.cache).WithToken(mockedIMDSToken)

		_, err := cloud.NewAWSMetadataWithFetchContext(ctx, suite.mockHTTPClient, metadataFetchContext)

		suite.NoError(err)
	}
}

func (suite *AWSMetadataTestSuite) TestRefreshIMDSTokenExceedsAttempts() {
	unauthorizedResponse := &http.Response{
		StatusCode: 401,
	}

	for _, response := range withEmptyBody(unauthorizedResponse) {
		ctx := context.TODO()

		suite.mockHTTPClient.On("Do", rootMetadataRequestWithInitialToken()).Return(
			response, nil,
		).Once()

		forbiddenResponse := &http.Response{
			StatusCode: 403,
		}

		suite.mockHTTPClient.On("Do", fetchTokenRequest()).Return(
			forbiddenResponse, nil,
		).Times(3)

		metadataFetchContext := cloud.NewMetadataFetchContext(suite.cache).WithToken(mockedIMDSToken)

		_, err := cloud.NewAWSMetadataWithFetchContext(ctx, suite.mockHTTPClient, metadataFetchContext)

		suite.ErrorContains(err, "failed to fetch metadata: reached maximum token refresh attempts")
	}
}

func (suite *AWSMetadataTestSuite) TestNewAWSMetadata() {
	mockInitialTokenResponse(suite.mockHTTPClient).Once()

	ctx := context.TODO()

	mockSuccessfulMetadataDiscovery(suite.mockHTTPClient, func(fixtureSpec metadataFixtureSpec) any {
		return matchesRequestPathAndMetadata(fixtureSpec.requestPath, mockedIMDSToken)
	})

	m, err := cloud.NewAWSMetadata(ctx, suite.mockHTTPClient, suite.cache)

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
