package v1alpha1

import (
	"context"
	valid "github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
	yamlv2 "gopkg.in/yaml.v2"
	"sigs.k8s.io/yaml"
	"strings"
)

var deploymentSchema = "file://./schema/deployment-schema.json"
var tileSchema = "file://./schema/tile-schema.json"

// Data as []byte
type Data []byte

// ParserCore has all parsing functions for Tile/Deployment
type ParserCore interface {
	ParseTile(ctx context.Context) (*Tile, error)
	ParseDeployment(ctx context.Context) (*Deployment, error)
	ValidateTile(ctx context.Context, tile *Tile) error
	ValidateDeployment(ctx context.Context, deployment *Deployment) error
	CheckParameter(ctx context.Context, deployment *Deployment) error
}

// ParseTile parse Tile
func (d *Data) ParseTile(ctx context.Context) (*Tile, error) {
	var tile Tile
	mapSlice := yamlv2.MapSlice{}
	if err := yamlv2.Unmarshal(*d, &mapSlice); err != nil {
		log.Errorf("Unmarshal mapSlice yaml error : %s\n", err)
		return &tile, errors.New("tile specification was invalid")
	}
	// Order for outputs
	var outputsOrder []string
	for _, item := range mapSlice {
		if item.Key == "spec" {
			if spec, ok := item.Value.(yamlv2.MapSlice); ok {
				for _, s := range spec {
					if s.Key == "outputs" {
						if outputs, ok := s.Value.([]interface{}); ok {
							for _, output := range outputs {
								if o, ok := output.(yamlv2.MapSlice); ok {
									for _, it := range o {
										if it.Key == "name" {
											outputsOrder = append(outputsOrder, it.Value.(string))
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	if outputsOrder == nil {
		return &tile, errors.New("tile specification was invalid without ordered outputs")
	}

	err := yaml.UnmarshalStrict(*d, &tile)
	if err != nil {
		return &tile, err
	}
	//Attache outputs order
	tile.Spec.OutputsOrder = outputsOrder
	return &tile, d.ValidateTile(ctx, &tile)
}

// ParseDeployment parse Deployment
func (d *Data) ParseDeployment(ctx context.Context) (*Deployment, error) {
	var deployment Deployment
	mapSlice := yamlv2.MapSlice{}
	if err := yamlv2.Unmarshal(*d, &mapSlice); err != nil {
		log.Errorf("Unmarshal mapSlice yaml error : %s\n", err)
		return &deployment, errors.New("deployment specification was invalid")
	}

	// Retrieve original order as presenting in the file
	var originalOrder []string
	for _, item := range mapSlice {
		if item.Key == "spec" {
			if spec, ok := item.Value.(yamlv2.MapSlice); ok {
				for _, s := range spec {
					if s.Key == "template" {
						if template, ok := s.Value.(yamlv2.MapSlice); ok {
							for _, t := range template {
								if t.Key == "tiles" {
									if tiles, ok := t.Value.(yamlv2.MapSlice); ok {
										for _, tile := range tiles {
											originalOrder = append(originalOrder, tile.Key.(string))
										}
									} else {
										return &deployment, errors.New("deployment specification was invalid : tiles ")
									}

								}
							}
						} else {
							return &deployment, errors.New("deployment specification was invalid : template")
						}

					}
				}
			} else {
				return &deployment, errors.New("deployment specification was invalid : spec")
			}

		}
	}
	if len(originalOrder) < 1 {
		return &deployment, errors.New("deployment specification didn't include tiles")
	}
	////

	if err := yaml.UnmarshalStrict(*d, &deployment); err != nil {
		log.Errorf("Unmarshal yaml error : %s\n", err)
		return &deployment, err
	}
	//obj, err := runtime.Decode(scheme.Codecs.UniversalDeserializer(), *d)
	//if err != nil {
	//	return &deployment, err
	//}
	//deployment, ok := obj.(Deployment)
	//if obj!=nil {
	//	return nil, fmt.Errorf("expected to decode object of type %T; got %T", &Deployment{}, deployment)
	//}
	// Attache original order
	deployment.OriginalOrder = originalOrder

	return &deployment, d.ValidateDeployment(ctx, &deployment)
}

// ValidateTile validates Tile as per tile-spec.yaml
func (d *Data) ValidateTile(ctx context.Context, tile *Tile) error {
	// Validate json schema
	schemaLoader := gojsonschema.NewReferenceLoader(tileSchema)
	jsonLoader := gojsonschema.NewGoLoader(tile)
	result, err := gojsonschema.Validate(schemaLoader, jsonLoader)
	if err != nil {
		log.Errorf("Failed to load schema : %s\n", err)
		return err
	}
	if result.Valid() {
		log.Printf("The document is valid\n")
	} else {
		log.Printf("The document is not valid. see errors :\n")

		for _, ret := range result.Errors() {
			// Err implements the ResultError interface
			log.Printf("- %s\n", ret)
			err = errors.Wrap(errors.New(ret.String()), ret.Description())
		}
		return err

	}
	// Validate as per annotation
	_, err = valid.ValidateStruct(tile)
	return err
}

// ValidateDeployment validate Deployment as per deployment-spec.yaml
func (d *Data) ValidateDeployment(ctx context.Context, deployment *Deployment) error {
	// Validate json schema
	schemaLoader := gojsonschema.NewReferenceLoader(deploymentSchema)
	jsonLoader := gojsonschema.NewGoLoader(deployment)
	result, err := gojsonschema.Validate(schemaLoader, jsonLoader)
	if err != nil {
		log.Errorf("Failed to load schema : %s\n", err)
		return err
	}
	if result.Valid() {
		log.Printf("The document is valid\n")
	} else {
		log.Printf("The document is not valid. see errors :\n")

		for _, ret := range result.Errors() {
			// Err implements the ResultError interface
			log.Printf("- %s\n", ret)
			err = errors.Wrap(errors.New(ret.String()), ret.Description())
		}
		return err

	}
	// Validate as per annotation
	_, err = valid.ValidateStruct(deployment)
	if err != nil {
		return err
	}
	// Check parameters if need to be replaced
	return d.CheckParameter(ctx, deployment)
}

func (d *Data) CheckParameter(ctx context.Context, deployment *Deployment) error {
	parameters := ""
	for _, tile := range deployment.Spec.Template.Tiles {
		for _, input := range tile.Inputs {
			if strings.Contains(input.InputValue, "<<parameter>>") {
				log.Errorf("Input %s needs value : %s", input.Name, input.InputValue)
				parameters = parameters + input.Name + ", "
			}
		}
	}
	if parameters != "" {
		parameters = strings.TrimSuffix(parameters, ", ")
		return errors.New("input parameters: [" + parameters + "] should be replaced by values")
	} else {
		return nil
	}
}
