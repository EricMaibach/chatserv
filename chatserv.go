package main

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"

	"log"

	"net/http"

	"github.com/kardianos/service"
)

var mode = flag.String("mode", "run", "Application mode.  Valid values are run, install, and uninstall.")
var logger service.Logger

type Config struct {
	WebsitePath, HTTPPort string
}

type program struct {
	config *Config
}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func (p *program) run() {
	var err error
	p.config, err = getConfig()
	if err != nil {
		logger.Error("Error getting config: ", err)
	}

	hub := newHub()
	go hub.run()
	http.Handle("/", http.FileServer(http.Dir(p.config.WebsitePath)))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	httpaddr := ":" + p.config.HTTPPort
	err2 := http.ListenAndServe(httpaddr, nil)
	if err2 != nil {
		log.Fatal("ListenAndServe: ", err2)
	}
}

func main() {
	flag.Parse()
	svcConfig := &service.Config{
		Name:        "ChatServ",
		DisplayName: "Web socket chat server",
		Description: "Simple web socket chat server",
	}
	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	switch *mode {
	case "run":
		err = s.Run()
	case "install":
		err = s.Install()
	case "uninstall":
		err = s.Uninstall()
	default:
		err = s.Run()
	}
	if err != nil {
		logger.Error(err)
	}
}

func getConfig() (*Config, error) {
	ex, err := os.Executable()
	if err != nil {
		return nil, err
	}
	logger.Info("Got file path: ", ex)
	dir, _ := filepath.Split(ex)
	logger.Info("Got folder: ", dir)
	path := filepath.Join(dir, "chatserv.config.json")
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	conf := &Config{}

	r := json.NewDecoder(f)
	err = r.Decode(&conf)
	if err != nil {
		return nil, err
	}

	if conf.HTTPPort == "" {
		conf.HTTPPort = "8080"
	}

	return conf, nil
}
