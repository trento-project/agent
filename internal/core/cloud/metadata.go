package cloud

import (
	"context"

	"github.com/trento-project/agent/internal/core/provider"
	"github.com/trento-project/agent/pkg/utils"
)

type Instance struct {
	Provider string
	Metadata interface{}
}

func NewCloudInstance(
	ctx context.Context,
	commandExecutor utils.CommandExecutor,
	client HTTPClient,
) (*Instance, error) {
	var err error
	var cloudMetadata interface{}

	providerIdentifier := provider.NewIdentifier(commandExecutor)

	identifiedProvider, err := providerIdentifier.IdentifyProvider()
	if err != nil {
		return nil, err
	}

	cInst := &Instance{
		Metadata: nil,
		Provider: identifiedProvider,
	}

	switch identifiedProvider {
	case provider.Azure:
		{
			cloudMetadata, err = NewAzureMetadata(ctx, client)
			if err != nil {
				return nil, err
			}
		}
	case provider.AWS:
		{
			awsMetadata, err := NewAWSMetadata(ctx, client)
			if err != nil {
				return nil, err
			}
			cloudMetadata = NewAWSMetadataDto(awsMetadata)
		}
	case provider.GCP:
		{
			gcpMetadata, err := NewGCPMetadata(ctx, client)
			if err != nil {
				return nil, err
			}
			cloudMetadata = NewGCPMetadataDto(gcpMetadata)
		}
	}

	cInst.Metadata = cloudMetadata

	return cInst, nil

}
