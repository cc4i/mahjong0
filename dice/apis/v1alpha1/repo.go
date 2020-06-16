package v1alpha1

type Repo struct {
	Tiles []TileMetadata `json:"tiles"`
	Hu []HuMetadata `json:"hus"`
}

type License int
const (
	Apache2 Category = iota
	MIT
)

func (l License) LicenseString() string {
	return [...]string{"Apache2.0", "MIT"}[l]
}

type TileMetadata struct {
	Name string `json:"name"`
	Version string `json:"version"`
	Description string `json:"description"`
	TileRepo string `json:"tileRepo"`
	VersionTag string `json:"versionTag"`
	Author string `json:"author,omitempty"`
	Email string `json:"email,omitempty"`
	License string `json:"license,omitempty"`
	Dependencies []TileMetadata `json:"dependencies,omitempty"`
}

type HuMetadata struct {
	Name string `json:"name"`
	Version string `json:"version"`
	Description string `json:"description"`
	RawUrl string `json:"rawUrl"`
	Author string `json:"author,omitempty"`
	Email string `json:"email,omitempty"`
	License string `json:"license,omitempty"`
	Dependencies []TileMetadata `json:"dependencies"`
}
