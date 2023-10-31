package gatherers_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

type ProductsGathererSuite struct {
	suite.Suite
	//fs             afero.Fs
	productsPath string
}

func TestProductsGathererSuite(t *testing.T) {
	suite.Run(t, new(ProductsGathererSuite))
}

func (s *ProductsGathererSuite) SetupTest() {
	// 	fs := afero.NewMemMapFs()
	// 	err := fs.MkdirAll("/etc/products.d/", 0644)
	// 	s.NoError(err)

	// 	err = afero.WriteFile(fs, "/etc/products.d/baseproduct", []byte(`
	// #!/bin/sh
	// limit.descriptors=1048576
	// systemctl --no-ask-password start SAPS41_40
	// systemctl --no-ask-password start SADS41_41
	// `), 0777)
	// 	s.NoError(err)

	// 	s.fs = fs
	s.productsPath = "/etc/products.d/"
}

func (s *ProductsGathererSuite) TestProductsGathererFolderMissingError() {
	fs := afero.NewMemMapFs()

	fr := []entities.FactRequest{
		{
			Name:     "missing_folder",
			Gatherer: "products@v1",
			CheckID:  "check1",
		},
	}

	gatherer := gatherers.NewProductsGatherer(fs, s.productsPath)

	results, err := gatherer.Gather(fr)
	s.Nil(results)
	s.EqualError(err, "fact gathering error: products-folder-missing-error - "+
		"products folder does not exist: /etc/products.d/")
}

func (s *ProductsGathererSuite) TestProductsGathererReadingError() {
	fs := afero.NewMemMapFs()

	err := afero.WriteFile(fs, "/etc/products.d/baseproduct", []byte(`
<?xml version="1.0" encoding="UTF-8"?>
<product schemeversion="0">
	<vendor>openSUSE</vendor>
	<name>Leap</name>
	<version>15.3</version>
	<release>2</releas
`), 0777)

	s.NoError(err)

	fr := []entities.FactRequest{
		{
			Name:     "invalid_xml",
			Gatherer: "products@v1",
			CheckID:  "check1",
		},
	}

	gatherer := gatherers.NewProductsGatherer(fs, s.productsPath)

	results, err := gatherer.Gather(fr)
	s.Nil(results)
	s.EqualError(err, "fact gathering error: products-file-reading-error - "+
		"error reading the products file: baseproduct: could not parse product file: "+
		"xml.Decoder.Token() - XML syntax error on line 8: unexpected EOF")
}

func (s *ProductsGathererSuite) TestProductsGathererSuccess() {
	fs := afero.NewMemMapFs()

	err := afero.WriteFile(fs, "/etc/products.d/baseproduct", []byte(`
<?xml version="1.0" encoding="UTF-8"?>
<product schemeversion="0">
	<vendor>openSUSE</vendor>
	<name>Leap</name>
	<version>15.3</version>
	<release>2</release>
	<urls>
		<url name="releasenotes">http://doc.opensuse.org/release-notes-openSUSE.rpm</url>
	</urls>
</product>
`), 0777)

	err = afero.WriteFile(fs, "/etc/products.d/otherproduct", []byte(`
<?xml version="1.0" encoding="UTF-8"?>
<product schemeversion="0">
	<vendor>openSUSE</vendor>
	<name>Other</name>
	<version>15.5</version>
	<release>1</release>
</product>
`), 0777)

	s.NoError(err)

	fr := []entities.FactRequest{
		{
			Name:     "products",
			Gatherer: "products@v1",
			CheckID:  "check1",
		},
	}

	gatherer := gatherers.NewProductsGatherer(fs, s.productsPath)

	expectedFacts := []entities.Fact{
		{
			Name:    "products",
			CheckID: "check1",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"baseproduct": &entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"product": &entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"schemeversion": &entities.FactValueInt{Value: 0},
									"vendor":        &entities.FactValueString{Value: "openSUSE"},
									"name":          &entities.FactValueString{Value: "Leap"},
									"version":       &entities.FactValueString{Value: "15.3"},
									"release":       &entities.FactValueString{Value: "2"},
									"urls": &entities.FactValueMap{
										Value: map[string]entities.FactValue{
											"url": &entities.FactValueList{
												Value: []entities.FactValue{
													&entities.FactValueMap{
														Value: map[string]entities.FactValue{
															"name":  &entities.FactValueString{Value: "releasenotes"},
															"#text": &entities.FactValueString{Value: "http://doc.opensuse.org/release-notes-openSUSE.rpm"},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					"otherproduct": &entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"product": &entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"schemeversion": &entities.FactValueInt{Value: 0},
									"vendor":        &entities.FactValueString{Value: "openSUSE"},
									"name":          &entities.FactValueString{Value: "Other"},
									"version":       &entities.FactValueString{Value: "15.5"},
									"release":       &entities.FactValueString{Value: "1"},
								},
							},
						},
					},
				},
			},
			Error: nil,
		},
	}

	results, err := gatherer.Gather(fr)
	s.NoError(err)
	s.EqualValues(expectedFacts, results)

}
