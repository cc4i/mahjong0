package engine

import (
	"sort"
	"time"
)


type TilesGrid struct {
	TileInstance string
	ExecutableOrder int
	TileName string
	TileVersion string
	TileCategory string
	rootTileInstance string
	ParentTileInstance string
}


// Ts is key struct to fulfil super.ts template and key element to generate execution plan.
type Ts struct {
	// TsLibs
	TsLibs []TsLib
	// TsLibsMap : TileName -> TsLib
	TsLibsMap map[string]TsLib

	// TsStacks
	TsStacks []*TsStack
	// TsStacksMap : TileInstance -> TsStack ////TileName -> TsStack
	TsStacksMapN map[string]*TsStack

	// AllTiles : TileInstance -> Tile ////"Category-TileName" -> Tile
	AllTilesN map[string]*Tile
	// AllOutputs :  TileInstance ->TsOutput ////TileName -> TsOutput
	AllOutputsN map[string]*TsOutput

	// Created time
	CreatedTime time.Time
}

type TsLib struct {
	TileInstance string
	TileName          string
	TileVersion       string
	TileConstructName string
	TileFolder        string
	TileCategory      string
}

type TsStack struct {
	TileInstance string
	TileName          string
	TileVersion       string
	TileConstructName string
	TileVariable      string
	TileStackName     string
	TileStackVariable string
	TileCategory      string
	InputParameters   map[string]TsInputParameter
	TsManifests       *TsManifests
	// Caching predefined & default ENV
	PredefinedEnv map[string]string
}

type TsInputParameter struct {
	InputName       string
	InputValue      string
	IsOverrideField string
}

type TsManifests struct {
	ManifestType         string
	Namespace            string
	Files                []string
	Folders              []string
	TileInstance string

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

// AllTs represents all information about tiles, input, output, etc.,  id(uuid) -> Ts
var AllTs = make(map[string]Ts)
// AllTilesGrid store all Tiles relationship, id(uuid) -> (tile-instance -> TilesGrid)
var AllTilesGrids = make(map[string]*map[string]TilesGrid)


// SortedTilesGrid return sorted TilesGrid array from AllTilesGrid
func SortedTilesGrid(dSid string) []TilesGrid {
	if allTG, ok := AllTilesGrids[dSid]; ok {
		var tg []TilesGrid
		for _, v := range *allTG {
			tg = append(tg, v)
		}
		sort.SliceStable(tg, func(i, j int) bool {
			return tg[i].ExecutableOrder < tg[j].ExecutableOrder
		})
		return tg
	}
	return nil
}

func DependentEKSTile(dSid string, tileInstance string) *Tile {

	if allTG, ok := AllTilesGrids[dSid]; ok {
		for _, v := range *allTG {
			if v.ParentTileInstance == tileInstance {
				if at, ok := AllTs[dSid]; ok {
					if tile, ok := at.AllTilesN[v.TileInstance]; ok {
						if tile.Metadata.VendorService == EKS.VSString() {
							return tile
						}
					}
				}
			}
		}
	}

	return nil

}

func AllDependentTiles(dSid string, tileInstance string) []Tile {

	if allTG, ok := AllTilesGrids[dSid]; ok {
		var tiles []Tile
		for _, v := range *allTG {
			if v.ParentTileInstance == tileInstance {
				if at, ok := AllTs[dSid]; ok {
					if tile, ok := at.AllTilesN[v.TileInstance]; ok {
						tiles = append(tiles, *tile)
					}
				}
			}
		}
		return tiles
	}
	return nil
}

func IsDuplicatedCategory(dSid string, rootTileInstance string, tileCategory string) bool {
	if allTG, ok := AllTilesGrids[dSid]; ok {
		for _, v := range *allTG {
			if v.rootTileInstance == rootTileInstance {
				if v.TileCategory == tileCategory {
					return true
				}
			}
		}
	}
	return false
}