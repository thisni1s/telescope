package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/thisni1s/telescope/cloudproviders"
	"github.com/thisni1s/telescope/helpers"
	"gopkg.in/yaml.v3"
)

var ninteractive bool
var status bool
var cfg cloudproviders.TelescopeConfig
var client *http.Client
var currentDels sync.Map

var cproviders []cloudproviders.CloudProvider

func main() {
	// Parse command-line arguments
	configFilePath := flag.String("config", "", "Path to the YAML configuration file")
	ni := flag.Bool("non-interactive", false, "Run in non interactive mode")
	st := flag.Bool("status", false, "Display only the telescope status")
	flag.Parse()

	// Check if the file path was provided
	if *configFilePath == "" {
		*configFilePath = "config.yml"
		log.Println("No config file provided! Using config.yml!")
	}

	// Define a binary flag with a default value (false) and a description
	ninteractive = *ni
	status = *st

	if !ninteractive && !status {
		fmt.Printf(banner)
	} else if !status {
		log.Println("Started telescope manager")
	}

	telescope(*configFilePath)

	ticker := time.NewTicker(time.Duration(cfg.Common.RoundTime) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Running Telescope loop")
			telescope(*configFilePath)
		}
	}

}

func telescope(cfgPath string) {
	// Read Config file
	file, err := os.ReadFile(cfgPath)
	if err != nil {
		log.Fatal("Could not read config file!")
	}
	err = yaml.Unmarshal(file, &cfg)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Skip SSL certificate validation
		},
	}
	client = &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	var influxCl *helpers.InfluxClient
	metrics := false
	if cfg.InfluxMetrics.Url != "" {
		log.Println("Setting up metric")
		influxCl = helpers.NewInfluxClient(cfg.InfluxMetrics)
		defer influxCl.Close()
		metrics = true
	} else {
		log.Println("No metric server")
	}
	mt := helpers.Metric{
		Diameter:      0,
		Created:       0,
		Deleted:       0,
		ProviderCount: make(map[string]int),
	}

	initProviders()

	// Print requested telescope size
	if !status {
		log.Println("Requested telescope diameter: ")
		for _, prov := range cproviders {
			log.Printf("\t%s\t%d\n", prov.Info().Name, prov.Info().Diameter)
		}
	}

	// Get current telescope status
	list := make(map[string][]cloudproviders.VMDescriptor)
	for _, prov := range cproviders {
		plist, err := prov.ListVMs()
		if err != nil {
			log.Println("Error getting VM list from provider: ", prov.Info().Name)
		} else {
			list[prov.Info().Name] = plist
		}
	}

	// Print actual telescope size
	if !status {
		log.Println("Actual telescope diameter: ")
		for _, prov := range cproviders {
			log.Printf("\t%s\t%d\n", prov.Info().Name, len(list[prov.Info().Name]))
		}
	}

	didTransactions := false
	deleteOldNodes(&list, &didTransactions, &mt)
	adjustSize(&didTransactions, &list, &mt)

	// Get current version of the list if we did something to it
	var nlist []cloudproviders.VMDescriptor
	for name, prov := range list {
		nlist = append(nlist, prov...)
		mt.ProviderCount[name] = len(prov)
	}
	mt.Diameter = len(nlist)
	if metrics {
		influxCl.WriteData(mt)
	}
	if didTransactions {
		log.Printf("The telescope consists of %d nodes \n", len(nlist))
		println(mt.Stringify())
	} else if !status {
		influxCl.WriteData(mt)
		log.Println("No changes performed, quitting.")
	}

	if !ninteractive {
		prettyPrint(&list)
	}
}

func initProviders() {
	// Delete old providers and initialize them again
	cproviders = []cloudproviders.CloudProvider{}

	// Create provider clients
	docl, err := cloudproviders.NewDigOceanClient(cloudproviders.GodoConfig{
		StorageConfig: cfg.Storage,
		GodoSpecifics: cfg.DigOcean,
		CommonConfig:  cfg.Common,
	})

	if err != nil {
		log.Println("Error creating provider!")
		log.Fatal(err)
	}

	cproviders = append(cproviders, docl)

	mocl, err := cloudproviders.NewMockClient(cloudproviders.MockConfig{
		StorageConfig: cfg.Storage,
		MockSpecifics: cfg.Mock,
		CommonConfig:  cfg.Common,
	})

	if err != nil {
		log.Println("Error creating Mock Provider!")
		log.Fatal(err)
	}

	cproviders = append(cproviders, mocl)

}

func adjustSize(didTransactions *bool, list *map[string][]cloudproviders.VMDescriptor, mt *helpers.Metric) {
	oldDeletions := mt.Deleted
	// What do we have to do?
	currRegs := calcCurrRegions(list)

	for _, prov := range cproviders {
		provName := prov.Info().Name
		desiredRegs := calcRegions(prov.Info().Regions, prov.Info().Diameter)

		if len((*list)[provName]) < prov.Info().Diameter {

			log.Printf("To few nodes for requested diameter for provider: %s. Creating new nodes! \n", provName)
			*didTransactions = true
			var toCreate int = (prov.Info().Diameter - len((*list)[provName]))
			regionsStack := calcRegionStack(currRegs[provName], desiredRegs)

			created := 0
			for i := 0; i < toCreate; i++ {
				if created < cfg.Common.NodesPerRound+oldDeletions {
					num := findLowestNum(list)
					reg := popSlice(&regionsStack)
					vm, err := prov.CreateVM(cloudproviders.VMDescriptor{
						Num:    num,
						Region: reg,
					})
					if err != nil {
						log.Printf("Could not create VM for Provider %s in region %s \n", provName, reg)
						log.Fatal(err)
					}
					(*list)[provName] = append((*list)[provName], vm)
					created++
					mt.Created += 1
					log.Printf("Created new node (%d) on %s with name: %s \n", vm.Num, provName, vm.Name)

				} else {
					log.Println("Already created maximum amount of nodes for this round. Waiting...")
					break
				}
			}
		}

		if len((*list)[provName]) > prov.Info().Diameter {
			log.Printf("The current %s telescope is too big! Deleting oldest nodes!\n", provName)
			*didTransactions = true
			plist := (*list)[provName]
			sort.Slice(plist, func(i, j int) bool {
				return plist[i].Created.Before(plist[j].Created)
			})

			var toDelete int = (len((*list)[provName]) - prov.Info().Diameter)
			for i := 0; i < toDelete; i++ {
				log.Println("I will delete node: ", plist[i].Name)
				mt.Deleted += 1
				go deleteOrder(plist[i], 10)
			}
		}
	}
}

func deleteOldNodes(list *map[string][]cloudproviders.VMDescriptor, didTransactions *bool, mt *helpers.Metric) {
	if cfg.Common.Lifetime == -1 {
		log.Println("No lifetime configured, keeping all nodes and adjusting diameter only!")
		return
	}

	deletedNodeIDs := []string{}

	for _, provider := range *list {
		for _, vm := range provider {
			if time.Now().Sub(vm.Created).Minutes() > float64(cfg.Common.Lifetime) {
				log.Printf("Node %s is too old, tearing down and ordering deletion!\n", vm.Name)
				id := vm.ID
				mt.Deleted += 1
				go deleteOrder(vm, 10)
				deletedNodeIDs = append(deletedNodeIDs, id)
			}
		}
	}

	if len(deletedNodeIDs) > 0 {
		*didTransactions = true
	}

	// Modifying the list to ensure that new nodes are created immediately and not wait until the next run
	for indx, provider := range *list {
		newPList := []cloudproviders.VMDescriptor{}
		for _, vm := range provider {
			deleted := false
			for _, del := range deletedNodeIDs {
				if del == vm.ID {
					deleted = true
				}
			}
			if !deleted {
				newPList = append(newPList, vm)
			}
		}
		(*list)[indx] = newPList
	}

}

func deleteOrder(vm cloudproviders.VMDescriptor, waitTimeS int) {
	_, loaded := currentDels.LoadOrStore(vm.ID, struct{}{})
	if loaded {
		// Another goroutine is working on a deletion of this node
		log.Printf("Ordered to delete %s but another goroutine is working on it. Returning\n", vm.Name)
		return
	}
	defer currentDels.Delete(vm.ID)

	_, err := teardownVM(&vm)
	if err != nil {
		log.Printf("Teardown not working for node: %s (%s) retrying later!\n", vm.Name, vm.Provider.Info().Name)
		if vm.Provider.Info().Name == "MockClient" {
			log.Println("Deleting mock node: ", vm.Name)
			vm.Provider.DestroyVM(vm)
			return
		}
	}
	// wait some time
	time.Sleep(time.Duration(waitTimeS) * time.Second)

	for i := range 6 {
		state, err := getVMStatus(&vm)
		if err != nil {
			log.Println("Cannot get Status of node: ", vm.Name)
		} else if state.Teardown == "finished" {
			log.Printf("Teardown for node %s finished after %d tries, deleting now.\n", vm.Name, i+1)
			vm.Provider.DestroyVM(vm)
			return
		} else if state.Teardown == "available" {
			log.Println("Teardown not started yet? Tearing down node: ", vm.Name)
			_, erro := teardownVM(&vm)
			if erro != nil {
				log.Println("Teardown not working for node: ", vm.Name)
			}
		} else if state.Teardown == "started" {
			log.Printf("Teardown still running on node: %s Retry: %d", vm.Name, i+1)
		}
		time.Sleep(time.Duration(waitTimeS) * time.Second)
	}

	// node unresponsive. Deleting by provider
	log.Printf("Node %s is unresponsive, deleting by calling provider!\n", vm.Name)
	vm.Provider.DestroyVM(vm)
}

func getVMStatus(vm *cloudproviders.VMDescriptor) (cloudproviders.VMStatusResponse, error) {
	var response cloudproviders.VMStatusResponse
	url := fmt.Sprintf("https://%s:51337/hooks/status", vm.IP)

	resp, err := client.Get(url)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}
	err = json.Unmarshal(body, &response)
	return response, err
}

func teardownVM(vm *cloudproviders.VMDescriptor) (*http.Response, error) {
	url := fmt.Sprintf("https://%s:51337/hooks/teardown?token=%s", vm.IP, cfg.Common.WebhookPw)
	resp, err := client.Get(url)
	return resp, err
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

func calcRegionStack(currRegs map[string]int, ideal map[string]int) []string {
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

func calcCurrRegions(list *map[string][]cloudproviders.VMDescriptor) map[string]map[string]int {
	res := make(map[string]map[string]int)
	for provName, vms := range *list {
		for _, vm := range vms {
			if _, ok := res[provName][vm.Region]; !ok {
				res[provName] = make(map[string]int)
			} else {
				res[provName][vm.Region]++
			}
		}
	}
	return res
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

func findLowestNum(list *map[string][]cloudproviders.VMDescriptor) int {
	var allNums []int
	for _, vms := range *list {
		for _, vm := range vms {
			allNums = append(allNums, vm.Num)
		}
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

func prettyPrint(list *map[string][]cloudproviders.VMDescriptor) {
	if !status {
		fmt.Println()
		fmt.Println("The current telescope:")
	}

	cList := make([]cloudproviders.VMDescriptor, 0)
	for _, vms := range *list {
		cList = append(cList, vms...)
	}

	if len(cList) < 1 {
		return
	}

	// Calculate max widths for each column
	var maxId, maxName, maxReg, maxIP, maxProv int
	for _, vm := range cList {
		maxId = int(math.Max(float64(maxId), float64(len(vm.ID))))
		maxName = int(math.Max(float64(maxName), float64(len(vm.Name))))
		maxReg = int(math.Max(float64(maxReg), float64(len(vm.Region))))
		maxIP = int(math.Max(float64(maxIP), float64(len(vm.IP))))
		maxProv = int(math.Max(float64(maxProv), float64(len(vm.Provider.Info().Name))))
	}

	// Set minimum column width
	maxId = int(math.Max(float64(maxId), 5))
	maxName = int(math.Max(float64(maxName), 5))
	maxReg = int(math.Max(float64(maxReg), 10))
	maxIP = int(math.Max(float64(maxIP), 5))
	maxProv = int(math.Max(float64(maxProv), 10))

	// Print table header with dynamic column widths
	s := fmt.Sprintf("%%-%ds\t%%-%ds\t%%-%ds\t%%-20s\t%%-%ds\t\t%%-%ds\n", maxId, maxName, maxReg, maxIP, maxProv)
	fmt.Printf(s, "[ID]", "Name", "[Region]", "Created", "IP", "(Provider)")

	for _, vm := range cList {
		fmt.Printf(s, vm.ID, vm.Name, vm.Region, vm.Created.Format("02.01.2006 15:04:05"), vm.IP, vm.Provider.Info().Name)
	}

}

var banner string = `
████████ ███████ ██      ███████ ███████  ██████  ██████  ██████  ███████ 
   ██    ██      ██      ██      ██      ██      ██    ██ ██   ██ ██      
   ██    █████   ██      █████   ███████ ██      ██    ██ ██████  █████   
   ██    ██      ██      ██           ██ ██      ██    ██ ██      ██      
   ██    ███████ ███████ ███████ ███████  ██████  ██████  ██      ███████ 
                                                                          
`
