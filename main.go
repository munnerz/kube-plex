package main

import (
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/go-yaml/yaml"

	"github.com/munnerz/plex-elastic-transcoder/common"

	_ "github.com/munnerz/plex-elastic-transcoder/executors/kubernetes"
)

const configFile = "/etc/plex-elastic-transcoder/config.yaml"

var executor common.Executor

func signals() {
	// Signal handling
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		log.Infof("shutting down...")

		if executor != nil {
			log.Infof("terminating running job...")
			err := executor.Stop()
			if err != nil {
				log.Fatalf("error terminating job: %s", err.Error())
				os.Exit(1)
			}

			log.Infof("successfully terminated running job")
		}

		os.Exit(0)
	}()
}

func loadConfig() (*common.Config, error) {
	data, err := ioutil.ReadFile(configFile)

	if err != nil {
		return nil, err
	}

	config := new(common.Config)

	err = yaml.Unmarshal(data, config)

	return config, err
}

func main() {
	log.SetLevel(log.DebugLevel)
	config, err := loadConfig()

	if err != nil {
		log.Fatalf("error loading config: %s", err.Error())
	}

	log.Debugf("loaded config: %s", config)

	// Setup signals
	signals()

	// Setup logs
	if len(config.LogFile) > 0 {
		fo, err := os.Create(config.LogFile)

		if err != nil {
			log.Fatalf("error opening log file: %s", err)
		}

		log.SetOutput(fo)
		defer func() {
			if err := fo.Close(); err != nil {
				log.Fatalf("error closing file: %s", err)
			}
		}()
	}

	// Get the arguments passed to Plex New Transcoder
	args := os.Args[1:]
	log.Debugf("executing job with args: %s", args)
	wd, _ := os.Getwd()
	for i, arg := range args {
		switch arg {
		case "-progressurl":
			log.Debugf("replacing progressURL with: %s", config.Plex.URL)
			args[i+1] = strings.Replace(args[i+1], "127.0.0.1:32400", config.Plex.URL, 1)
			break
		case "-loglevel":
		case "loglevel_plex":
			args[i+1] = "debug"
		}
	}
	args = append([]string{wd}, args...)

	log.Debugf("current working directory: %s", wd)

	job := common.Job{
		Args: args,
	}

	executor, err = common.CreateExecutor(*config, job)

	if err != nil {
		log.Fatalf("error creating executor: %s", err.Error())
	}

	log.Infof("created executor: ", executor)

	err = executor.Start()

	if err != nil {
		log.Fatalf("failed to start job: %s", err.Error())
	}

	log.Print("waiting for build pod to enter Running state...")

	err = executor.WaitForState(common.ExecutorPhaseRunning)
	if err != nil {
		log.Fatal("Error waiting for pod to enter running state: ", err)
	}

	log.Infof("job has started running...")
	log.Infof("waiting for job to complete...")

	err = executor.WaitForState(common.ExecutorPhaseSucceeded)
	if err != nil {
		log.Fatalf("error waiting for job to complete: %s", err.Error())
	}

	log.Infof("job completed. cleaning up...")

	err = executor.Stop()
	if err != nil {
		log.Fatalf("error stopping job: %s", err.Error())
	}

	log.Printf("cleaned up successfully")
}
