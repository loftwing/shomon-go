package main

import (
	"flag"
	"log"
	"time"

	"github.com/loftwing/shomon-go/shomon"
)

// This is just here for testing right now, eventually will be main entrypoint
func main() {
	ptrConfigPath := flag.String("config", "config.json", "Path to alternate config file")
	isLearning := flag.Bool("learning", false, "Learning mode")
	isDebug := flag.Bool("debug", false, "Debug mode")
	flag.Parse()

	log.Println("Starting monitor with configpath: ", *ptrConfigPath)
	mon := shomon.NewMonitor(*ptrConfigPath, *isLearning, *isDebug)
	mon.Status()
	mon.RegisterAlerts()

	for {
		firehose := mon.Start()

		for {
			banner, ok := <-firehose
			if !ok {
				log.Println("channel closed")
				time.Sleep(time.Second * 600)
				break
			} else {
				mon.ProcessBanner(banner)
				err := mon.SendBannerEmail(banner)
				if err != nil {
					log.Println("Failed to send email: ", err)
				}
			}
		}
	}
}
