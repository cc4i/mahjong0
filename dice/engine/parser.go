package engine


import (
	"sigs.k8s.io/yaml"
)

type Data []byte

type Deployment struct {
	ApiVersion string `json:"apiVersion"`
	Kind string `json:"kind"`
	Metadata Metadata `json:"metadata"`
	Spec DeploymentSpec `json:"spec"`
}


type DeploymentSpec struct {
	template string `json:"template"`
}

type DeploymentTemplate struct {
	Network []TileSpec `json:"network"`
	Compute []TileSpec `json:"compute"`
	Container []TileSpec `json:"container"`
	Database []TileSpec `json:"database"`
	Application []TileSpec `json:"application"`
	Analysis []TileSpec `json:"analysis"`
	ML []TileSpec `json:"ml"`

}

type Tile struct {
	ApiVersion string `json:"apiVersion"`
	Kind string `json:"kind"`
	Metadata Metadata `json:"metadata"`
	Spec TileSpec `json:"spec"`

}


type Metadata struct {
	Name string	`json:"name"`
	Category string `json:"category"`
	Version string `json:"version"`
}

type TileSpec struct {
	Inputs []TileInput `json:"inputs"`
	Outputs []TileOutput `json:"outputs"`
	Notes []string `json:"notes"`
}

type TileInput struct {
	Name string `json:"name"`
	InputType string `json:"inputType"`
	Description string `json:"description"`
	Dependencies []TileInputDependency `json:"dependencies"`
	DefaultValue string `json:"defaultValue"`
	Manifests []string `json:"manifests"`
	Require string `json:"require"` // yes/no
}

type TileInputDependency struct {
	tile string `json:"tile"`
	version string `json:"version"`
	field string `json:"field"`
}

type TileOutput struct {
	Name string	`json:"name"`
	OutputType string `json:"outputType"`
	Description string `json:"description"`

}


type ParserCore interface {
	ParseTile()(*Tile, error)
	ParseDeployment()(*Deployment, error)
}


func (d *Data) ParseTile()(*Tile, error) {
	var tile Tile

	err := yaml.Unmarshal(*d, &tile)
	if err != nil {
		return &tile, err
	}

	return &tile, nil
}

func (d *Data) ParseDeployment()(*Deployment, error) {
	var deployment Deployment

	err := yaml.Unmarshal(*d, &deployment)
	if err != nil {
		return &deployment, err
	}

	return &deployment, nil
}
