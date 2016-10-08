// Command line interface to run Cross Post Detector
package main

import (
	"flag"
	"fmt"
	"os"
	"github.com/xpd-org/xpd"
)

const defaultConfigFile = "xpd.yml"

type Params struct {
	configfile string
}

func parseArgs() Params {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options]\n\n", os.Args[0])
		flag.PrintDefaults()
	}

	config := flag.String("config", defaultConfigFile, "path to configuration file")
	flag.Parse()

	return Params{*config}
}

func main() {
	params := parseArgs()

	xpd.Run(params.configfile)
}
