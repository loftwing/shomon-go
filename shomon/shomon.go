package shomon

import (
	"encoding/json"
	"log"
	"os"

	"gopkg.in/ns3777k/go-shodan.v3/shodan"
)

//ShodanMon todo
type ShodanMon struct {
	ShodanClient *shodan.Client
	Config       *Config
}

//Config defines json structure for config file
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

//InitMon creates a new ShodanMon and returns it
func NewMonitor(configpath string) *ShodanMon {
	conf := loadConfig(configpath)
	newClient := shodan.NewClient(nil, conf.Shodan.APIKey)
	return &ShodanMon{
		ShodanClient: newClient,
		Config:       conf}
}

func (sm *ShodanMon) RegisterAlerts() error {
	c := sm.ShodanClient
	cAlerts := sm.Config.Shodan.Networks

	if rAlerts, err := c.GetAlerts(nil); err != nil {
		return err
	}

	log.Println("cAlerts")
	for _, ca := range cAlerts {
		log.Println(ca)
	}

	log.Println("rAlerts")
	for _, ra := range rAlerts {
		log.Printf("%+v\n", ra)
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

//Status prints current status of monitor to logger, or returns an error
func (sm *ShodanMon) Status() {
	c := sm.ShodanClient
	log.Println("Monitor Status")

	log.Println("======PROFILE======")
	if profile, err := c.GetAccountProfile(nil); err != nil {
		log.Println("Error pulling profile info from API.")
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
