package gatherers

import (
	"context"
	"log/slog"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

const (
	ProductsGathererName = "products"
	productsDefaultPath  = "/etc/products.d/"
)

// nolint:gochecknoglobals
var (
	ProductsFolderMissingError = entities.FactGatheringError{
		Type:    "products-folder-missing-error",
		Message: "products folder does not exist",
	}

	ProductsFolderReadingError = entities.FactGatheringError{
		Type:    "products-folder-reading-error",
		Message: "error reading the products folder",
	}

	ProductsFileReadingError = entities.FactGatheringError{
		Type:    "products-file-reading-error",
		Message: "error reading the products file",
	}

	productsXMLelementsToList = map[string]bool{
		"distrotarget":      true,
		"repository":        true,
		"language":          true,
		"url":               true,
		"productdependency": true,
	}
)

type ProductsGatherer struct {
	fs           afero.Fs
	productsPath string
}

func NewProductsGatherer(fs afero.Fs, productsPath string) *ProductsGatherer {
	return &ProductsGatherer{
		fs:           fs,
		productsPath: productsPath,
	}
}

func NewDefaultProductsGatherer() *ProductsGatherer {
	return &ProductsGatherer{fs: afero.NewOsFs(), productsPath: productsDefaultPath}
}

func (g *ProductsGatherer) Gather(_ context.Context, factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	slog.Info("Starting facts gathering process", "gatherer", ProductsGathererName)

	if exists, _ := afero.DirExists(g.fs, g.productsPath); !exists {
		gatheringError := ProductsFolderMissingError.Wrap(g.productsPath)
		slog.Error(gatheringError.Error())
		return nil, gatheringError
	}

	productFiles, err := afero.ReadDir(g.fs, g.productsPath)
	if err != nil {
		gatheringError := ProductsFolderReadingError.Wrap(g.productsPath).Wrap(err.Error())
		slog.Error(gatheringError.Error())
		return nil, gatheringError
	}

	productsFactValueMap := make(map[string]entities.FactValue)
	for _, productFile := range productFiles {
		productFileName := productFile.Name()
		product, err := parseProductFile(g.fs, path.Join(g.productsPath, productFileName))
		if err != nil {
			gatheringError := ProductsFileReadingError.Wrap(productFileName).Wrap(err.Error())
			slog.Error(gatheringError.Error())
			return nil, gatheringError
		}

		productsFactValueMap[productFileName] = product
	}

	for _, requestedFact := range factsRequests {
		facts = append(facts, entities.NewFactGatheredWithRequest(
			requestedFact, &entities.FactValueMap{Value: productsFactValueMap}))
	}

	slog.Info("Requested facts gathered", "gatherer", ProductsGathererName)
	return facts, nil
}

func parseProductFile(fs afero.Fs, productFilePath string) (entities.FactValue, error) {
	productFile, err := afero.ReadFile(fs, productFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "could not open product file")
	}

	factValueMap, err := parseXMLToFactValueMap(productFile, productsXMLelementsToList)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse product file")
	}

	return factValueMap, nil
}
