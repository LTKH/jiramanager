package template

import (
	"os"
	"io/ioutil"
	"errors"
	"log"
	"sync"
	"encoding/json"
	"net/http"
	"bytes"
	"text/template"
	"crypto/tls"
	"github.com/naoina/toml"
	"github.com/ltkh/jiramanager/internal/db"
	"github.com/ltkh/jiramanager/internal/config"
)

type Template struct {
	Alerts       Alerts
	Jira         Jira
}

type Alerts struct {
	Api          string
	Login        string
	Passwd       string
}

type Jira struct {
	Api          string
	Dir          string 
	Src          string
	Login        string
	Passwd       string
}

type Data struct {
	Status       string                  
	Error        string                  
	Data struct {
		Alerts   []interface{}
	}            
}

type Create struct {
	Id           string
	Key          string
	Self         string
}

type Issue struct {
	Fields struct {
		Status struct {
			Id   string
			Name string
		}
	}
}

func Request(method, url string, data []byte, login, passwd string) ([]byte, error){

	//ignore certificate
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

    req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	if login != "" && passwd != "" {
        req.SetBasicAuth(login, passwd)
	}

	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return nil, errors.New(string(body))
	}

	return body, nil
}

func New(filename string, tmpl Template) (*Template, error) {
	f, err := os.Open(filename)
	if err != nil {
		return &tmpl, err
	}
	defer f.Close()

	if err := toml.NewDecoder(f).Decode(&tmpl); err != nil {
		return &tmpl, err
	}

	return &tmpl, nil
}

func (tl *Template) getAlerts(cfg *config.Config) ([]interface{}, error) {
	
	body, err := Request("GET", tl.Alerts.Api, nil, tl.Alerts.Login, tl.Alerts.Passwd)
    if err != nil {
		return nil, err
	}
	
	var resp Data
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, errors.New(string(body))
	}

	return resp.Data.Alerts, nil
}

func (tl *Template) newTemplate(alert interface{}) ([]byte, error) {

	funcMap := template.FuncMap{
		"int": tmpl_int,
		"float": tmpl_float,
		"add": tmpl_add,
	}

	tmpl, err := template.New(tl.Jira.Src).Funcs(funcMap).ParseFiles(tl.Jira.Dir+"/"+tl.Jira.Src)
	if err != nil {
		return nil, err
	}

	var tpl bytes.Buffer
	if err = tmpl.Execute(&tpl, &alert); err != nil {
		return nil, err
	}

	return tpl.Bytes(), nil
}

func (tl *Template) createTask(data []byte) (*Create, error) {

    var resp *Create

	body, err := Request("POST", tl.Jira.Api, data, tl.Jira.Login, tl.Jira.Passwd)
    if err != nil {
		return resp, err
	}
	
	if err := json.Unmarshal(body, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func Process(cfg *config.Config, clnt db.DbClient, test *string) error {

    paths, err := ioutil.ReadDir(cfg.Server.Conf_dir+"/conf.d")
	if err != nil {
		return err
	}
	
	if len(paths) < 1 {
		return errors.New("found no templates")
	}

	template := Template{
		Alerts: Alerts{
			Login:      cfg.Alerts.Login,
			Passwd:     cfg.Alerts.Passwd,
		},
		Jira: Jira{
            Api:        cfg.Jira.Api,
			Dir:        cfg.Server.Conf_dir+"/templates/",
			Login:      cfg.Jira.Login,
			Passwd:     cfg.Jira.Passwd,
		},
	}

	var wg sync.WaitGroup

	for _, p := range paths {

        wg.Add(1)

		go func(wg *sync.WaitGroup, filename string, cfg *config.Config, clnt db.DbClient, template Template, test *string){

			defer wg.Done()

			tmpl, err := New(cfg.Server.Conf_dir+"/conf.d/"+filename, template)
			if err != nil {
				log.Printf("[error] %v", err)
				return
			}

			alrts, err := tmpl.getAlerts(cfg)
			if err != nil {
				log.Printf("[error] %v", err)
				return
			}

			for _, alrt := range alrts {

				a := alrt.(map[string]interface{})

				if a["groupId"] == "" {
					log.Print("[error] undefined field groupId")
					continue
				}
				
				//get a record from the database
				ltask, err := clnt.LoadTask(a["groupId"].(string))
				if err != nil {
					log.Printf("[error] %v", err)
					continue
				}
				
				if ltask.Group_id == "" {
					//generate template
					data, err := tmpl.newTemplate(alrt)
					if err != nil {
						log.Printf("[error] %v", err)
					    continue
					}

					//test
					if *test != "" {
						log.Printf("[test] %v", string(data))
						continue
					}

					var dat interface{}
					if err := json.Unmarshal(data, &dat); err != nil {
						log.Printf("[warning] %v: %v", tmpl.Jira.Src, err)
						log.Printf("%#v", data)
						continue
					}

					//created new task
					ctask, err := tmpl.createTask(data)
					if err != nil {
						log.Printf("[error] %v", err)
						continue
					}
					log.Printf("[info] task created in jira: %s", ctask.Self)

					//set a record from the database
					stask := &config.Task{
						Group_id:  a["groupId"].(string),
						Task_id:   ctask.Id,
						Task_key:  ctask.Key,
						Task_self: ctask.Self,
						Template:  tmpl.Jira.Src,
					}
					if err := clnt.SaveTask(*stask); err != nil {
						log.Printf("[error] %v", err)
					}
					log.Printf("[info] task saved to database: %s", ctask.Self)
				}
			}

		}(&wg, p.Name(), cfg, clnt, template, test)
	}

	wg.Wait()
	
	return nil
}