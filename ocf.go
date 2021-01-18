package main

import (
	"free5gc/src/app"
	"free5gc/src/ocf/logger"
	"free5gc/src/ocf/service"
	"free5gc/src/ocf/version"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var OCF = &service.OCF{}

var appLog *logrus.Entry

func init() {
	appLog = logger.AppLog
}

func main() {
	app := cli.NewApp()
	app.Name = "ocf"
	appLog.Infoln(app.Name)
	appLog.Infoln("OCF version: ", version.GetVersion())
	app.Usage = "-free5gccfg common configuration file -amfcfg ocf configuration file"
	app.Action = action
	app.Flags = OCF.GetCliCmd()
	if err := app.Run(os.Args); err != nil {
		logger.AppLog.Errorf("OCF Run error: %v", err)
	}
}

func action(c *cli.Context) {
	app.AppInitializeWillInitialize(c.String("free5gccfg"))
	OCF.Initialize(c)
	OCF.Start()
}
