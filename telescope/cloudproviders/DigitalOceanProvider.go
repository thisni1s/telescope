package cloudproviders

import (
	"context"
	"fmt"
	//"log"
	//"os"
	"strconv"
	"strings"

	"github.com/digitalocean/godo"
	"github.com/thisni1s/telescope/helpers"
)

type GodoConfig struct {
	StorageConfig
	GodoSpecifics
    CommonConfig
}

type GodoSpecifics struct {
	ApiToken string   `yaml:"apiToken"`
	Size     string   `yaml:"size"`
	Image    string   `yaml:"os"`
	SSHKeys  []int    `yaml:"sshKeys"`
	Regions  []string `yaml:"regions"`
	NumVMs   int      `yaml:"numVMs"`
}

type doClient struct {
	CloudProviderInfo
	client  *godo.Client
	context context.Context
	config  GodoConfig
}

func NewDigOceanClient(cfg GodoConfig) (*doClient, error) {
	client := godo.NewFromToken(cfg.ApiToken)
	ctx := context.TODO()
	return &doClient{
		CloudProviderInfo: CloudProviderInfo{
			Name:     "DigitalOcean",
			Diameter: cfg.NumVMs,
			Regions:  cfg.Regions,
		},
		client:  client,
		context: ctx,
		config:  cfg,
	}, nil
}

func (c *doClient) CreateVM(desc VMDescriptor) (VMDescriptor, error) {
	drop, err := createDroplet(c, desc)
	if err != nil {
		return VMDescriptor{}, err
	}

	descr := dropToDesc(drop)
	descr.Num = desc.Num
	descr.Provider = c
	return descr, nil
}

func (c *doClient) DestroyVM(vm VMDescriptor) error {
	num, err := strconv.Atoi(vm.ID)
	_, err = c.client.Droplets.Delete(c.context, num)
	return err
}

func (c *doClient) ListVMs() ([]VMDescriptor, error) {
	var vmlist []VMDescriptor
	droplist, err := dropletList(c.context, c.client)
	if err != nil {
		return vmlist, err
	}
	for _, drop := range droplist {
		desc := dropToDesc(&drop)
		num, err := strconv.Atoi(strings.Split(drop.Name, "-")[1])
		if err != nil {
			num = 9999
		}
		desc.Num = num
		desc.Provider = c
		vmlist = append(vmlist, desc)
	}
	return vmlist, nil
}

func (c *doClient) Info() CloudProviderInfo {
	return c.CloudProviderInfo
}

func createDroplet(c *doClient, desc VMDescriptor) (*godo.Droplet, error) {
	var keys []godo.DropletCreateSSHKey
	for _, key := range c.config.SSHKeys {
		keys = append(keys, godo.DropletCreateSSHKey{ID: key})
	}
    println("creating vm in region: ", desc.Region)
	scr := fmt.Sprintf(`#cloud-config
packages:
  - curl
runcmd:
  - curl -sSL https://raw.githubusercontent.com/thisni1s/telescope/refs/heads/main/telescope/assets/startup.sh | bash -s -- %s %s %s %s %s %s %s
`, c.config.StorageBucket, c.config.StorageLocation, c.config.StorageAccessKey, c.config.StorageSecretKey, c.config.CommonConfig.WebhookPw, "DigitalOcean", desc.Region)

	createRequest := &godo.DropletCreateRequest{
		Name:   fmt.Sprintf("telescope-%d", desc.Num),
		Region: desc.Region,
		Size:   c.config.Size,
		Image: godo.DropletCreateImage{
			Slug: c.config.Image,
		},
		Tags:     []string{"telescope"},
		SSHKeys:  keys,
		UserData: scr,
	}

	newDroplet, _, err := c.client.Droplets.Create(c.context, createRequest)
	return newDroplet, err
}

func dropletList(ctx context.Context, client *godo.Client) ([]godo.Droplet, error) {
	// create a list to hold our droplets
	list := []godo.Droplet{}

	// create options. initially, these will be blank
	opt := &godo.ListOptions{}
	for {
		droplets, resp, err := client.Droplets.List(ctx, opt)
		if err != nil {
			return nil, err
		}

		// append the current page's droplets to our list
		list = append(list, droplets...)

		// if we are at the last page, break out the for loop
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		// set the page we want for the next request
		opt.Page = page + 1
	}

	return list, nil
}

func safeIP(drop *godo.Droplet) string {
	ip, err := drop.PublicIPv4()
	if err != nil {
		return ""
	} else {
		return ip
	}
}

func dropToDesc(drop *godo.Droplet) VMDescriptor {
	return VMDescriptor{
		ID:      strconv.Itoa(drop.ID),
		Name:    drop.Name,
		OS:      drop.Image.Name,
		Size:    drop.Size.Slug,
		Region:  drop.Region.Name,
		IP:      safeIP(drop),
		Created: helpers.GetTimefromStr(drop.Created),
	}
}
