package engine

import (
	"sigs.k8s.io/yaml"
)

type Data []byte

type Deployment struct {
	ApiVersion string         `json:"apiVersion"`
	Kind       string         `json:"kind"`
	Metadata   Metadata       `json:"metadata"`
	Spec       DeploymentSpec `json:"spec"`
}

type DeploymentSpec struct {
	Template DeploymentTemplate `json:"template"`
}

type DeploymentTemplate struct {
	Network     []DeploymentTemplateDetail `json:"network"`
	Compute     []DeploymentTemplateDetail `json:"compute"`
	Container   []DeploymentTemplateDetail `json:"container"`
	Database    []DeploymentTemplateDetail `json:"database"`
	Application []DeploymentTemplateDetail `json:"application"`
	Analysis    []DeploymentTemplateDetail `json:"analysis"`
	ML          []DeploymentTemplateDetail `json:"ml"`
}

type DeploymentTemplateDetail struct {
	TileReference string      `json:"tileReference"`
	TileVersion   string      `json:"tileVersion"`
	Inputs        []TileInput `json:"inputs"`
}

type Tile struct {
	ApiVersion string   `json:"apiVersion"`
	Kind       string   `json:"kind"`
	Metadata   Metadata `json:"metadata"`
	Spec       TileSpec `json:"spec"`
}

type Metadata struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Version  string `json:"version"`
}

type TileSpec struct {
	Dependencies []TileDependency `json:"dependencies"`
	Inputs       []TileInput      `json:"inputs"`
	Manifests    TileManifest     `json:"manifests"`
	Outputs      []TileOutput     `json:"outputs"`
	Notes        []string         `json:"notes"`
}

type TileDependency struct {
	Name          string `json:"name"`
	TileReference string `json:"tileReference"`
	TileVersion   string `json:"tileVersion"`
}

type TileInput struct {
	Name         string                `json:"name"`
	InputType    string                `json:"inputType"`
	Description  string                `json:"description"`
	Dependencies []TileInputDependency `json:"dependencies"`
	DefaultValue string                `json:"defaultValue"`
	DefaultValues []string `json:"defaultValues"`
	InputValue   string                `json:"inputValue"`
	InputValues   string                `json:"inputValues"`
	Manifests    []string              `json:"manifests"`
	Require      string                `json:"require"` // yes/no
}

type TileManifest struct {
	ManifestType string   `json:"manifestType"`
	Files        []string `json:"files"`
	Folders      []string `json:"folders"`
}

type TileInputDependency struct {
	Name  string `json:"name"`
	Field string `json:"field"`
}

type TileOutput struct {
	Name                string `json:"name"`
	OutputType          string `json:"outputType"`
	DefaultValue        string `json:"defaultValue"`
	DefaultValueCommand string `json:"defaultValueCommand"`
	Description         string `json:"description"`
}

type ParserCore interface {
	ParseTile() (*Tile, error)
	ParseDeployment() (*Deployment, error)
}

func (d *Data) ParseTile() (*Tile, error) {
	var tile Tile

	err := yaml.Unmarshal(*d, &tile)
	if err != nil {
		return &tile, err
	}

	return &tile, nil
}

func (d *Data) ParseDeployment() (*Deployment, error) {
	var deployment Deployment

	err := yaml.Unmarshal(*d, &deployment)
	if err != nil {
		return &deployment, err
	}

	return &deployment, nil
}
