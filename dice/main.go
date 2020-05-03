package main

import (
	"context"
	"dice/apis"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Fatal(apis.Router(context.TODO()).Run("0.0.0.0:9090"))
}
