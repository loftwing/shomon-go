package shomon

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"gopkg.in/gomail.v2"
	"gopkg.in/ns3777k/go-shodan.v3/shodan"
)

// ShodanMon todo
type ShodanMon struct {
	ShodanClient *shodan.Client
	Config       *Config
}

// Config defines json structure for config file
type Config struct {
	Shodan struct {
		APIKey   string            `json:"apikey"`
		Networks map[string]string `json:"networks"`
	} `json:"shodan"`
	Email struct {
		Server string   `json:"server"`
		From   string   `json:"from"`
		To     []string `json:"to"`
	} `json:"email"`
}

// NewMonitor creates a new ShodanMon and returns it
func NewMonitor(configpath string) *ShodanMon {
	conf := loadConfig(configpath)
	newClient := shodan.NewClient(nil, conf.Shodan.APIKey)
	return &ShodanMon{
		ShodanClient: newClient,
		Config:       conf}
}

func (sm *ShodanMon) checkAlert(name string) bool {
	c := sm.ShodanClient
	if rAlerts, err := c.GetAlerts(nil); err != nil {
		log.Panic("Couldnt get alerts.")
	} else {
		for _, ra := range rAlerts {
			if ra.Name == name {
				return true
			}
		}
	}
	return false
}

// RegisterAlerts loops through configured alerts, and registers those that are not registered
func (sm *ShodanMon) RegisterAlerts() {
	c := sm.ShodanClient
	cAlerts := sm.Config.Shodan.Networks

	for n, b := range cAlerts {
		if prs := sm.checkAlert(n); !prs {
			log.Println("Adding: ", n, b)
			if _, err := c.CreateAlert(nil, n, []string{b}, 0); err != nil {
				log.Println("Failed to register alert: ", n, b)
			}
		}
	}
}

func loadConfig(file string) *Config {
	var config Config
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		log.Panic("Could not open config file!: ", err)
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return &config
}

// Send a single banner by smtp
func (sm *ShodanMon) SendBannerEmail(b *shodan.HostData) error {
	m := gomail.NewMessage()
	m.SetHeader("From", sm.Config.Email.From)
	m.SetHeader("To", sm.Config.Email.To[0])
	m.SetHeader("Subject", "ShoMon: Service Found")
	body := fmt.Sprintf(`
<b>IP:</b> %s
<b>Port:</b> %d
<b>Transport:</b> %s
<b>Title:</b> %s
<b>Opts:</b>
%+v`, string(b.IP), b.Port, b.Transport, b.Title, b.Opts)

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

// Status prints current status of monitor to logger, or returns an error
func (sm *ShodanMon) Status() {
	c := sm.ShodanClient
	log.Println("Monitor Status")

	log.Println("======PROFILE======")
	if profile, err := c.GetAccountProfile(nil); err != nil {
		log.Println("Error pulling profile info from API: ", err)
	} else {
		log.Printf("%+v\n", profile)
	}

	log.Println("======ALERTS======")
	if alerts, err := c.GetAlerts(nil); err != nil {
		log.Println("Error pulling alert info from API.")
	} else {
		for _, a := range alerts {
			log.Printf("%+v\n", a)
		}
	}
}
