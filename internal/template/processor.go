package template

import (
	"os"
	"io/ioutil"
	"errors"
	"log"
	"encoding/json"
	"net/http"
	"strconv"
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
	Alerts       string
	Login        string
	Passwd       string
}
type Jira struct {
	Jira_api     string
	Tmpl_dir     string 
	Tmpl_src     string
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

func request(method, url string, data []byte, login, passwd string) ([]byte, error){

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

func New(filename string, tmpl *Template) (*Template, error) {
	f, err := os.Open(filename)
	if err != nil {
		return tmpl, err
	}
	defer f.Close()

	if err := toml.NewDecoder(f).Decode(&tmpl); err != nil {
		return tmpl, err
	}

	return tmpl, nil
}

func (tl *Template) getAlerts(cfg *config.Config) ([]interface{}, error) {
	
	body, err := request("GET", tl.Alerts.Alerts, nil, tl.Alerts.Login, tl.Alerts.Passwd)
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
		"int": func(s string) (int, error){
			c, err := strconv.Atoi(s)
			if err != nil {
				return 0, err
			}
			return c, nil
		},
	}

	tmpl, err := template.New(tl.Jira.Tmpl_src).Funcs(funcMap).ParseFiles(tl.Jira.Tmpl_dir+"/"+tl.Jira.Tmpl_src)
	if err != nil {
		return nil, err
	}

	var tpl bytes.Buffer
	if err = tmpl.Execute(&tpl, &alert); err != nil {
		return nil, err
	}

	var dat interface{}
	if err := json.Unmarshal(tpl.Bytes(), &dat); err != nil {
		return nil, err
	}

	return tpl.Bytes(), nil
}

func (tl *Template) createTask(data []byte) (*Create, error) {

    var resp *Create

	body, err := request("POST", tl.Jira.Jira_api, data, tl.Jira.Login, tl.Jira.Passwd)
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
	template := &Template{
		Alerts: Alerts{
			Login:      cfg.Alerts.Login,
			Passwd:     cfg.Alerts.Passwd,
		},
		Jira: Jira{
            Jira_api:   cfg.Jira.Jira_api,
			Tmpl_dir:   cfg.Server.Conf_dir+"/templates/",
			Login:      cfg.Jira.Login,
			Passwd:     cfg.Jira.Passwd,
		},
	}
	for _, p := range paths {
		go func(filename string, cfg *config.Config, clnt db.DbClient, template *Template, test *string){

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

				//getting group id
				a := alrt.(map[string]interface{})
				groupId, ok := a["groupId"].(string)
				if !ok {
					log.Print("[error] undefined groupId field")
					continue
				}
				
				//get a record from the database
				ltask, err := clnt.LoadTask(groupId)
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

					//created new task
					ctask, err := tmpl.createTask(data)
					if err != nil {
						log.Printf("[error] %v", err)
						continue
					}

					//set a record from the database
					stask := &config.Task{
						Group_id:  groupId,
						Task_id:   ctask.Id,
						Task_key:  ctask.Key,
						Task_self: ctask.Self,
					}
					if err := clnt.SaveTask(*stask); err != nil {
						log.Printf("[error] %v", err)
					}
					log.Printf("[info] task saved: %s", ctask.Self)
				}
			}

		}(p.Name(), cfg, clnt, template, test)
	}
	
	return nil
}