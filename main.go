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

	if sm.ShodanClient.Debug {
		log.Printf("Startup Rcpt: %s\n", strings.Join(sm.Config.Email.To, ","))
	}

	m := gomail.NewMessage()
	m.SetHeader("From", sm.Config.Email.From)
	m.SetHeader("Subject", "ShoMon started.")

	body := fmt.Sprintf(`
<h1>ShoMon started with %d known services:</h1> <br>
============================ <br>
<b>Learning?: </b> %t <br>
`, len(sm.Known), sm.Learning)

	for _, v := range sm.Known {
		svcd := fmt.Sprintf(`
============================ <br>
<b>Name: </b> %s <br>
<b>IP:</b> %s <br>
<b>Port:</b> %d <br>
<b>Transport: </b> %s <br>
`, v.Name, v.IP, v.Port, v.Transport)
		body = body + svcd
	}

	m.SetBody("text/html", body)

	d := &gomail.Dialer{
		Port: 25,
		Auth: nil,
		Host: sm.Config.Email.Server,
		SSL:  false,
	}

	addrs := make([]string, len(sm.Config.Email.To))
	for i, rcpt := range sm.Config.Email.To {
		addrs[i] = m.FormatAddress(rcpt, "")
	}

	m.SetHeader("To", addrs...)

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
	if err := SendStartupEmail(mon); err != nil {
		log.Println(err)
	}

	for {
		firehose := mon.Start()

		for {
			banner, ok := <-firehose
			if !ok {
				log.Println("channel closed")
				time.Sleep(time.Second * 600)
				break
			} else {
				err := mon.ProcessBanner(banner)
				if err != nil {
					log.Println("Error processing banner: ", err)
				}
			}
		}
	}
}
