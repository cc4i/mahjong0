package engine

import "container/list"

// Ts is key struct to fulfil super.ts template and key element to generate execution plan.
type Ts struct {
	// TsLibs
	TsLibs []TsLib
	// TsLibsMap : TileName -> TsLib
	TsLibsMap map[string]TsLib
	// TsStacks
	TsStacks []*TsStack
	// TsStacksMap : TileName -> TsStack
	TsStacksMap map[string]*TsStack
	// TsStacksOrder is an order of execution./ tileName ...> ....>
	TsStacksOrder *list.List

	// AllTiles : "Category-TileName" -> Tile
	AllTiles map[string]Tile
	// AllOutputs :  TileName -> TsOutput
	AllOutputs map[string]*TsOutput
}

type TsLib struct {
	TileName     string
	TileVersion  string
	TileConstructName string
	TileFolder   string
	TileCategory string
}

type TsStack struct {
	TileName          string
	TileVersion       string
	TileConstructName string
	TileVariable      string
	TileStackName     string
	TileStackVariable string
	TileCategory      string
	// inputName -> TsInputParameter
	InputParameters   map[string]TsInputParameter
	TsManifests       *TsManifests
	EnvList map[string]string //caching all env

}

type TsInputParameter struct {
	InputName  string
	InputValue string
	IsOverrideField string
}

type TsManifests struct {
	ManifestType         string
	Namespace            string
	Files                []string
	Folders              []string
	VendorService        string
	DependentTile        string
	DependentTileVersion string
}

type TsOutput struct {
	TileName    string
	TileVersion string
	StageName   string
	TsOutputs   map[string]*TsOutputDetail //OutputName -> TsOutputDetail
}

type TsOutputDetail struct {
	Name                string
	OutputType          string
	DefaultValue        string
	DefaultValueCommand string
	OutputValue         string
	Description         string
}

// AllTs represents all information about tiles, input, output, etc.,  session-id -> Ts
var AllTs = make(map[string]Ts)
