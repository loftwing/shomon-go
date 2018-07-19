package main

import (
	"flag"
	"log"

	"github.com/loftwing/shomon-go/shomon"
)

//This is just here for testing right now, eventually will be main entrypoint
func main() {
	ptrConfigPath := flag.String("config", "config.json", "Path to alternate config file")
	flag.Parse()

	// if *ptrConfigPath != "config.json" {
	// 	cwd, err := os.Getwd()
	// 	if err != nil {
	// 		log.Panic(err)
	// 	}
	// }

	log.Println("Starting monitor with configpath: ", *ptrConfigPath)
	mon := shomon.NewMonitor(*ptrConfigPath)
	if err := mon.Status(); err != nil {
		log.Println(err)
	}
}
