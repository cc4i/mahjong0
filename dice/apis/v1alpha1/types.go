package v1alpha1

import "time"

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
	Region string `json:"region"`
	Profile string `json:"profile"`
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
	Description string `json:"description,omitempty"`
	TileRepo string `json:"tileRepo,omitempty"`
	VersionTag string `json:"versionTag,omitempty"`
	Author string `json:"author,omitempty"`
	Email string `json:"email,omitempty"`
	License string `json:"license,omitempty"`
	Released time.Time `json:"released,omitempty"`
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
