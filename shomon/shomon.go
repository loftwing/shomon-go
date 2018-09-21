package shomon

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"gopkg.in/gomail.v2"
	"gopkg.in/ns3777k/go-shodan.v3/shodan"
)

// ShodanMon todo
type ShodanMon struct {
	ShodanClient *shodan.Client
	Config       *Config
	ConfigPath   string
	Known        []Service
	Learning     bool
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
	Known []Service `json:"known"`
}

// Service reps a single service
type Service struct {
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	Transport string `json:"transport"`
}

// NewMonitor creates a new ShodanMon and returns it
func NewMonitor(configpath string, isLearning, isDebug bool) *ShodanMon {
	conf := loadConfig(configpath)
	newClient := shodan.NewClient(nil, conf.Shodan.APIKey)
	if isDebug {
		newClient.SetDebug(true)
	}

	return &ShodanMon{
		ShodanClient: newClient,
		Config:       conf,
		ConfigPath:   configpath,
		Known:        conf.Known,
		Learning:     isLearning,
	}
}

func (sm *ShodanMon) writeServiceToConfig(s *Service) error {
	sm.Config.Known = append(sm.Config.Known, *s)
	if nc, err := json.MarshalIndent(sm.Config, "", "    "); err != nil {
		return err
	} else {
		err := ioutil.WriteFile(sm.ConfigPath, nc, 0644)
		if err != nil {
			return err
		} else {
			return nil
		}
	}
}

func (sm *ShodanMon) Start() chan *shodan.HostData {
	nc := make(chan *shodan.HostData)
	err := sm.ShodanClient.GetBannersByAlerts(context.Background(), nc)
	if err != nil {
		log.Panic("Couldnt start shomon firehose!: ", err)
	}
	return nc
}

func (sm *ShodanMon) ProcessBanner(h *shodan.HostData) {
	DescribeBanner(h)

	s := Service{
		IP:        h.IP.String(),
		Port:      h.Port,
		Transport: h.Transport,
	}

	if !sm.IsKnown(s) {
		sm.AddKnown(s)
	}
}

func (sm *ShodanMon) IsKnown(s Service) bool {
	known := false
	for _, v := range sm.Known {
		if v.Transport == s.Transport && v.IP == s.IP && v.Port == s.Port {
			known = true
		}
	}

	return known
}

func DescribeBanner(h *shodan.HostData) {
	log.Println("========================")
	log.Printf("IP: %s\n", string(h.IP))
	log.Printf("Port: %d\n", h.Port)
	log.Printf("Transport: %s\n", h.Transport)
}

func (sm *ShodanMon) AddKnown(s Service) {
	if sm.Learning {
		if err := sm.writeServiceToConfig(&s); err != nil {
			log.Println("Couldnt write service to config: ", err)
		}
	}

	sm.Known = append(sm.Known, s)
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
	m.SetHeader("To", strings.Join(sm.Config.Email.To, ","))
	m.SetHeader("Subject", "ShoMon: Service Found")
	body := fmt.Sprintf(`
<b>IP:</b> %s <br>
<b>Port:</b> %d <br>
<b>Transport:</b> %s <br>
<b>Title:</b> %s <br>
<b>Opts:</b> <br>
%+v`, b.IP.String(), b.Port, b.Transport, b.Title, b.Opts)

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

	asciiArt := `
_____/\\\\\\\\\\\____/\\\________________________/\\\\____________/\\\\_____________________________        
 ___/\\\/////////\\\_\/\\\_______________________\/\\\\\\________/\\\\\\_____________________________       
  __\//\\\______\///__\/\\\_______________________\/\\\//\\\____/\\\//\\\_____________________________      
   ___\////\\\_________\/\\\_____________/\\\\\____\/\\\\///\\\/\\\/_\/\\\_____/\\\\\_____/\\/\\\\\\___     
    ______\////\\\______\/\\\\\\\\\\____/\\\///\\\__\/\\\__\///\\\/___\/\\\___/\\\///\\\__\/\\\////\\\__    
     _________\////\\\___\/\\\/////\\\__/\\\__\//\\\_\/\\\____\///_____\/\\\__/\\\__\//\\\_\/\\\__\//\\\_   
      __/\\\______\//\\\__\/\\\___\/\\\_\//\\\__/\\\__\/\\\_____________\/\\\_\//\\\__/\\\__\/\\\___\/\\\_  
       _\///\\\\\\\\\\\/___\/\\\___\/\\\__\///\\\\\/___\/\\\_____________\/\\\__\///\\\\\/___\/\\\___\/\\\_ 
        ___\///////////_____\///____\///_____\/////_____\///______________\///_____\/////_____\///____\///__`
	log.Println(asciiArt)
	log.Println("v3")
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
