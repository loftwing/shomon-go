package main

import (
	"flag"
	"fmt"
	"gopkg.in/gomail.v2"
	"log"
	"strings"
	"time"

	"github.com/loftwing/shomon-go/shomon"
)

func SendStartupEmail(sm *shomon.ShodanMon) error {
	m := gomail.NewMessage()
	m.SetHeader("From", sm.Config.Email.From)
	m.SetHeader("To", strings.Join(sm.Config.Email.To, ","))
	m.SetHeader("Subject", "ShoMon started.")
	body := fmt.Sprintf(`
<h1>ShoMon started with %d known services:</h1>
`, len(sm.Known))

	for _, v := range sm.Known {
		svcd := fmt.Sprintf(`
============================ <br>
<b>IP:</b> %s <br>
<b>Port:</b> %d <br>
<b>Transport: </b> %s <br>
`, v.IP, v.Port, v.Transport)
		body = body + svcd
	}

	m.SetBody("text/html", body)

	d := &gomail.Dialer{
		Port: 25,
		Auth: nil,
		Host: sm.Config.Email.Server,
		SSL:  false,
	}

	// Send it
	if err := d.DialAndSend(m); err != nil {
		return err
	} else {
		return nil
	}
}

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
	SendStartupEmail(mon)

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
