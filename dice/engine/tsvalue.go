package engine

import (
	"dice/apis/v1alpha1"
	"dice/utils"
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Status of deployment
type DeploymentStatus int

const (
	Created     DeploymentStatus = iota // Created indicate the deployment just kicks off
	Progress                            // Progress indicate the deployment is running
	Done                                // Done indicate the deployment is Done
	Interrupted                         // Interrupted indicate the deployment stop at somewhere
)

func (c DeploymentStatus) DSString() string {
	return [...]string{"Created", "Progress", "Done", "Interrupted"}[c]
}

// TilesGrid represents relationship table of all Tile for each deployment
type TilesGrid struct {
	TileInstance        string   // TileInstance is unique ID of Tile as instance
	ExecutableOrder     int      // ExecutableOrder is execution order of Tile
	TileName            string   // TileName is the name of Tile
	TileVersion         string   // TileVersion is the version of Tile
	TileCategory        string   // TileCategory is the category of Tile
	RootTileInstance    string   // RootTileInstance indicates Tiles are in the same group
	ParentTileInstances []string // ParentTileInstances indicates who's dependent on Me - Tile
	Status              string   // Status of deployment
}

// DeploymentRecord is a record of each deployment
type DeploymentRecord struct {
	SID         string    // SID is session ID for each deployment
	Name        string    // Name is unique identifier for each deployment: metadata.name
	Created     time.Time // Created time
	Updated     time.Time // Updated time
	SuperFolder string    // Main folder for all stuff per deployment
	Status      string    // Status of deployment
}

// Ts represents all referred CDK resources
type TsLib struct {
	TileInstance      string
	TileName          string
	TileVersion       string
	TileConstructName string
	TileFolder        string
	TileCategory      string
}

// TsStack represents all detail for each stack
type TsStack struct {
	TileInstance      string //Unique name for Tile instance, either given or generated
	TileName          string // Name of Tile
	TileVersion       string // Version of Tile
	TileConstructName string
	TileVariable      string
	TileStackName     string
	TileStackVariable string
	TileCategory      string
	InputParameters   map[string]*TsInputParameter //input name -> TsInputParameter
	TsManifests       *TsManifests
	TileFolder        string // The relative folder for Tile
	Region            string // target region
	Profile           string // specified profile
}

// TsInputParameter
type TsInputParameter struct {
	InputName              string
	InputValue             string
	InputValueForTemplate  string
	InputType              string
	IsOverrideField        string
	DependentTileInstance  string
	DependentTileInputName string
}

// TsManifests
type TsManifests struct {
	ManifestType string
	Namespace    string
	Files        []string
	Folders      []string
	TileInstance string
	Flags        []string
}

// TsOutput
type TsOutput struct {
	TileName     string
	TileVersion  string
	StageName    string
	OutputsOrder []string                    //Extract data as per order
	TsOutputs    *map[string]*TsOutputDetail //OutputName -> TsOutputDetail
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

// Ts is key struct to fulfil super.ts template and key element to generate execution plan.
type Ts struct {
	DR           *DeploymentRecord         // DR is a deployment record
	TsLibs       []TsLib                   // TsLibs is only for go template
	TsLibsMap    map[string]TsLib          // TsLibsMap : TileName -> TsLib
	TsStacks     []*TsStack                // TsStacks is only for go template
	TsStacksMapN map[string]*TsStack       // TsStacksMap: TileInstance -> TsStack, all initialized values will be store here, include input, env, etc
	AllTilesN    map[string]*v1alpha1.Tile // AllTiles: TileInstance -> Tile
	AllOutputsN  *map[string]*TsOutput     // AllOutputs:  TileInstance ->TsOutput, all output values will be store here
}

// AllTs represents all information about tiles, input, output, etc.,  id(uuid) -> Ts
var AllTs = make(map[string]Ts)

// AllTilesGrid store all Tiles relationship, id(uuid) -> (tile-instance -> TilesGrid)
var AllTilesGrids = make(map[string]*map[string]*TilesGrid)

// AllPlans store all execution plan, id(uuid) -> ExecutionPlan
var AllPlans = make(map[string]*ExecutionPlan)

// SortedTilesGrid return sorted TilesGrid array from AllTilesGrid
func SortedTilesGrid(dSid string) []TilesGrid {
	if allTG, ok := AllTilesGrids[dSid]; ok {
		var tg []TilesGrid
		for _, v := range *allTG {
			tg = append(tg, *v)
		}
		sort.SliceStable(tg, func(i, j int) bool {
			return tg[i].ExecutableOrder < tg[j].ExecutableOrder
		})
		return tg
	}
	return nil
}

// DependentEKSTile return dependent EKS tile in the the group
func DependentEKSTile(dSid string, tileInstance string) *v1alpha1.Tile {

	pTileInstance := ParentTileInstance(dSid, tileInstance)
	if allTG, ok := AllTilesGrids[dSid]; ok {
		for _, v := range *allTG {
			if utils.Contains(pTileInstance, v.TileInstance) {
				if at, ok := AllTs[dSid]; ok {
					if tile, ok := at.AllTilesN[v.TileInstance]; ok {
						if tile.Metadata.VendorService == v1alpha1.EKS.VSString() {
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
func AllDependentTiles(dSid string, tileInstance string) []v1alpha1.Tile {

	if allTG, ok := AllTilesGrids[dSid]; ok {
		var tiles []v1alpha1.Tile
		for _, v := range *allTG {
			if utils.Contains(v.ParentTileInstances, tileInstance) {
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

func IsDependenciesDone(dSid string, tileInstance string)(string, bool, error) {

	if allTG, ok := AllTilesGrids[dSid]; ok {
		if tg, ok := (*allTG)[tileInstance];ok {

			for _, dp := range tg.ParentTileInstances {
				if ctg, ok := (*allTG)[dp];ok {
					if ctg.Status == Interrupted.DSString() {
						return ctg.TileInstance, false, errors.New(ctg.TileInstance + " was failed to provision")
					} else if  ctg.Status != Done.DSString() {
						return ctg.TileInstance, false, nil
					}
				}
			}

		}
	}
	return "", true, nil
}

// IsDuplicatedTile determine if it's duplicated Tile under same root-tile-instance
func IsDuplicatedTile(dSid string, rootTileInstance string, tileName string) bool {
	if allTG, ok := AllTilesGrids[dSid]; ok {
		for _, v := range *allTG {
			if v.RootTileInstance == rootTileInstance {
				if v.TileName == tileName {
					//strings.Contains(tileInstance, "Generated")
					//utils.ContainsArray(v.ParentTileInstances, parentTileInstances)
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
					if ts, ok := AllTs[dSid]; ok {
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
	if strings.Contains(ref, "$") {
		re := regexp.MustCompile(`^\$\(([[:alnum:]]*\.[[:alnum:]]*\.[[:alnum:]]*)\)$`)
		ms := re.FindStringSubmatch(ref)
		if len(ms) == 2 {
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
					if tileInstance != "self" {
						if tsStack, ok := at.TsStacksMapN[tileInstance]; ok {
							for _, input := range tsStack.InputParameters {
								if field == input.InputName {
									return input.InputValue, nil
								}
							}
						}
					} else {
						//TODO: Any possible value ?! May not right
						for _, tsStack := range at.TsStacksMapN {
							for _, input := range tsStack.InputParameters {
								if field == input.InputName {
									return input.InputValue, nil
								}
							}
						}

					}

				case "outputs":
					if tileInstance != "self" {
						if outputs, ok := (*at.AllOutputsN)[tileInstance]; ok {
							for name, output := range *outputs.TsOutputs {
								if name == field {
									return output.OutputValue, nil
								}
							}
						}
					} else {
						//TODO: Any possible value ?!
						for _, outputs := range *at.AllOutputsN {
							for name, output := range *outputs.TsOutputs {
								if name == field {
									return output.OutputValue, nil
								}
							}
						}
					}

				}
			}

		} else {
			return "", errors.New("expression: " + ref + " was error")
		}
	}
	return ref, nil
}

// ParentTileInstances return Tile instance of parent Tile
func ParentTileInstance(dSid string, tileInstance string) []string {
	if allTG, ok := AllTilesGrids[dSid]; ok {
		if tg, ok := (*allTG)[tileInstance]; ok {
			return tg.ParentTileInstances
		}
	}
	return nil
}

func CDKAllValueRef(dSid string, str string) (string, error) {
	max := strings.Count(str, "$")
	for {
		re := regexp.MustCompile(`^.*\$cdk\(([[:alnum:]]*\.[[:alnum:]]*\.[[:alnum:]]*)\).*$`)
		s := re.FindStringSubmatch(str)
		//
		if len(s) == 2 {
			if def := strings.Split(s[1], "."); len(def) != 3 {
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
		// avoid infinite loop due to replacement failure
		max--
		if max < 0 {
			break
		}
	}
	return str, nil
}

func CDKValueRef(dSid string, tileInstance string, tileName string, field string) string {

	rootTileInstance := RootTileInstance(dSid, tileInstance)
	ts := ReferencedTsStack(dSid, rootTileInstance, tileName)
	if ts != nil {
		return ts.TileStackVariable + "." + ts.TileVariable + "." + field
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

// AllRootTileInstance return all root tileInstance
func AllRootTileInstance(dSid string) []string {
	var root []string
	if allTG, ok := AllTilesGrids[dSid]; ok {
		for _, v := range *allTG {
			if len(v.ParentTileInstances) == 1 && v.ParentTileInstances[0] == "root" {
				root = append(root, v.TileInstance)
			}
		}
	}
	return root
}

// TsContent returns content as per d-sid
func TsContent(sid string) *Ts {
	if ts, ok := AllTs[sid]; ok {
		return &ts
	}
	return nil
}

// AllTsDeployment returns all records of deployment
func AllTsDeployment() []DeploymentRecord {
	var ds []DeploymentRecord
	for _, ts := range AllTs {
		ds = append(ds, *ts.DR)
	}
	return ds
}

// IsRepeatedDeployment return flag of repeated deployment and sid if repeated
func IsRepeatedDeployment(name string) (string, bool) {
	tss := make([]Ts, 0, len(AllTs))
	for _, ts := range AllTs {
		tss = append(tss, ts)
	}
	// By create time (descending)
	sort.SliceStable(tss, func(i, j int) bool {
		return tss[j].DR.Created.Sub(tss[i].DR.Created) < 0
	})

	for _, ts := range tss {
		if ts.DR.Name == name {
			return ts.DR.SID, true
		}
	}
	return "", false

}

// IsProcessed check if tile instance has been processed
func IsProcessed(dSid string, tileInstances []string) bool {
	checkCount := 0
	if allTG, ok := AllTilesGrids[dSid]; ok && allTG != nil {
		for _, ti := range tileInstances {
			if _, ok := (*allTG)[ti]; ok {
				checkCount++
			}
		}
	}
	return checkCount == len(tileInstances)
}
