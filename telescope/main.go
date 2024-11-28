package main

import (
	"fmt"
	"github.com/thisni1s/telescope/cloudproviders"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"sort"
	"time"
)

type Config struct {
	ApiToken         string                       `yaml:"apiToken"`
	SSHKeys          []int                        `yaml:"sshKeys"`
	Diameter         int                          `yaml:"diameter"` // Diameter is the telescope size
	Regions          []string                     `yaml:"regions"`
	StartUpScript    string                       `yaml:"startupScript"` // StartUpScript is a path to a bash script executed on node creation
	StorageLocation  string                       `yaml:"storageLocation"`
	StorageAccessKey string                       `yaml:"accessKey"`
	StorageSecretKey string                       `yaml:"secretKey"`
	StorageBucket    string                       `yaml:"storageBucket"`
	DigitalOcean     cloudproviders.GodoSpecifics `yaml:"digitalOcean"`
}

var cfg Config

var cproviders []cloudproviders.CloudProvider

func main() {
	fmt.Printf(banner)
	file, err := os.ReadFile("config.yml")
	if err != nil {
		log.Fatal("Could not read config file!")
	}
	err = yaml.Unmarshal(file, &cfg)

	docl, err := cloudproviders.NewDigOceanClient(cloudproviders.GodoConfig{
		ApiToken:         cfg.ApiToken,
		SSHKeys:          cfg.SSHKeys,
		StartUpScript:    cfg.StartUpScript,
		StorageLocation:  cfg.StorageLocation,
		StorageAccessKey: cfg.StorageAccessKey,
		StorageSecretKey: cfg.StorageSecretKey,
		StorageBucket:    cfg.StorageBucket,
		Image:            cfg.DigitalOcean.Image,
		Size:             cfg.DigitalOcean.Size,
	})
	if err != nil {
		log.Println("Error creating provider!")
		log.Fatal(err)
	}

	cproviders = append(cproviders, docl)

	fmt.Println("Requested telescope diameter: ", cfg.Diameter)

	var list []cloudproviders.VMDescriptor
	for _, prov := range cproviders {
		plist, err := prov.ListVMs()
		if err != nil {
			log.Println("Error getting VM list from provider: ", prov.Info().Name)
		} else {
			list = append(list, plist...)
		}
	}

	currRegs := calcCurrRegions(&list)
	desiredRegs := calcRegions(cfg.Regions, cfg.Diameter)
	didTransactions := false
	if len(list) < cfg.Diameter {
		fmt.Println("To few nodes for requested diameter. Creating new nodes!")
		didTransactions = true
		var toCreate int = (cfg.Diameter - len(list))
		regionsStack := calcRegionStack(currRegs, desiredRegs, toCreate)

		// !TODO -> Spread new vms accross providers by key?
		distribution := calcVmPerProvider(&cproviders, toCreate)

		for _, prov := range cproviders {
			provCreate := distribution[prov.Info().Name]
			for i := 0; i < provCreate; i++ {
				num := findLowestNum(&list)
				reg := popSlice(&regionsStack)
				vm, err := prov.CreateVM(cloudproviders.VMDescriptor{
					Num:    num,
					Region: reg,
				})
				if err != nil {
					log.Printf("Could not create VM for Provider %s in region %s \n", prov.Info().Name, reg)
					log.Fatal(err)
				}
				list = append(list, vm)
				fmt.Printf("Created new node (%d) on %s with name: %s \n", vm.Num, prov.Info().Name, vm.Name)
			}
		}

	}
	if len(list) > cfg.Diameter {
		fmt.Println("The current telescope is too big! Deleting oldest nodes!")
		sort.Slice(list, func(i, j int) bool {
			return list[i].Created.Before(list[j].Created)
		})
		didTransactions = true
		var toDelete int = (len(list) - cfg.Diameter)
		for i := 0; i < toDelete; i++ {
			fmt.Println("I will delete node: ", list[i].Name)
			err := list[i].Provider.DestroyVM(list[i])
			if err != nil {
				log.Println("Error deleting node!")
				log.Fatal(err)
			}
		}
	}
	// Get current version of the list if we did something to it
	if didTransactions {
		var list []cloudproviders.VMDescriptor
		for _, prov := range cproviders {
			plist, err := prov.ListVMs()
			if err != nil {
				log.Println("Could not get Droplet list for provider: ", prov.Info().Name)
				log.Fatal(err)
			}
			list = append(list, plist...)
		}
	}

	fmt.Printf("The telescope consists of %d nodes \n", len(list))
	for _, vm := range list {
		ip := "not yet available"
		if vm.IP != "" {
			ip = vm.IP
		}
		fmt.Printf("  [%s] %s [%s] Created: %s IP: %s \n", vm.ID, vm.Name, vm.Region, vm.Created.Format("02.01.2006 15:04:05"), ip)
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

func calcVmPerProvider(provlist *[]cloudproviders.CloudProvider, numCreate int) map[string]int {

	distribution := make(map[string]int)
	if numCreate >= len(*provlist) {
		baseNum := numCreate / len(*provlist)
		for _, prov := range *provlist {
			distribution[prov.Info().Name] = baseNum
		}
	}

	remainder := numCreate % len(*provlist)

	for _, prov := range *provlist {
		if remainder > 0 {
			distribution[prov.Info().Name]++
			remainder--
		} else {
			break
		}
	}
	return distribution
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
		for range val {
			resStack = append(resStack, key)
		}
	}

	return resStack
}

func calcCurrRegions(list *[]cloudproviders.VMDescriptor) map[string]int {
	result := make(map[string]int)
	for _, vm := range *list {
		if _, ok := result[vm.Region]; !ok {
			result[vm.Region]++
		} else {
			result[vm.Region] = 1
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

func findLowestNum(list *[]cloudproviders.VMDescriptor) int {
	var allNums []int
	for _, vm := range *list {
		allNums = append(allNums, vm.Num)
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

var banner string = `
████████ ███████ ██      ███████ ███████  ██████  ██████  ██████  ███████ 
   ██    ██      ██      ██      ██      ██      ██    ██ ██   ██ ██      
   ██    █████   ██      █████   ███████ ██      ██    ██ ██████  █████   
   ██    ██      ██      ██           ██ ██      ██    ██ ██      ██      
   ██    ███████ ███████ ███████ ███████  ██████  ██████  ██      ███████ 
                                                                          
`
