package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/loftwing/shomon-go/shomon"
	"gopkg.in/ns3777k/go-shodan.v3/shodan"
)

// This is just here for testing right now, eventually will be main entrypoint
func main() {
	ptrConfigPath := flag.String("config", "config.json", "Path to alternate config file")
	flag.Parse()

	log.Println("Starting monitor with configpath: ", *ptrConfigPath)
	mon := shomon.NewMonitor(*ptrConfigPath)
	c := mon.ShodanClient
	c.SetDebug(true)
	mon.Status()
	mon.RegisterAlerts()

	for {
		firehose := make(chan *shodan.HostData)

		err := c.GetBannersByAlerts(context.Background(), firehose)
		if err != nil || firehose == nil {
			panic(err)
		}
		for {
			banner, ok := <-firehose
			if !ok {
				log.Println("channel closed")
				time.Sleep(time.Second * 600)
				break
			}

			log.Printf("%+v\n", banner)

			err := mon.SendBannerEmail(banner)
			if err != nil {
				log.Println("Failed to send email: ", err)
			}
		}

	}
}
