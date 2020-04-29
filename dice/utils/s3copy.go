package utils

type s3Functions interface {
	PullTile(tile string, version string) error
	PullSuper() error
}
