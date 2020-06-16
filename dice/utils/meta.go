package utils

import (
	"context"
	"dice/apis/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"strings"
)

func TilesInRepo(ctx context.Context, bucket string) ([]v1alpha1.TileMetadata, error) {

	var meta = make(map[string]*v1alpha1.TileMetadata)
	var tiles = make(map[string]v1alpha1.Tile)
	session, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-1"),
	})
	if err != nil {
		return nil, err
	}
	svc := s3.New(session)
	input := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String("tiles-repo"),
	}

	err = svc.ListObjectsPages(input, func(output *s3.ListObjectsOutput, b bool) bool {
		for _, obj := range output.Contents {

			if strings.Contains(*obj.Key, "tile-spec.yaml") {
				log.Println("Object:", *obj.Key)
				objInput := &s3.GetObjectInput{
					Bucket: aws.String(bucket),
					Key: obj.Key,
				}
				objOutput, err := svc.GetObject(objInput)
				if err != nil {
					log.Error(err)
					continue
				}

				buf, err := ioutil.ReadAll(objOutput.Body)
				d := v1alpha1.Data(buf)
				tile, err := d.ParseTile(ctx)
				if err != nil {
					log.Errorf("parsing %s with error : %s\n", *obj.Key, err)
					continue
				}
				log.Infof("%s : %s \n", tile.Metadata.Name, tile.Metadata.Version)
				tiles[tile.Metadata.Name+"-"+tile.Metadata.Version]=*tile
				meta[tile.Metadata.Name+"-"+tile.Metadata.Version]=&v1alpha1.TileMetadata{
					Name: tile.Metadata.Name,
					Version: tile.Metadata.Version,
					Description: "---",
					TileRepo: "https://git",
					VersionTag: "0.1.0",
					Author: "aws",
					Email: "aws@builder.com",
					License: v1alpha1.MIT.LicenseString(),
					Released: *objOutput.LastModified,
				}
			}
		}
		return true
	})

	return addDependencies(tiles, meta),err
}

func addDependencies(tiles map[string]v1alpha1.Tile, meta map[string]*v1alpha1.TileMetadata) []v1alpha1.TileMetadata {
	var ret []v1alpha1.TileMetadata
	for _, m := range meta {
		dts := dependentTiles(tiles, m.Name, m.Version)
		for k, v := range dts {
			m.Dependencies = append(m.Dependencies, *meta[k+"-"+v])
		}
		ret = append(ret, *m)
	}
	return ret
}

func dependentTiles(tiles map[string]v1alpha1.Tile, tileName string, tileVersion string) map[string]string {
	var td = make(map[string]string)
	if tile, ok := tiles[tileName+"-"+tileVersion]; ok {
		for _, d := range tile.Spec.Dependencies {
			td[d.TileReference]= d.TileVersion
			if subTile, ok := tiles[ d.TileReference+"-"+d.TileVersion]; ok {
				if len(subTile.Spec.Dependencies)>0 {
					m := dependentTiles(tiles, subTile.Metadata.Name, subTile.Metadata.Version)
					for k,v := range m {
						td[k]=v
					}
				}
			}
		}
	}
	return td
}