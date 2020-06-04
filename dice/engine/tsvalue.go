package engine

import (
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"
)

// TilesGrid represents relationship table of all Tile for each deployment
type TilesGrid struct {
	TileInstance       string 	// TileInstance is unique ID of Tile as instance
	ExecutableOrder    int    	// ExecutableOrder is execution order of Tile
	TileName           string 	// TileName is the name of Tile
	TileVersion        string 	// TileVersion is the version of Tile
	TileCategory       string 	// TileCategory is the category of Tile
	RootTileInstance   string 	// RootTileInstance indicates Tiles are in the same group
	ParentTileInstance string	// ParentTileInstance indicates who's dependent on Me - Tile
}


// Ts is key struct to fulfil super.ts template and key element to generate execution plan.
type Ts struct {
	TsLibs []TsLib // TsLibs is only for go template
	TsLibsMap map[string]TsLib // TsLibsMap : TileName -> TsLib
	TsStacks []*TsStack // TsStacks is only for go template
	TsStacksMapN map[string]*TsStack	// TsStacksMap: TileInstance -> TsStack, all initialized values will be store here, include input, env, etc
	AllTilesN map[string]*Tile // AllTiles: TileInstance -> Tile
	AllOutputsN map[string]*TsOutput // AllOutputs:  TileInstance ->TsOutput, all output values will be store here
	CreatedTime time.Time // Created time
}

// Ts represents all referred CDK resources
type TsLib struct {
	TileInstance string
	TileName          string
	TileVersion       string
	TileConstructName string
	TileFolder        string
	TileCategory      string
}

// TsStack represents all detail for each stack
type TsStack struct {
	TileInstance string
	TileName          string
	TileVersion       string
	TileConstructName string
	TileVariable      string
	TileStackName     string
	TileStackVariable string
	TileCategory      string
	InputParameters   map[string]TsInputParameter //input name -> TsInputParameter
	TsManifests       *TsManifests
}

// TsInputParameter
type TsInputParameter struct {
	InputName       string
	InputValue      string
	IsOverrideField string
	DependentTileInstance string
	DependentTileInputName string
}

// TsManifests
type TsManifests struct {
	ManifestType         string
	Namespace            string
	Files                []string
	Folders              []string
	TileInstance string

}

// TsOutput
type TsOutput struct {
	TileName    string
	TileVersion string
	StageName   string
	TsOutputs   map[string]*TsOutputDetail //OutputName -> TsOutputDetail
}

// TsOutputDetail
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

// DependentEKSTile return dependent EKS tile in the the group
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

// AllDependentTiles return all dependent Tiles
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

// IsDuplicatedCategory determine if it's duplicated Tile under same group
func IsDuplicatedCategory(dSid string, rootTileInstance string, tileCategory string) bool {
	if allTG, ok := AllTilesGrids[dSid]; ok {
		for _, v := range *allTG {
			if v.RootTileInstance == rootTileInstance {
				if v.TileCategory == tileCategory {
					return true
				}
			}
		}
	}
	return false
}

// ReferencedTsStack return referred TsStack
func ReferencedTsStack(dSid string, rootTileInstance string, tileName string) *TsStack {
	if allTG, ok := AllTilesGrids[dSid]; ok {
		for _, v := range *allTG {
			if v.RootTileInstance == rootTileInstance {
				if v.TileName == tileName {
					if ts, ok  := AllTs[dSid]; ok {
						return ts.TsStacksMapN[v.TileInstance]
					}
				}
			}
		}
	}
	return nil
}

// ValueRef return actual value of referred input/output
func ValueRef(dSid string, ref string, ti string) (string, error) {
	re := regexp.MustCompile(`^\$\(([[:alnum:]]*\.[[:alnum:]]*\.[[:alnum:]]*)\)$`)
	ms := re.FindStringSubmatch(ref)
	if len(ms)==2 {
		str := strings.Split(ms[1], ".")
		tileInstance := str[0]
		where := str[1]
		field := str[2]
		if tileInstance == "self" && ti != "" {
			tileInstance = ti
		}
		if at, ok := AllTs[dSid]; ok {

				switch where {
				case "inputs":
					if tsStack, ok := at.TsStacksMapN[tileInstance]; ok {
						for _, input := range tsStack.InputParameters {
							if field == input.InputName {
								return input.InputValue, nil
							}
						}
					}
				case "outputs":
					if outputs, ok := at.AllOutputsN[tileInstance]; ok {
						for name, output := range outputs.TsOutputs {
							if name == field {
								return output.OutputValue, nil
							}
						}
					}

				}
		}

	} else {
		return "", errors.New("expression: "+ref+" was error")
	}
	return "", errors.New("referred value wasn't exist")
}

// ParentTileInstance return Tile instance of parent Tile
func ParentTileInstance(dSid string, tileInstance string) string {
	if allTG, ok := AllTilesGrids[dSid]; ok {
		if tg, ok := (*allTG)[tileInstance]; ok {
			return tg.ParentTileInstance
		}
	}
	return ""
}


func CDKAllValueRef(dSid string, str string) (string, error) {
	for {
		re := regexp.MustCompile(`^.*\$cdk\(([[:alnum:]]*\.[[:alnum:]]*\.[[:alnum:]]*)\).*$`)
		s := re.FindStringSubmatch(str)
		//
		if len(s) == 2 {
			if def := strings.Split(s[1], "."); len(def)!=3 {
				return "", errors.New("error cdk reference : " + s[1])
			} else {
				tileInstance := def[0]
				tileName := def[1]
				field := def[2]
				str = strings.ReplaceAll(str, "$cdk("+s[1]+")", CDKValueRef(dSid, tileInstance, tileName, field))
			}

		} else {
			break
		}
	}
	return str, nil
}

func CDKValueRef(dSid string, tileInstance string, tileName string, field string) string {

	rootTileInstance := RootTileInstance(dSid, tileInstance)
	ts := ReferencedTsStack(dSid, rootTileInstance, tileName)
	if ts != nil {
		return ts.TileStackVariable+"."+ts.TileVariable+"."+field
	}

	return ""
}

// RootTileInstance return root tileInstance as per tileInstance
func RootTileInstance(dSid string, tileInstance string) string {
	if allTG, ok := AllTilesGrids[dSid]; ok {
		for _, v := range *allTG {
			if v.TileInstance == tileInstance {
				return v.RootTileInstance
			}
		}
	}
	return ""
}

// FamilyTileInstance array of tileInstance as per tileInstance
func FamilyTileInstance(dSid string, tileInstance string) []string {
	rootTileInstance := RootTileInstance(dSid, tileInstance)
	if allTG, ok := AllTilesGrids[dSid]; ok {
		var tileInstances []string
		for _, v := range *allTG {
			if v.RootTileInstance == rootTileInstance {
				tileInstances = append(tileInstances, v.TileInstance)
			}
		}
		return tileInstances
	}
	return nil
}

