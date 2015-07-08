package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/mistify-agent-ovs"
	logx "github.com/mistifyio/mistify-logrus-ext"
	flag "github.com/ogier/pflag"
)

func main() {
	// Handle cli flags
	var port uint
	var bridge, logLevel string
	flag.UintVarP(&port, "port", "p", 40001, "listen port")
	flag.StringVarP(&bridge, "bridge", "b", "mistify0", "bridge to join interfaces to with OVS")
	flag.StringVarP(&logLevel, "log-level", "l", "warning", "log level: debug/info/warning/error/critical/fatal")
	flag.Parse()

	// Set up logging
	if err := logx.DefaultSetup(logLevel); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "logx.DefaultSetup",
		}).Fatal("Could not set up logging")
	}

	o := ovs.NewOVS(bridge)
	// Run HTTP Server
	if err := o.RunHTTP(port); err != nil {
		os.Exit(1)
	}
}
