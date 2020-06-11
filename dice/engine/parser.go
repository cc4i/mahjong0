package engine

import (
	"context"
	valid "github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
	yamlv2 "gopkg.in/yaml.v2"
	"sigs.k8s.io/yaml"
)

// Data as []byte
type Data []byte

// Enumeration for category of metadata in Tile specification.
type Category int

const (
	Network Category = iota
	Compute
	ContainerProvider
	Storage
	Database
	Application
	ContainerApplication
	Analysis
	ML
)

func (c Category) CString() string {
	return [...]string{"Network", "Compute", "ContainerProvider", "Storage", "Database", "Application", "ContainerApplication", "Analysis", "ML"}[c]
}

// Enumeration for services' vendor
type VendorService int

const (
	EKS VendorService = iota
	ECS
	Kubernetes
	Kops
)

func (vs VendorService) VSString() string {
	return [...]string{"EKS", "ECS", "Kubernetes", "Kops"}[vs]
}

// Enumeration for input/output in Tile specification
type IOType int

const (
	String IOType = iota
	Number
	CDKObject
	FromCommand
)

func (iot IOType) IOTString() string {
	return [...]string{"String", "Number", "CDKObject", "FromCommand"}[iot]
}

// Manifest for manifest type in Tile specification
type ManifestType int

const (
	K8s ManifestType = iota
	Helm
	Kustomize
)

func (mts ManifestType) MTString() string {
	return [...]string{"K8s", "Helm", "Kustomize"}[mts]
}

// Deployment specification
type Deployment struct {
	ApiVersion    string         `json:"apiVersion" jsonschema:"required" valid:"in(mahjong.io/v1alpha1)"`
	Kind          string         `json:"kind" jsonschema:"required" valid:"in(Deployment)"`
	Metadata      Metadata       `json:"metadata" jsonschema:"required"`
	Spec          DeploymentSpec `json:"spec" jsonschema:"required"`
	OriginalOrder []string       `json:"originalOrder,omitempty"` // Stored TileInstance, keep original order as same as in yaml
}

// DeploymentSpec deployment.spec
type DeploymentSpec struct {
	Template DeploymentTemplate `json:"template" jsonschema:"required"`
	Summary  DeploymentSummary  `json:"summary"`
}

// DeploymentTemplate deployment.spec.template
type DeploymentTemplate struct {
	Tiles map[string]DeploymentTemplateDetail `json:"tiles" jsonschema:"required"` // Tiles represent all to be deployed Tiles
}

// DeploymentSummary deployment.spec.summary
type DeploymentSummary struct {
	Description string                    `json:"description"`
	Outputs     []DeploymentSummaryOutput `json:"outputs"`
	Notes       []string                  `json:"notes"`
}

// DeploymentSummaryOutput deployment.spec.summary.outputs
type DeploymentSummaryOutput struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// DeploymentTemplateDetail deployment.spec.template
type DeploymentTemplateDetail struct {
	TileReference string       `json:"tileReference" `
	TileVersion   string       `json:"tileVersion" `
	DependsOn     string       `json:"dependsOn"`
	Inputs        []TileInput  `json:"inputs"`
	Manifests     TileManifest `json:"manifests"`
}

// Tile specification
type Tile struct {
	TileInstance string   `json:"tileInstance"`
	ApiVersion   string   `json:"apiVersion" valid:"in(mahjong.io/v1alpha1)"`
	Kind         string   `json:"kind" valid:"in(Tile)"`
	Metadata     Metadata `json:"metadata" `
	Spec         TileSpec `json:"spec" `
}

// Metadata for Tile & Deployment
type Metadata struct {
	Name                     string `json:"name"`
	Category                 string `json:"category"`
	VendorService            string `json:"vendorService,omitempty"`
	DependentOnVendorService string `json:"dependentOnVendorService,omitempty"`
	Version                  string `json:"version"`
}

// TileSpec tile.spec
type TileSpec struct {
	Global       GlobalDetail     `json:"global"`
	PreRun       PreRunDetail     `json:"preRun"`
	Dependencies []TileDependency `json:"dependencies"`
	Inputs       []TileInput      `json:"inputs"`
	Manifests    TileManifest     `json:"manifests"`
	Outputs      []TileOutput     `json:"outputs" `
	PostRun      PostRunDetail    `json:"PostRun"`
	Notes        []string         `json:"notes"`
}

// GlobalDetail tile.spec.global
type GlobalDetail struct {
	Env []GlobalDetailEnv `json:"env"`
}

// GlobalDetailEnv tile.spec.global.env
type GlobalDetailEnv struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// PreRunDetail tile.spec.preRun
type PreRunDetail struct {
	Stages []PreRunStages `json:"stages"`
}

// PreRunStages tile.spec.preRun.stage
type PreRunStages struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

// TileDependency tile.spec.dependencies
type TileDependency struct {
	Name          string `json:"name"`
	TileReference string `json:"tileReference"`
	TileVersion   string `json:"tileVersion"`
}

// TileInput tile.spec.input
type TileInput struct {
	Name          string                `json:"name"`
	InputType     string                `json:"inputType"`
	Description   string                `json:"description"`
	Dependencies  []TileInputDependency `json:"dependencies"`
	DefaultValue  string                `json:"defaultValue"`
	DefaultValues []string              `json:"defaultValues"`
	InputValue    string                `json:"inputValue"`
	InputValues   []string              `json:"inputValues"`
	Require       bool                  `json:"require"` // true/false
	Override      TileInputOverride     `json:"override"`
}

// TileInputOverride tile.spec.input.override
type TileInputOverride struct {
	Name  string `json:"name"`
	Field string `json:"field"`
	//InputName     string
	OverrideValue string
}

// TileManifest tile.spec.manifest
type TileManifest struct {
	ManifestType string   `json:"manifestType"`
	Namespace    string   `json:"namespace"`
	Files        []string `json:"files"`
	Folders      []string `json:"folders"`
}

// TileInputDependency tile.spec.input.dependency
type TileInputDependency struct {
	Name  string `json:"name"`
	Field string `json:"field"`
}

// TileOutput tile.spec.output
type TileOutput struct {
	Name                string `json:"name"`
	OutputType          string `json:"outputType"`
	DefaultValue        string `json:"defaultValue"`
	DefaultValueCommand string `json:"defaultValueCommand"`
	Description         string `json:"description"`
}

// PostRunDetail tile.spec.postRun
type PostRunDetail struct {
	Stages []PostRunStages `json:"stages"`
}

// PostRunStages tile.spec.postRun.stage
type PostRunStages struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

// ParserCore has all parsing functions for Tile/Deployment
type ParserCore interface {
	ParseTile(ctx context.Context) (*Tile, error)
	ParseDeployment(ctx context.Context) (*Deployment, error)
	ValidateTile(ctx context.Context, tile *Tile) error
	ValidateDeployment(ctx context.Context, deployment *Deployment) error
}

// ParseTile parse Tile
func (d *Data) ParseTile(ctx context.Context) (*Tile, error) {
	var tile Tile

	err := yaml.UnmarshalStrict(*d, &tile)
	if err != nil {
		return &tile, err
	}

	return &tile, d.ValidateTile(ctx, &tile)
}

// ParseDeployment parse Deployment
func (d *Data) ParseDeployment(ctx context.Context) (*Deployment, error) {
	var deployment Deployment
	mapSlice := yamlv2.MapSlice{}
	if err := yamlv2.Unmarshal(*d, &mapSlice); err != nil {
		log.Errorf("Unmarshal mapSlice yaml error : %s\n", err)
		return &deployment, errors.New(" Deployment specification was invalid")
	}

	// Retrieve original order as presenting in the file
	var originalOrder []string
	for _, item := range mapSlice {
		if item.Key == "spec" {
			spec := item.Value.(yamlv2.MapSlice)
			for _, s := range spec {
				if s.Key == "template" {
					template := s.Value.(yamlv2.MapSlice)
					for _, t := range template {
						if t.Key == "tiles" {
							tiles := t.Value.(yamlv2.MapSlice)
							for _, tile := range tiles {
								originalOrder = append(originalOrder, tile.Key.(string))
							}
						}
					}
				}
			}
		}
	}
	if len(originalOrder) < 1 {
		return &deployment, errors.New(" Deployment specification didn't include tiles")
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
	//TODO find a better way to verify yaml
	_, err := valid.ValidateStruct(tile)
	return err
}

// ValidateDeployment validate Deployment as per deployment-spec.yaml
func (d *Data) ValidateDeployment(ctx context.Context, deployment *Deployment) error {
	//TODO find a better way to verify yaml
	schemaLoader := gojsonschema.NewReferenceLoader("file://./schema/deployment-schema.json")
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
	_, err = valid.ValidateStruct(deployment)
	return err
}
