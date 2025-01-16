package helpers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type InfluxConfig struct {
	Url    string `yaml:"url"`
	Token  string `yaml:"token"`
	Org    string `yaml:"org"`
	Bucket string `yaml:"bucket"`
}

type InfluxClient struct {
	writeAPI api.WriteAPIBlocking
	client   influxdb2.Client
	cfg      InfluxConfig
}

type Metric struct {
	Diameter      int            `json:"diameter"`
	Created       int            `json:"created"`
	Deleted       int            `json:"deleted"`
	ProviderCount map[string]int `json:"providers"`
}

func NewInfluxClient(cfg InfluxConfig) *InfluxClient {
	cl := influxdb2.NewClientWithOptions(cfg.Url, cfg.Token, influxdb2.DefaultOptions().SetBatchSize(20))
	wa := cl.WriteAPIBlocking(cfg.Org, cfg.Bucket)
	return &InfluxClient{
		writeAPI: wa,
		client:   cl,
		cfg:      cfg,
	}
}

func (ic *InfluxClient) WriteData(mt Metric) {
	var points []*write.Point
	p := influxdb2.NewPoint(
		"telescope",
		map[string]string{
			"telescope": "telescope",
		},
		map[string]interface{}{
			"diameter": mt.Diameter,
			"created":  mt.Created,
			"deleted":  mt.Deleted,
		},
		time.Now())
	points = append(points, p)
	for key, val := range mt.ProviderCount {
		pt := influxdb2.NewPoint(
			key,
			map[string]string{
				"telescope": "telescope",
			},
			map[string]interface{}{
				"nodes": val,
			},
			time.Now())
		points = append(points, pt)
	}
	for _, point := range points {
		if err := ic.writeAPI.WritePoint(context.Background(), point); err != nil {
			log.Println("Failed to write to metric server ", err)
		}
	}
}

func (ic *InfluxClient) Close() {
	ic.client.Close()
}

func (mt *Metric) Stringify() string {
	// Use strings.Builder to concatenate the map items into one string
	var sb strings.Builder
	for key, value := range mt.ProviderCount {
		sb.WriteString(fmt.Sprintf("%s: %d, ", key, value))
	}

	// Convert strings.Builder to string and remove the trailing comma and space
	result := sb.String()
	if len(result) > 0 {
		result = result[:len(result)-2] // Remove the last ", "
	}

	return fmt.Sprintf("Metric: Diamter: %d, Created: %d, Deleted: %d, Providers: %s", mt.Diameter, mt.Created, mt.Deleted, result)
}
