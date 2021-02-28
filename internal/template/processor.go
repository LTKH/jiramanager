package template

import (
	"bytes"
	"strings"
	"text/template"
	//"crypto/tls"
	//"github.com/naoina/toml"
	//"github.com/ltkh/jiramanager/internal/db"
	//"github.com/ltkh/jiramanager/internal/config"
)

type Template struct {
	tmpl   *template.Template
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
		Alerts   []map[string]interface{}
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

var funcMap = template.FuncMap{
	"toInt":           toInt,
	"toFloat":         toFloat,
	"add":             addFunc,
	"replace":         strings.Replace,
	"regexReplace":    regexReplaceAll,
	"lookupIP":        LookupIP,
	"lookupIPV4":      LookupIPV4,
	"lookupIPV6":      LookupIPV6,
	"strQuote":        strQuote,
}

// LoadTemplate reads and parses all templates defined in the given file and constructs a jiralert.Template.
func LoadTemplate(path string) (*Template, error) {
	tmpl, err := template.New(path).Option("missingkey=zero").Funcs(funcMap).ParseFiles(path)
	if err != nil {
		return nil, err
	}
	return &Template{tmpl: tmpl}, nil
}

// Execute parses the provided text (or returns it unchanged if not a Go template), associates it with the templates
// defined in t.tmpl (so they may be referenced and used) and applies the resulting template to the specified data
// object, returning the output as a string .
func (t *Template) Execute(text string, data interface{}) (string, error) {
	if !strings.Contains(text, "{{") {
		return text, nil
	}

	tmpl, err := t.tmpl.Clone()
	if err != nil {
		return "", err
	}

	tmpl, err = tmpl.New("").Parse(text)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

    if err = tmpl.Execute(&buf, &data); err != nil {
		return "", err
	}
	ret := buf.String()
	return ret, nil
}
