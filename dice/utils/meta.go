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

func HusMetadata(ctx context.Context, bucket string) ([]v1alpha1.HuMetadata, error) {
	var hus []v1alpha1.HuMetadata
	tiles, err := AllTiles(ctx, bucket)
	if err != nil {
		return nil, err
	}
	session, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-1"),
	})
	if err != nil {
		return nil, err
	}
	svc := s3.New(session)

	input := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String("templates"),
	}
	err = svc.ListObjectsPages(input, func(output *s3.ListObjectsOutput, b bool) bool {
		for _, obj := range output.Contents {

			if strings.Contains(*obj.Key, ".yaml") {
				log.Println("Object:", *obj.Key)
				objInput := &s3.GetObjectInput{
					Bucket: aws.String(bucket),
					Key:    obj.Key,
				}
				objOutput, err := svc.GetObject(objInput)
				if err != nil {
					log.Error(err)
					continue
				}

				buf, err := ioutil.ReadAll(objOutput.Body)
				d := v1alpha1.Data(buf)
				deploy, err := d.ParseDeployment(ctx)
				if err != nil {
					log.Errorf("parsing %s with error : %s\n", *obj.Key, err)
					continue
				}
				hu := v1alpha1.HuMetadata{
					Name:        deploy.Metadata.Name,
					Version:     deploy.Metadata.Version,
					Description: deploy.Metadata.Description,
					RawUrl:      deploy.Metadata.TileRepo,
					Author:      deploy.Metadata.Author,
					Email:       deploy.Metadata.Email,
					License:     deploy.Metadata.License,
					Released:    deploy.Metadata.Released,
				}
				for _, t := range deploy.Spec.Template.Tiles {
					dt := dependentTiles(tiles, t.TileReference, t.TileVersion)
					dt[t.TileReference] = t.TileVersion
					var dtm []v1alpha1.TileMetadata
					for k, v := range dt {
						if tile, ok := tiles[k+"-"+v]; ok {
							tm := v1alpha1.TileMetadata{
								Name:        tile.Metadata.Name,
								Version:     tile.Metadata.Version,
								Category:    tile.Metadata.Category,
								Description: tile.Metadata.Description,
								TileRepo:    tile.Metadata.TileRepo,
								VersionTag:  tile.Metadata.Version,
								Author:      tile.Metadata.Author,
								Email:       tile.Metadata.Email,
								License:     tile.Metadata.License,
								Released:    tile.Metadata.Released,
							}
							dtm = append(dtm, tm)
						}
					}
					hu.Dependencies = dtm
				}
				hus = append(hus, hu)

			}
		}
		return true
	})
	return hus, nil
}

func TilesMetadata(ctx context.Context, bucket string) ([]v1alpha1.TileMetadata, error) {

	var meta = make(map[string]*v1alpha1.TileMetadata)
	tiles, err := AllTiles(ctx, bucket)
	if err != nil {
		return nil, err
	}
	for _, tile := range tiles {
		meta[tile.Metadata.Name+"-"+tile.Metadata.Version] = &v1alpha1.TileMetadata{
			Name:        tile.Metadata.Name,
			Version:     tile.Metadata.Version,
			Category:    tile.Metadata.Category,
			Description: tile.Metadata.Description,
			TileRepo:    tile.Metadata.TileRepo,
			VersionTag:  tile.Metadata.Version,
			Author:      tile.Metadata.Author,
			Email:       tile.Metadata.Email,
			License:     tile.Metadata.License,
			Released:    tile.Metadata.Released,
		}
	}
	return addDependencies(tiles, meta), err
}

func AllTiles(ctx context.Context, bucket string) (map[string]v1alpha1.Tile, error) {
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
					Key:    obj.Key,
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
				tiles[tile.Metadata.Name+"-"+tile.Metadata.Version] = *tile
			}
		}
		return true
	})

	return tiles, err
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

// dependentTiles returns map: TileName->TileVersion
func dependentTiles(tiles map[string]v1alpha1.Tile, tileName string, tileVersion string) map[string]string {
	var td = make(map[string]string)
	if tile, ok := tiles[tileName+"-"+tileVersion]; ok {
		for _, d := range tile.Spec.Dependencies {
			td[d.TileReference] = d.TileVersion
			if subTile, ok := tiles[d.TileReference+"-"+d.TileVersion]; ok {
				if len(subTile.Spec.Dependencies) > 0 {
					m := dependentTiles(tiles, subTile.Metadata.Name, subTile.Metadata.Version)
					for k, v := range m {
						td[k] = v
					}
				}
			}
		}
	}
	return td
}
