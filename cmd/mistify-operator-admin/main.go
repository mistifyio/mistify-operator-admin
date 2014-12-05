package main

import (
	"os"

	"local/mistify-operator-admin"
	"local/mistify-operator-admin/config"
	"local/mistify-operator-admin/db"

	flag "github.com/docker/docker/pkg/mflag"
	"github.com/mistifyio/mistify-agent/log"
)

func main() {
	var port uint
	var configFile, logLevel string
	var h bool

	flag.BoolVar(&h, []string{"h", "#help", "-help"}, false, "display the help")
	flag.UintVar(&port, []string{"p", "#port", "-port"}, 15000, "listen port")
	flag.StringVar(&configFile, []string{"c", "#config-file", "-config-file"}, "", "config file")
	flag.StringVar(&logLevel, []string{"l", "#log-level", "-log-level"}, "warning", "log level: debug/info/warning/error/critical/fatal")
	flag.Parse()

	if h {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if err := log.SetLogLevel(logLevel); err != nil {
		log.Fatal(err)
	}

	if configFile == "" {
		log.Fatal("need a config file")
	}

	if err := config.Load(configFile); err != nil {
		log.Fatal(err)
	}

	_, err := db.Connect(nil)
	if err != nil {
		log.Fatal(err)
	}
	operator.Run(port)
}
