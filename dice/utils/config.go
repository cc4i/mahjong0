package utils

// DiceConfig includes all key configuration.
type DiceConfig struct {
	WorkHome   string // WorkHome is the main working folder for all activities
	Region     string // Region is where the S3 bucket is
	BucketName string // BucketName is the name S3 bucket

	/*
		prod - Production and looking for Tiles in remote repo only.
		dev - Local development and looking for Tiles in local & remote repo.
	*/
	Mode string // Mode represents mode for different purpose: dev/prod.

	LocalRepo string // LocalRepo is folder to store Tiles on 'dev' mode

}
