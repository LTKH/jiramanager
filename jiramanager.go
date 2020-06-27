package main

import (
	//"net/http"
	//"crypto/tls"
	"time"
	"log"
	"os"
	"os/signal"
	"syscall"
	//"encoding/json"
	//"io/ioutil"
	//"bytes"
	"runtime"
	//"reflect"
	//"text/template"
	"flag"
	//"regexp"
	//"strconv"
	"gopkg.in/natefinch/lumberjack.v2"
	"github.com/ltkh/jiramanager/internal/db"
	"github.com/ltkh/jiramanager/internal/config"
	"github.com/ltkh/jiramanager/internal/template"
)

func main() {

  	//limits the number of operating system threads
	runtime.GOMAXPROCS(runtime.NumCPU())

	//command-line flag parsing
	cfFile := flag.String("config", "", "config file")
	lgFile := flag.String("logfile", "", "log file") 
	flTest := flag.String("test", "", "config test") 
	flag.Parse()

	//loading configuration file
	cfg, err := config.New(*cfFile)
	if err != nil {
		log.Fatalf("[error] %v", err)
	}

	//connection to data base
	client, err := db.NewClient(&cfg.DB); 
	if err != nil {
		log.Fatalf("[error] %v", err)
	}

	log.Print("[info] jiramanager running ^_-")

	if *lgFile != "" {
		if cfg.Server.Log_max_size == 0 {
			cfg.Server.Log_max_size = 1
		}
		if cfg.Server.Log_max_backups == 0 {
			cfg.Server.Log_max_backups = 3
		}
		if cfg.Server.Log_max_age == 0 {
			cfg.Server.Log_max_age = 28
		}
		log.SetOutput(&lumberjack.Logger{
			Filename:   *lgFile,
			MaxSize:    cfg.Server.Log_max_size,    // megabytes after which new file is created
			MaxBackups: cfg.Server.Log_max_backups, // number of backups
			MaxAge:     cfg.Server.Log_max_age,     // days
			Compress:   cfg.Server.Log_compress,    // using gzip
		})
	}

	//program completion signal processing
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<- c
		log.Print("[info] jiramanager stopped")
		os.Exit(0)
	}()

	//daemon mode
	for {

		if err := template.Process(&cfg, client, flTest); err != nil {
			log.Printf("[error] %v", err)
		}
        /*
		//
		body, err := newRequest(cfg, "GET", cfg.Jiramanager.Get_alerts, []byte(""), "", "")
		if err != nil {
			log.Printf("[error] %v", err)
		} else {

			//
			log.Print("[debug] parsing alerts")
			var dat []map[string]interface{}
			if err := json.Unmarshal(body, &dat); err != nil {
				log.Printf("[error] %v", err)
			}

			//
			for _, alrt := range dat {
				if reflect.TypeOf(alrt).Kind() == reflect.Map {
					createTask(cfg, "default", alrt)
				}
			}
		}
		*/

		time.Sleep(cfg.Alerts.Interval * time.Second)
	}
}
