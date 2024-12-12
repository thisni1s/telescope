package cloudproviders

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

type MockSpecifics struct {
	ApiToken        string   `yaml:"apiToken"`
	Regions         []string `yaml:"regions"`
	ImportantFile   string   `yaml:"impFile"`
	ImportantString string   `yaml:"impStr"`
	Diameter        int      `yaml:"diameter"`
}

type MockConfig struct {
	MockSpecifics
	StorageConfig
    CommonConfig
}

type mockClient struct {
	CloudProviderInfo
	client string
	config MockConfig
	vms    []VMDescriptor
}

func NewMockClient(cfg MockConfig) (*mockClient, error) {
	initFile(cfg.ImportantFile)
	client := "Mock Client"
	return &mockClient{
		CloudProviderInfo: CloudProviderInfo{
			Name:     "MockClient",
			Diameter: cfg.Diameter,
			Regions:  cfg.Regions,
		},
		client: client,
		config: cfg,
	}, nil
}

func (c *mockClient) CreateVM(desc VMDescriptor) (VMDescriptor, error) {
	desc.Created = time.Now()
	desc.Provider = c
	desc.IP = generateRandomIP()
	desc.Name = fmt.Sprintf("telescope-%d", desc.Num)
	desc.ID = strconv.Itoa(rand.Intn(9000) + 1000)
	desc.OS = "templeOS"
	desc.Size = "verybig"

	c.vms = append(c.vms, desc)
	saveVMs(c.vms, c.config.ImportantFile)

	return desc, nil

}

func (c *mockClient) ListVMs() ([]VMDescriptor, error) {
	var list []VMDescriptor
	list, err := readVMs(c.config.ImportantFile)
	for i := range list {
		list[i].Provider = c
	}
	c.vms = list
	return list, err
}

func (c *mockClient) DestroyVM(desc VMDescriptor) error {
	var list []VMDescriptor
	list, err := readVMs(c.config.ImportantFile)
	if err != nil {
		log.Println("Error reading VM list, cannot delete VM")
	}
	index := indexOf(&list, desc)
	deleteFromSlice(&list, index)
	c.vms = list
	err = saveVMs(list, c.config.ImportantFile)
	return err
}

func (c *mockClient) Info() CloudProviderInfo {
	return c.CloudProviderInfo
}

// Function to generate a random IP address
func generateRandomIP() string {

	// Generate 4 random numbers for each octet of the IP address
	octet1 := rand.Intn(256)
	octet2 := rand.Intn(256)
	octet3 := rand.Intn(256)
	octet4 := rand.Intn(256)

	// Return the IP address as a string
	return fmt.Sprintf("%d.%d.%d.%d", octet1, octet2, octet3, octet4)
}

func saveVMs(list []VMDescriptor, filepath string) error {
	for i := range list {
		list[i].Provider = nil
	}

	// Create a buffer to hold the binary data
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	// Serialize the struct into the buffer
	err := encoder.Encode(list)
	if err != nil {
		log.Println("Error encoding mock provider file")
		log.Println(err)
		return err
	}

	// Write the buffer data to a file
	err = os.WriteFile(filepath, buf.Bytes(), 0644)
	if err != nil {
		log.Println("Error saving mock provider file")
		return err
	}
	return nil
}

func readVMs(filepath string) ([]VMDescriptor, error) {
	var list []VMDescriptor
	// Read the data back from the file
	fileData, err := os.ReadFile(filepath)
	if err != nil {
		log.Println(err)
		return list, err
	}

	// Create a buffer from the file data
	buf := bytes.NewBuffer(fileData)
	decoder := gob.NewDecoder(buf)

	// Deserialize the data into the struct
	err = decoder.Decode(&list)
	if err != nil {
		log.Println(err)
	}
	return list, err
}

func initFile(filepath string) error {
	gob.Register(mockClient{})
	gob.Register(VMDescriptor{})
	// Check if the file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		var buf bytes.Buffer
		encoder := gob.NewEncoder(&buf)

		// Serialize the struct into the buffer
		err := encoder.Encode(make([]VMDescriptor, 0))
		if err != nil {
			log.Println("Cannot encode the Mock file!")
			return err
		}

		// Write the buffer data to the file
		err = os.WriteFile(filepath, buf.Bytes(), 0644)
		if err != nil {
			log.Println("Cannot write Mock file!")
			return err
		}
	}
	return nil
}

// Function to find the index of a struct in a slice
func indexOf(slice *[]VMDescriptor, target VMDescriptor) int {
	for i, p := range *slice {
		if p.ID == target.ID {
			return i
		}
	}
	return -1 // Return -1 if the struct is not found
}

// Function to delete an element from the slice by index
func deleteFromSlice(slice *[]VMDescriptor, index int) {
	(*slice)[index] = (*slice)[len(*slice)-1]
	*slice = (*slice)[:len(*slice)-1]
}
