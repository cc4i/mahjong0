package utils

type DiceConfig struct {
	WorkHome   string
	Region     string
	BucketName string
	Mode       string // mode for develop purpose: dev/prod
	LocalRepo  string // folder to store Tiles on 'dev' mode
}
