package cloudproviders

import (
	"github.com/thisni1s/telescope/helpers"
	"time"
)

type CloudProvider interface {
	CreateVM(VMDescriptor) (VMDescriptor, error)
	ListVMs() ([]VMDescriptor, error)
	DestroyVM(VMDescriptor) error
	Info() CloudProviderInfo
}

type CloudProviderInfo struct {
	Name     string
	Regions  []string
	Diameter int
}

type StorageConfig struct {
	StorageLocation  string `yaml:"storageLocation"`
	StorageAccessKey string `yaml:"accessKey"`
	StorageSecretKey string `yaml:"secretKey"`
	StorageBucket    string `yaml:"storageBucket"`
}

type TelescopeConfig struct {
	Diameter      int                  `yaml:"diameter"`
	Common        CommonConfig         `yaml:"common"`
	Storage       StorageConfig        `yaml:"storage"`
	DigOcean      GodoSpecifics        `yaml:"digitalOcean"`
	Mock          MockSpecifics        `yaml:"mockProvider"`
	InfluxMetrics helpers.InfluxConfig `yaml:"influxdb"`
}

type CommonConfig struct {
	WebhookPw     string `yaml:"webhookPw"`
	Lifetime      int    `yaml:"lifetime"`
	RoundTime     int    `yaml:"roundTime"`
	NodesPerRound int    `yaml:"nodesPerRound"`
}

type VMDescriptor struct {
	ID        string
	Num       int
	Name      string
	OS        string
	Size      string
	Region    string
	IP        string
	Created   time.Time
	Destroyed time.Time
	Provider  CloudProvider
}

type VMStatusResponse struct {
	Hostname string `json:"hostname"`
	Uptime   struct {
		Time               string  `json:"time"`
		Uptime             string  `json:"uptime"`
		Load1M             float64 `json:"load_1m"`
		Load5M             float64 `json:"load_5m"`
		Load15M            float64 `json:"load_15m"`
		TimeHour           int     `json:"time_hour"`
		TimeMinute         int     `json:"time_minute"`
		TimeSecond         int     `json:"time_second"`
		UptimeDays         int     `json:"uptime_days"`
		UptimeHours        int     `json:"uptime_hours"`
		UptimeMinutes      int     `json:"uptime_minutes"`
		UptimeTotalSeconds int     `json:"uptime_total_seconds"`
	} `json:"uptime"`
	Ipv4            string    `json:"ipv4"`
	Teardown        string    `json:"teardown"`
	Created         time.Time `json:"created"`
	Provider        string    `json:"provider"`
	Region          string    `json:"region"`
	Os              string    `json:"os"`
	SSHService      string    `json:"ssh.service"`
	SSHSocket       string    `json:"ssh.socket"`
	WebhookService  string    `json:"webhook.service"`
	WebhookSocket   string    `json:"webhook.socket"`
	TcpdumpdService string    `json:"tcpdumpd.service"`
}
