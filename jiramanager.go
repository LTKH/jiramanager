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
    "github.com/ltkh/jiramanager/internal/db"
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

    // Connection to data base
    _, err = db.NewClient(cfg.DB); 
    if err != nil {
        log.Fatalf("[error] %v", err)
    }

    // Creating api v1
    apiV1, err := v1.New(cfg)
    if err != nil {
        log.Fatalf("[error] %v", err)
    }

    // Enabled listen port
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

    /*

    // Checking the status of tasks
    if cfg.Server.Check_enabled {

        go func(cfg *config.Config, clnt db.DbClient) {

            if cfg.Server.Check_interval == 0 {
                cfg.Server.Check_interval = 600
            }

            for {

                //geting tasks from database
                tasks, err := clnt.LoadTasks()
                if err != nil {
                    log.Printf("[error] %v", err)
                    continue
                }

                for _, task := range tasks {
                    time.Sleep(cfg.Server.Check_delay * time.Second)

                    //task.Updated = time.Now().UTC().Unix()
                    issue, err := template.UpdateTaskStatus(task, cfg, clnt)
                    if err != nil {
                        log.Printf("[error] task update id %s: %v", task.Task_id, err)
                        continue
                    }

                    if issue.Fields.Status.Id != task.Status_id {
                        log.Printf("[info] task status updated: %s", task.Task_self)
                    }

                    if task.Updated + cfg.Server.Check_resolve < time.Now().UTC().Unix() {
                        for _, s := range cfg.Server.Check_status {
                            if issue.Fields.Status.Id == s {
                                if err := clnt.DeleteTask(task.Group_id); err != nil {
                                    log.Printf("[error] task delete id %s: %v", task.Task_id, err)
                                    continue
                                }
                                log.Printf("[info] task is removed from the database: %v", task.Task_self)
                            }
                        }
                    }

                }

                time.Sleep(cfg.Server.Check_interval * time.Second)
            }
        }(&cfg, client)
    }

    if cfg.Server.Alerts_interval == 0 {
        cfg.Server.Alerts_interval = 600
    }

    
    */
}
