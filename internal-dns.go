package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	api "gopkg.in/ns1/ns1-go.v2/rest"
	// "gopkg.in/ns1/ns1-go.v2/rest/model/data"
	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
	// "gopkg.in/ns1/ns1-go.v2/rest/model/filter"
	sockaddr "github.com/hashicorp/go-sockaddr"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"golang.org/x/exp/slices"
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

func getAllRecordAnswers(r *dns.Record) []string {
	answers := []string{}

	if r != nil {
		for _, a := range r.Answers {
			answers = append(answers, a.Rdata...)
		}
	}

	return answers
}

func createOrUpdateRecord(ctx context.Context, client *api.Client, zone, domain string) error {
	privateIP, _ := sockaddr.GetPrivateIP()
	if privateIP != "" {
		recType := "A"
		logger := log.WithFields(log.Fields{
			"zone":        zone,
			"domain":      domain,
			"record_type": recType,
		})

		fullDomain := domain
		if !strings.Contains(fullDomain, zone) {
			fullDomain = fmt.Sprintf("%s", domain+"."+zone)
		}

		rec, _, err := client.Records.Get(zone, fullDomain, recType)
		if err != nil {
			if err == api.ErrRecordMissing {
				// record not found, create new
				newRec := dns.NewRecord(zone, domain, recType)
				ans := dns.NewAv4Answer(privateIP)

				newRec.AddAnswer(ans)
				if _, err = client.Records.Create(newRec); err != nil {
					logger.WithFields(log.Fields{
						"error": err,
					}).Error("Failed to create record")

					return err
				}
			} else {
				// something else went wrong
				logger.WithFields(log.Fields{
					"error": err,
				}).Error("Failed to get record")

				return err
			}
		}

		if !slices.Contains(getAllRecordAnswers(rec), privateIP) {
			// record exists and needs updating
			logger.Debug("Updating record")

			newRec := dns.NewRecord(zone, domain, recType)
			ans := dns.NewAv4Answer(privateIP)

			newRec.AddAnswer(ans)
			_, err = client.Records.Update(newRec)
			if err != nil {
				logger.WithFields(log.Fields{
					"error": err,
				}).Error("Failed to update record")

				return err
			}
		}
	}

	return nil
}

func init() {
	// init logging
	log.SetOutput(io.Discard) // Send all logs to nowhere by default

	log.AddHook(&writer.Hook{ // Send logs with level higher than warning to stderr
		Writer: os.Stderr,
		LogLevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
			log.WarnLevel,
		},
	})
	log.AddHook(&writer.Hook{ // Send info and debug logs to stdout
		Writer: os.Stdout,
		LogLevels: []log.Level{
			log.InfoLevel,
			log.DebugLevel,
		},
	})
}

func main() {
	// flags
	logLevelFlag := flag.String("log-level", "info", "Logging level may be one of: trace, debug, info, warning, error, fatal and panic")
	logFmtFlag := flag.String("log-format", "logfm", "Output format of logs: [`logfmt`, `json`] (default: `logfmt`)")
	zoneFlag := flag.String("zone", "", "The zone to add records to")
	domainFlag := flag.String("domain", "", "The domain to create a record for (defaults to system hostname if empty)")

	flag.Parse()

	// validate flags
	if *zoneFlag == "" {
		log.WithFields(log.Fields{"flag": "zone"}).Fatal("Missing command line flag")
	}

	domain := *domainFlag
	if domain == "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Fatal("Failed to retrieve hostname")
		}

		domain = hostname
	}

	// set log format based on flag
	// default is logfmt, so only make changes if json requested
	if strings.ToLower(*logFmtFlag) == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	}

	logPrettyfierFunc := func(f *runtime.Frame) (string, string) {
		fileName := filepath.Base(f.File)
		funcName := filepath.Base(f.Function)
		return fmt.Sprintf("%s()", funcName), fmt.Sprintf("%s:%d", fileName, f.Line)
	}

	// set log level based on flag
	level, err := log.ParseLevel(*logLevelFlag)
	if err != nil {
		// if log level couldn't be parsed from config, default to info level
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(level)

		if level >= log.DebugLevel {
			// enable func/file logging
			log.SetReportCaller(true)

			if strings.ToLower(*logFmtFlag) == "json" {
				log.SetFormatter(&log.JSONFormatter{CallerPrettyfier: logPrettyfierFunc})
			} else {
				log.SetFormatter(&log.TextFormatter{CallerPrettyfier: logPrettyfierFunc})
			}

		}

		log.Infof("Log level set to: %s", level)
	}

	logger := log.WithFields(log.Fields{
		"version":    Version,
		"build_date": BuildDate,
		"commit":     Commit,
		"go_version": runtime.Version(),
	})
	logger.Info("Internal DNS server started")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := newNS1APIClient()

	if err := createOrUpdateRecord(ctx, client, *zoneFlag, domain); err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"domain": domain,
			"zone":   *zoneFlag,
		}).Error("Failed to set record for internal DNS")
	} else {
		log.WithFields(log.Fields{
			"domain": domain,
			"zone":   *zoneFlag,
		}).Info("Internal DNS record set")
	}

	logger.Info("Internal DNS exited")
}
