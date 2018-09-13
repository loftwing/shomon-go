package main

import (
	"context"
	"flag"
	"log"

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
	// Buffered channel to move banners from the shodan stream to the Consume function above
	firehose := make(chan *shodan.HostData)

	err := c.GetBannersByAlerts(context.Background(), firehose)
	if err != nil {
		panic(err)
	}

	for {
		banner, ok := <-firehose
		if !ok {
			log.Println("channel closed")
		}

		log.Printf("%+v\n", banner)

		mon.SendBannerEmail(banner)
	}
}
