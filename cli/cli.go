package cli

import (
	"os"
	"path"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

var Version string

func Run() {

	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "zanecloud apiserver"
	app.Version = Version
	app.Author = "zhengtao.wuzt"
	app.Email = "zhengtao.wuzt@gmail.com"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "log-level, l",
			Value:  "info",
			EnvVar: "LOG_LEVEL",
			Usage:  "Log level (options: debug, info, warn, error, fatal, panic)",
		},
	}

	app.Before = func(c *cli.Context) error {
		logrus.SetOutput(os.Stderr)
		level, err := logrus.ParseLevel(c.String("log-level"))
		if err != nil {
			logrus.Fatalf(err.Error())
		}
		logrus.SetLevel(level)
		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:  startCommandName,
			Usage: "start a zanecloud apiserver ",
			Flags: []cli.Flag{
				flRedisAddr,
				flMgoUrls,
				flMgoDB,
				flAddr,
				flPort,
				flRootDir,
				//flClusterTls,
				//flClusterTlsKeyFile,
				//flClusterTlsCertFile,
			},
			Action: startCommand,
		},
		{
			Name:  initCommandName,
			Usage: "init root user",
			Flags: []cli.Flag{
				flMgoUrls,
				flMgoDB,
			},
			Action: initCommand,
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
