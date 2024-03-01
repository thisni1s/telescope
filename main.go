package main

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	//"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/digitalocean/godo"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ApiToken        string   `yaml:"apiToken"`
	SSHKeys         []int    `yaml:"sshKeys"`
	Diameter        int      `yaml:"diameter"` // Diameter is the telescope size
	Regions         []string `yaml:"regions"`
	StartUpScript   string   `yaml:"startupScript"` // StartUpScript is a path to a bash script executed on node creation
	StorageLocation string   `yaml:"storageLocation"`
	StorageToken    string   `yaml:"storageToken"`
}

var cfg Config

func main() {
	fmt.Printf(banner)
	file, err := os.ReadFile("config.yml")
	if err != nil {
		log.Fatal("Could not read config file!")
	}
	err = yaml.Unmarshal(file, &cfg)

	client := godo.NewFromToken(cfg.ApiToken)
	ctx := context.TODO()

	fmt.Println("Requested telescope diameter: ", cfg.Diameter)

	list, err := DropletList(ctx, client)
	if err != nil {
		log.Println("Could not get Droplet list")
		log.Fatal(err)
	}
	currRegs := calcCurrRegions(&list)
	desiredRegs := calcRegions(cfg.Regions, cfg.Diameter)
	didTransactions := false
	if len(list) < cfg.Diameter {
		fmt.Println("To few nodes for requested diameter. Creating new nodes!")
		didTransactions = true
		var toCreate int = (cfg.Diameter - len(list))
		regionsStack := calcRegionStack(currRegs, desiredRegs, toCreate)

		for i := 0; i < toCreate; i++ {
			num := findLowestNum(&list)
			reg := popSlice(&regionsStack)
			drop, err := CreateDroplet(ctx, client, num, reg)
			if err != nil {
				log.Printf("Could not create droplet in region %s \n", reg)
				log.Fatal(err)
			}
			list = append(list, *drop)
			fmt.Printf("Created new node (%d) with name: %s \n", drop.ID, drop.Name)
		}
	}
	if len(list) > cfg.Diameter {
		fmt.Println("The current telescope is too big! Deleting oldest nodes!")
		sort.Slice(list, func(i, j int) bool {
			return getTimefromStr(list[i].Created).Before(getTimefromStr(list[j].Created))
		})
		didTransactions = true
		var toDelete int = (len(list) - cfg.Diameter)
		for i := 0; i < toDelete; i++ {
			fmt.Println("I will delete node: ", list[i].Name)
			_, err := client.Droplets.Delete(ctx, list[i].ID)
			if err != nil {
				log.Println("Error deleting node!")
				log.Fatal(err)
			}
		}
	}
	// Get current version of the list if we did something to it
	if didTransactions {
		list, err = DropletList(ctx, client)
		if err != nil {
			log.Println("Could not get Droplet list")
			log.Fatal(err)
		}
	}

	fmt.Printf("The telescope consists of %d nodes \n", len(list))
	for _, droplet := range list {
		ip, _ := droplet.PublicIPv4()
		if ip == "" {
			ip = "not yet Available"
		}
		fmt.Printf("  [%d] %s [%s] Created: %s IP: %s \n", droplet.ID, droplet.Name, droplet.Region.Name, getTimefromStr(droplet.Created).Format("02.01.2006 15:04:05"), ip)
		//a, _ := json.Marshal(droplet)
		//fmt.Println(string(a))

	}

}

func popSlice(arr *[]string) string {
	if len(*arr) == 0 {
		return ""
	}
	lastIndex := len(*arr) - 1
	popped := (*arr)[lastIndex]
	*arr = (*arr)[:lastIndex]
	return popped
}

func calcRegionStack(currRegs map[string]int, ideal map[string]int, toCreate int) []string {
	result := make(map[string]int)
	var resStack []string

	for key, value := range ideal {
		result[key] = value - currRegs[key]
	}

	// Add keys from currRegs that are not in ideal
	for key, value := range currRegs {
		if _, ok := result[key]; !ok {
			result[key] = -value
		}
	}

	for key, val := range result {
		for _ = range val {
			resStack = append(resStack, key)
		}
	}

	return resStack
}

func calcCurrRegions(list *[]godo.Droplet) map[string]int {
	result := make(map[string]int)
	for _, drop := range *list {
		if _, ok := result[drop.Region.Slug]; !ok {
			result[drop.Region.Slug]++
		} else {
			result[drop.Region.Slug] = 1
		}
	}
	return result
}

func calcRegions(regions []string, diameter int) map[string]int {
	result := make(map[string]int)

	if len(regions) == 0 {
		return result
	}

	// Calculate the quotient
	quotient := diameter / len(regions)

	// Calculate the remainder
	remainder := diameter % len(regions)

	// Assign the quotient to each string
	for _, str := range regions {
		result[str] = quotient
	}

	// Distribute the remainder
	for i := 0; i < remainder; i++ {
		result[regions[i]]++
	}

	return result
}

func getTimefromStr(str string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05Z", str)
	if err != nil {
		log.Fatal("Error parsing time!")
	}
	return t
}

func findLowestNum(list *[]godo.Droplet) int {
	var allNums []int
	for _, drop := range *list {
		num, err := strconv.Atoi(strings.Split(drop.Name, "-")[1])
		if err != nil {
			num = 9999
		}
		allNums = append(allNums, num)
	}
	sort.Ints(allNums)
	// Überprüfe jede Zahl in der sortierten Liste
	// und finde die kleinste freie Zahl
	lowestFree := 1
	for _, num := range allNums {
		if num == lowestFree {
			lowestFree++
		}
	}
	return lowestFree

}

func CreateDroplet(ctx context.Context, client *godo.Client, num int, region string) (*godo.Droplet, error) {
	var keys []godo.DropletCreateSSHKey
	for _, key := range cfg.SSHKeys {
		keys = append(keys, godo.DropletCreateSSHKey{ID: key})
	}

	script, err := os.ReadFile(cfg.StartUpScript)
	if err != nil {
		log.Println("Failed to read startup script!")
		log.Fatal(err)
	}

	scr := fmt.Sprintf(`%s
(crontab -l ; echo "0 * * * * sh /root/upload.sh %s %s") | crontab -
curl -u "%s:" -X MKCOL "%s/$ip"
systemctl start tcpdumpd
reboot`, string(script), cfg.StorageLocation, cfg.StorageToken, cfg.StorageToken, cfg.StorageLocation)

	createRequest := &godo.DropletCreateRequest{
		Name:   fmt.Sprintf("telescope-%d", num),
		Region: region,
		Size:   "s-1vcpu-512mb-10gb",
		Image: godo.DropletCreateImage{
			Slug: "ubuntu-23-10-x64",
		},
		Tags:     []string{"telescope"},
		SSHKeys:  keys,
		UserData: scr,
	}

	newDroplet, _, err := client.Droplets.Create(ctx, createRequest)
	return newDroplet, err
}

func DropletList(ctx context.Context, client *godo.Client) ([]godo.Droplet, error) {
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

var banner string = `
████████ ███████ ██      ███████ ███████  ██████  ██████  ██████  ███████ 
   ██    ██      ██      ██      ██      ██      ██    ██ ██   ██ ██      
   ██    █████   ██      █████   ███████ ██      ██    ██ ██████  █████   
   ██    ██      ██      ██           ██ ██      ██    ██ ██      ██      
   ██    ███████ ███████ ███████ ███████  ██████  ██████  ██      ███████ 
                                                                          
`
