package cloudproviders

import "time"

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
	Diameter int           `yaml:"diameter"`
	Common   CommonConfig  `yaml:"common"`
	Storage  StorageConfig `yaml:"storage"`
	DigOcean GodoSpecifics `yaml:"digitalOcean"`
	Mock     MockSpecifics `yaml:"mockProvider"`
}

type CommonConfig struct {
	WebhookPw string `yaml:"webhookPw"`
	Lifetime  int    `yaml:"lifetime"`
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
