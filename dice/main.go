package main

import (
	"context"
	"dice/web"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Fatal(web.Router(context.Background()).Run("0.0.0.0:9090"))
}
