package main

import (
    "time"
    "log"
    "os"
    "os/signal"
    "syscall"
    "runtime"
    "flag"
    "net/http"
    "gopkg.in/natefinch/lumberjack.v2"
    //"github.com/ltkh/jiramanager/internal/db"
    "github.com/ltkh/jiramanager/internal/config"
    "github.com/ltkh/jiramanager/internal/api/v1"
    //"github.com/ltkh/jiramanager/internal/template"
)

func main() {

      // Limits the number of operating system threads
    runtime.GOMAXPROCS(runtime.NumCPU())

    // Command-line flag parsing
    listen          := flag.String("listen-address", ":9097", "listen address")
    cfFile          := flag.String("config", "", "config file")
    lgFile          := flag.String("logfile", "", "log file")
    interval        := flag.Duration("interval", 600, "interval")
    logMaxSize      := flag.Int("log.max-size", 1, "log max size") 
    logMaxBackups   := flag.Int("log.max-backups", 3, "log max backups")
    logMaxAge       := flag.Int("log.max-age", 10, "log max age")
    logCompress     := flag.Bool("log.compress", true, "log compress")
    flag.Parse()

    // Logging settings
    if *lgFile != "" {
        log.SetOutput(&lumberjack.Logger{
            Filename:   *lgFile,
            MaxSize:    *logMaxSize,    // megabytes after which new file is created
            MaxBackups: *logMaxBackups, // number of backups
            MaxAge:     *logMaxAge,     // days
            Compress:   *logCompress,   // using gzip
        })
    }

    // Loading configuration file
    cfg, err := config.New(*cfFile)
    if err != nil {
        log.Fatalf("[error] %v", err)
    }

    // Creating api v1
    apiV1, err := v1.New(cfg)
    if err != nil {
        log.Fatalf("[error] %v", err)
    }

    // Enabled listen port
    http.HandleFunc("/-/healthy", apiV1.ApiHealthy)
    http.HandleFunc("/api/v1/alerts", apiV1.ApiAlerts)

    go func(){
        if err := http.ListenAndServe(*listen, nil); err != nil {
            log.Fatalf("[error] %v", err)
        }
    }()

    log.Print("[info] jiramanager running ^_-")

    // Program completion signal processing
    c := make(chan os.Signal, 2)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <- c
        log.Print("[info] jiramanager stopped")
        os.Exit(0)
    }()

    // Daemon mode
    for {
        //if err := template.Process(&cfg, client, flTest); err != nil {
        //    log.Printf("[error] %v", err)
        //}

        // Updating statuses
        if err := apiV1.UpdateStatus(); err != nil {
            log.Printf("[error] update status %v", err)
        }

        time.Sleep(*interval * time.Second)
    }
}

