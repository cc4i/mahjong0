package utils

import (
	"bufio"
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strings"
)

type s3Functions interface {
	LoadTile(tile string, version string) (string, error)
	LoadTileDev(tile string, version string) (string, error)
	LoadSuper() (string, error)
	LoadSuperDev() (string, error)
	Decompress(tile string, version string) error
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var Client HttpClient

func (s3 *DiceConfig) LoadTile(tile string, version string) (string, error) {
	if s3.Mode == "dev" {
		dest, err := s3.LoadTileDev(tile, version)
		if err != nil {
			dest, err = s3.LoadTileS3(tile, version)
		}
		return dest, err
	} else {
		return s3.LoadTileS3(tile, version)
	}
}

func initHttpClient() {
	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	tr := &http.Transport{TLSClientConfig: config}
	Client = &http.Client{Transport: tr}

}

func (s3 *DiceConfig) LoadTileS3(tile string, version string) (string, error) {
	//https://<bucket-name>.s3-<region>.amazonaws.com/tiles-repo/<tile name>/<tile version>/<tile name>.tgz
	tileUrl := fmt.Sprintf("https://%s.s3-%s.amazonaws.com/tiles-repo/%s/%s/%s.tgz",
		s3.BucketName,
		s3.Region,
		strings.ToLower(tile),
		version,
		strings.ToLower(tile))
	if Client == nil {
		initHttpClient()
	}
	destDir := s3.WorkHome + "/super/lib/" + strings.ToLower(tile)
	tileSpecFile := destDir + "/tile-spec.yaml"

	req, err := http.NewRequest(http.MethodGet, tileUrl, nil)
	resp, err := Client.Do(req)
	if err != nil {
		log.Printf("API call was failed from %s with Err: %s. \n", tileUrl, err)
		return tileSpecFile, err
	}

	return tileSpecFile, UnTarGz(destDir, bufio.NewReader(resp.Body))

}

func (s3 *DiceConfig) LoadTileDev(tile string, version string) (string, error) {

	repoDir := s3.LocalRepo
	srcDir := repoDir + "/" + strings.ToLower(tile) + "/" + strings.ToLower(version)
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

func (s3 *DiceConfig) LoadSuper() (string, error) {
	destDir := s3.WorkHome + "/super"
	if f, err := os.Stat(destDir); err==nil && f.IsDir() {
		os.RemoveAll(destDir)
	}

	if s3.Mode == "dev" {
		dest, err := s3.LoadSuperDev()
		if err != nil {
			dest, err = s3.LoadSuperS3()
		}
		return dest, err
	} else {
		return s3.LoadSuperS3()
	}
}

func (s3 *DiceConfig) LoadSuperS3() (string, error) {
	//https://<bucket-name>.s3-<region>.amazonaws.com/tiles-repo/<tile name>/<tile version>/<tile name>.tgz
	tileUrl := fmt.Sprintf("https://%s.s3-%s.amazonaws.com/tiles-repo/%s/%s.tgz",
		s3.BucketName,
		s3.Region,
		"super",
		"super")
	if Client == nil {
		initHttpClient()
	}
	destDir := s3.WorkHome + "/super"

	req, err := http.NewRequest(http.MethodGet, tileUrl, nil)
	resp, err := Client.Do(req)
	if err != nil {
		log.Printf("API call was failed from %s with Err: %s. \n", tileUrl, err)
		return destDir, err
	}

	return destDir, UnTarGz(destDir, bufio.NewReader(resp.Body))

}

func (s3 *DiceConfig) LoadSuperDev() (string, error) {
	repoDir := s3.LocalRepo + "/super"
	destDir := s3.WorkHome + "/super"
	return destDir, Copy(repoDir, destDir)
}
