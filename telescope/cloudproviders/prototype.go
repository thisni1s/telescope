package cloudproviders

import "time"

type CloudProvider interface {
	CreateVM(VMDescriptor) (VMDescriptor, error)
	ListVMs() ([]VMDescriptor, error)
	DestroyVM(VMDescriptor) error
	Info() CloudProviderInfo
}

type CloudProviderInfo struct {
	Name string
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
