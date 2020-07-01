package engine

import (
	"dice/apis/v1alpha1"
	"github.com/stretchr/testify/assert"
	"testing"
)

var dSid = []string{"1000-1000-1000", "1000-1000-1001", "1000-1000-1002"}
var tilesGrid1 = TilesGrid{
	TileInstance:        "tileInstance01",
	ExecutableOrder:     2,
	TileName:            "EKS",
	TileVersion:         "0.10.0",
	TileCategory:        v1alpha1.ContainerProvider.CString(),
	RootTileInstance:    "tileInstance01",
	ParentTileInstances: []string{"tileInstance03"},
}
var tilesGrid2 = TilesGrid{
	TileInstance:        "tileInstance02",
	ExecutableOrder:     1,
	TileName:            "Network",
	TileVersion:         "0.11.0",
	TileCategory:        v1alpha1.Network.CString(),
	RootTileInstance:    "tileInstance01",
	ParentTileInstances: []string{"tileInstance01"},
}
var tilesGrid3 = TilesGrid{
	TileInstance:        "tileInstance03",
	ExecutableOrder:     3,
	TileName:            "ArgoCD",
	TileVersion:         "0.12.0",
	TileCategory:        v1alpha1.ContainerApplication.CString(),
	RootTileInstance:    "tileInstance01",
	ParentTileInstances: []string{"root"},
}
var tilesGrid4 = TilesGrid{
	TileInstance:        "tileInstance04",
	ExecutableOrder:     3,
	TileName:            "Bumblebee",
	TileVersion:         "0.14.0",
	TileCategory:        v1alpha1.Application.CString(),
	RootTileInstance:    "tileInstance04",
	ParentTileInstances: []string{"root"},
}
var tilesGridMap1 = make(map[string]*TilesGrid)
var tilesGridMap2 = make(map[string]*TilesGrid)
var tilesGridMap3 = make(map[string]*TilesGrid)

func init() {
	AllTs[dSid[0]] = Ts{
		AllTilesN: map[string]*v1alpha1.Tile{
			"tileInstance01": &v1alpha1.Tile{},
			"tileInstance02": &v1alpha1.Tile{},
			"tileInstance03": &v1alpha1.Tile{},
			"tileInstance04": &v1alpha1.Tile{},
		},
	}
	AllTs[dSid[1]] = Ts{
		AllTilesN: map[string]*v1alpha1.Tile{
			"tileInstance01": &v1alpha1.Tile{},
			"tileInstance02": &v1alpha1.Tile{},
			"tileInstance03": &v1alpha1.Tile{},
			"tileInstance04": &v1alpha1.Tile{},
		},
	}
	AllTs[dSid[2]] = Ts{
		AllTilesN: map[string]*v1alpha1.Tile{
			"tileInstance01": &v1alpha1.Tile{},
			"tileInstance02": &v1alpha1.Tile{},
			"tileInstance03": &v1alpha1.Tile{},
			"tileInstance04": &v1alpha1.Tile{},
		},
	}
	tilesGridMap1[tilesGrid1.TileInstance] = &tilesGrid1
	tilesGridMap1[tilesGrid2.TileInstance] = &tilesGrid2
	tilesGridMap1[tilesGrid3.TileInstance] = &tilesGrid3
	tilesGridMap1[tilesGrid4.TileInstance] = &tilesGrid4
	AllTilesGrids[dSid[0]] = &tilesGridMap1
	AllTilesGrids[dSid[1]] = &tilesGridMap1
	AllTilesGrids[dSid[2]] = &tilesGridMap1

}

func TestAllDependentTiles(t *testing.T) {
	tiles := AllDependentTiles(dSid[0], "tileInstance03")
	assert.Equal(t, 1, len(tiles))
}

func TestFamilyTileInstance(t *testing.T) {
	tileName := FamilyTileInstance(dSid[0], "tileInstance03")
	assert.Equal(t, 3, len(tileName))
}

func TestIsDuplicatedCategory(t *testing.T) {

}
