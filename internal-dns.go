package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	api "gopkg.in/ns1/ns1-go.v2/rest"
	// "gopkg.in/ns1/ns1-go.v2/rest/model/data"
	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
	// "gopkg.in/ns1/ns1-go.v2/rest/model/filter"
	log "github.com/sirupsen/logrus"
	// sockaddr "github.com/hashicorp/go-sockaddr"
)

var (
	Version   string // will be populated by linker during `go build`
	BuildDate string // will be populated by linker during `go build`
	Commit    string // will be populated by linker during `go build`
)

func newNS1APIClient() *api.Client {
	token := os.Getenv("NS1_TOKEN")
	if token == "" {
		log.Fatalln("NS1_TOKEN environment variable is not set")
	}

	httpClient := &http.Client{Timeout: time.Second * 10}
	client := api.NewClient(httpClient, api.SetAPIKey(token))

	return client
}

func main() {
	// fmt.Println("Private IP:")
	// ip, _ := sockaddr.GetPrivateIP()
	// fmt.Println(ip)
	//
	// hostname, err := os.Hostname()
	// if err != nil {
	// log it
	// }

	// real
	logger := log.WithFields(log.Fields{
		"version":    Version,
		"build_date": BuildDate,
		"commit":     Commit,
		"go_version": runtime.Version(),
	})
	logger.Info("Internal DNS server started")

	client := newNS1APIClient()
	zones, _, err := client.Zones.List()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to list zones")
	}

	for _, z := range zones {
		fmt.Println("Zone: ", z)
		fmt.Printf("Full zone object: %#v\n", z)
	}

	testA := dns.NewRecord("ns1.work.tjhop.io", "test", "A")
	testAAnswer := dns.NewAv4Answer("1.2.3.4")
	testAAnswer.Meta.Priority = 1

	testA.AddAnswer(testAAnswer)
	_, err = client.Records.Update(testA)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to create record")
	}

	// TODO: need to make a `UpdateOrCreateRecord` function -- 
	// create only works for new records
	// update only works for existing records
	// need to check if exists to determine if create/update
	newA := dns.NewRecord("ns1.work.tjhop.io", "new", "A")
	newAAnswer := dns.NewAv4Answer("1.2.3.4")
	newAAnswer.Meta.Priority = 1

	newA.AddAnswer(newAAnswer)
	_, err = client.Records.Update(newA)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to create record")
	}
	
	logger.Info("Internal DNS exited")
}
