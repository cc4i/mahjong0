package utils

import (
	"log"
	"strings"
)

type S3Config struct {
	WorkHome string
	Region string
	BucketName string
	Mode string // mode for develop purpose: dev/prod
	LocalRepo string // folder to store Tiles on 'dev' mode
}

type s3Functions interface {
	LoadTile(tile string, version string) (string, error)
	LoadTileDev(tile string, version string) (string, error)
	LoadSuper() (string, error)
	LoadSuperDev() (string, error)
	Decompress(tile string, version string) error

}

func (s3 *S3Config) LoadTile(tile string, version string) (string, error) {
	if s3.Mode == "dev" {
		return s3.LoadTileDev(tile, version)
	} else {
		// TODO: not quite yet
		return "", nil
	}
}

func (s3 *S3Config) LoadTileDev(tile string, version string) (string, error) {

	repoDir := s3.LocalRepo
	srcDir := repoDir + "/"+strings.ToLower(tile) + "/" + strings.ToLower(version)
	destDir := s3.WorkHome + "/super/lib/" + strings.ToLower(tile)
	tileSpecFile := destDir + "/tile-spec.yaml"
	log.Printf("Load Tile < %s - %s > ... from < %s >\n", tile, version, s3.LocalRepo)

	return tileSpecFile, Copy(srcDir, destDir,
		Options{
			OnSymlink: func(src string) SymlinkAction {
				return Skip
			},
			Skip: func(src string) bool {
				return strings.Contains(src, "node_modules")
			},
		})

}

func (s3 *S3Config) LoadSuper() (string, error) {
	if s3.Mode == "dev" {
		return s3.LoadSuperDev()
	} else {
		// TODO: not quite yet
		return "", nil
	}
}


func (s3 *S3Config) LoadSuperDev() (string, error) {
	repoDir := s3.LocalRepo + "/super"
	destDir := s3.WorkHome+"/super"
	return destDir,  Copy(repoDir, destDir)
}