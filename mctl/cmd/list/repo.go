package list

import "time"

type License int

const (
	Apache2 License = iota
	MIT
)

func (l License) LicenseString() string {
	return [...]string{"Apache2.0", "MIT"}[l]
}

type TileMetadata struct {
	Name         string         `json:"name"`
	Version      string         `json:"version"`
	Category     string         `json:"category"`
	Description  string         `json:"description"`
	TileRepo     string         `json:"tileRepo"`
	VersionTag   string         `json:"versionTag"`
	Author       string         `json:"author,omitempty"`
	Email        string         `json:"email,omitempty"`
	License      string         `json:"license,omitempty"`
	Dependencies []TileMetadata `json:"dependencies,omitempty"`
	Released     time.Time      `json:"released"`
}

type HuMetadata struct {
	Name         string         `json:"name"`
	Version      string         `json:"version"`
	Description  string         `json:"description"`
	RawUrl       string         `json:"rawUrl"`
	Author       string         `json:"author,omitempty"`
	Email        string         `json:"email,omitempty"`
	License      string         `json:"license,omitempty"`
	Dependencies []TileMetadata `json:"dependencies"`
	Released     time.Time      `json:"released"`
}
