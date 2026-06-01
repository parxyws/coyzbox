package main

import (
	"github.com/parxyws/cozybox/internal/config"
	"github.com/parxyws/cozybox/pkg/logger"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.InitAppConfig()
	if err != nil {
		logrus.Fatal(err)
	}

	logrus := logger.NewLogger(cfg)

}
