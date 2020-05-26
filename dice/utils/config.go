package utils

// DiceConfig includes all key configuration.
type DiceConfig struct {
	// WorkHome is the main working folder for all activities
	WorkHome   string
	// Region is where the S3 bucket is
	Region     string
	// BucketName is the name S3 bucket
	BucketName string
	// Mode represents mode for different purpose: dev/prod.
	// prod - Production and looking for Tiles in remote repo only.
	// dev - Local development and looking for Tiles in local & remote repo.
	Mode       string
	// LocalRepo is folder to store Tiles on 'dev' mode
	LocalRepo  string

}
