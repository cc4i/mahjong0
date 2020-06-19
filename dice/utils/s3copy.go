package utils

import (
	"bufio"
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type S3Functions interface {
	LoadTile(tile string, version string) (string, error)
	LoadTileDev(tile string, version string) (string, error)
	LoadTileS3(tile string, version string, folder string) (string, error)
	LoadSuper(folder string) (string, error)
	LoadSuperDev(folder string) (string, error)
	LoadSuperS3(folder string) (string, error)
	Decompress(tile string, version string) error
	LoadTestOutput(tile string) ([]byte, error)
	CleanJunk()
	LoadTileSpec(tile string, version string) ([]byte, error)
	LoadTileSpecS3(tile string, version string) ([]byte, error)
	LoadTileSpecDev(tile string, version string) ([]byte, error)
	LoadHuSpec(hu string) ([]byte, error)
	LoadHuSpecS3(hu string) ([]byte, error)
	LoadHuSpecDev(hu string) ([]byte, error)
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var Client HttpClient

func (dc *DiceConfig) LoadTile(tile string, version string, folder string) (string, error) {
	if dc.Mode == "dev" {
		dest, err := dc.LoadTileDev(tile, version, folder)
		if err != nil {
			dest, err = dc.LoadTileS3(tile, version, folder)
		}
		return dest, err
	} else {
		return dc.LoadTileS3(tile, version, folder)
	}
}

func initHttpClient() {
	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	tr := &http.Transport{TLSClientConfig: config}
	Client = &http.Client{Transport: tr}

}

func (dc *DiceConfig) LoadTileS3(tile string, version string, folder string) (string, error) {
	//https://<bucket-name>.s3-<region>.amazonaws.com/tiles-repo/<tile name>/<tile version>/<tile name>.tgz
	tileUrl := fmt.Sprintf("https://%s.s3-%s.amazonaws.com/tiles-repo/%s/%s/%s.tgz",
		dc.BucketName,
		dc.Region,
		strings.ToLower(tile),
		version,
		strings.ToLower(tile))
	if Client == nil {
		initHttpClient()
	}
	destDir := dc.WorkHome + folder + "/lib/" + strings.ToLower(tile)
	tileSpecFile := destDir + "/tile-spec.yaml"

	req, err := http.NewRequest(http.MethodGet, tileUrl, nil)
	resp, err := Client.Do(req)
	if err != nil {
		log.Printf("API call was failed from %s with Err: %s. \n", tileUrl, err)
		return tileSpecFile, err
	}

	return tileSpecFile, UnTarGz(destDir, bufio.NewReader(resp.Body))

}

func (dc *DiceConfig) LoadTileDev(tile string, version string, folder string) (string, error) {

	repoDir := dc.LocalRepo
	srcDir := repoDir + "/" + strings.ToLower(tile) + "/" + strings.ToLower(version)
	destDir := dc.WorkHome + folder + "/lib/" + strings.ToLower(tile)
	tileSpecFile := destDir + "/tile-spec.yaml"
	log.Printf("Load Tile < %s - %s > ... from < %s >\n", tile, version, dc.LocalRepo)

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

// CleanJunk removes all *.log / *.sh under super-*/
func (dc *DiceConfig) CleanJunk(folder string) {
	destDir := dc.WorkHome + folder
	if f, err := os.Stat(destDir); err == nil && f.IsDir() {
		for _, suffix := range []string{"*.sh", "*.log"} {
			files, err := filepath.Glob(destDir + "/" + suffix)
			if err == nil {
				for _, f := range files {
					os.Remove(f)
				}
			}
		}
		os.RemoveAll(destDir + "/lib")
	}

}
func (dc *DiceConfig) LoadSuper(folder string) (string, error) {

	dc.CleanJunk(folder)
	if dc.Mode == "dev" {
		dest, err := dc.LoadSuperDev(folder)
		if err != nil {
			dest, err = dc.LoadSuperS3(folder)
		}
		return dest, err
	} else {
		return dc.LoadSuperS3(folder)
	}
}

func (dc *DiceConfig) LoadSuperS3(folder string) (string, error) {
	//https://<bucket-name>.s3-<region>.amazonaws.com/tiles-repo/<tile name>/<tile version>/<tile name>.tgz
	tileUrl := fmt.Sprintf("https://%s.s3-%s.amazonaws.com/tiles-repo/%s/%s.tgz",
		dc.BucketName,
		dc.Region,
		"super",
		"super")
	if Client == nil {
		initHttpClient()
	}
	destDir := dc.WorkHome + folder

	req, err := http.NewRequest(http.MethodGet, tileUrl, nil)
	resp, err := Client.Do(req)
	if err != nil {
		log.Printf("API call was failed from %s with Err: %s. \n", tileUrl, err)
		return destDir, err
	}

	return destDir, UnTarGz(destDir, bufio.NewReader(resp.Body))

}

func (dc *DiceConfig) LoadSuperDev(folder string) (string, error) {
	repoDir := dc.LocalRepo + "/super"
	destDir := dc.WorkHome + folder
	return destDir, Copy(repoDir, destDir)
}

func (dc *DiceConfig) LoadTestOutput(tile string, folder string) ([]byte, error) {
	testOutputFile := dc.WorkHome + folder + "/lib/" + strings.ToLower(tile) + "/test/" + tile + ".output"
	f, err := os.Open(testOutputFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func (dc *DiceConfig) LoadTileSpec(tile string, version string) ([]byte, error) {
	if dc.Mode == "dev" {
		dest, err := dc.LoadTileSpecDev(tile, version)
		if err != nil {
			dest, err = dc.LoadTileSpecS3(tile, version)
		}
		return dest, err
	} else {
		return dc.LoadTileSpecS3(tile, version)
	}
}

func (dc *DiceConfig) LoadTileSpecDev(tile string, version string) ([]byte, error) {
	repoDir := dc.LocalRepo
	srcDir := repoDir + "/" + strings.ToLower(tile) + "/" + strings.ToLower(version)
	tileSpecFile := srcDir + "/tile-spec.yaml"
	log.Printf("Load Tile < %s - %s > ... from < %s >\n", tile, version, dc.LocalRepo)

	return ioutil.ReadFile(tileSpecFile)

}

func (dc *DiceConfig) LoadTileSpecS3(tile string, version string) ([]byte, error) {
	tileUrl := fmt.Sprintf("https://%s.s3-%s.amazonaws.com/tiles-repo/%s/%s/tile-spec.yaml",
		dc.BucketName,
		dc.Region,
		strings.ToLower(tile),
		version)
	if Client == nil {
		initHttpClient()
	}
	req, err := http.NewRequest(http.MethodGet, tileUrl, nil)
	resp, err := Client.Do(req)
	if err != nil {
		log.Printf("API call was failed from %s with Err: %s. \n", tileUrl, err)
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}

func (dc *DiceConfig) LoadHuSpec(hu string) ([]byte, error) {
	if dc.Mode == "dev" {
		dest, err := dc.LoadHuSpecDev(hu)
		if err != nil {
			dest, err = dc.LoadHuSpecS3(hu)
		}
		return dest, err
	} else {
		return dc.LoadHuSpecS3(hu)
	}
}
func (dc *DiceConfig) LoadHuSpecDev(hu string) ([]byte, error) {
	repoDir := dc.LocalRepo
	srcDir := repoDir + "../templates/"
	huSpecFile := srcDir + strings.ToLower(hu) + ".yaml"

	return ioutil.ReadFile(huSpecFile)
}
func (dc *DiceConfig) LoadHuSpecS3(hu string) ([]byte, error) {
	tileUrl := fmt.Sprintf("https://%s.s3-%s.amazonaws.com/templates/%s.yaml",
		dc.BucketName,
		dc.Region,
		strings.ToLower(hu))
	if Client == nil {
		initHttpClient()
	}
	req, err := http.NewRequest(http.MethodGet, tileUrl, nil)
	resp, err := Client.Do(req)
	if err != nil {
		log.Printf("API call was failed from %s with Err: %s. \n", tileUrl, err)
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}
