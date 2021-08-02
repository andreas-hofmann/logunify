package main

import (
	"io/ioutil"
	"log"
	"sort"

	"gopkg.in/yaml.v3"
)

type ConfigParameters struct {
	Loop       bool  `yaml:"loop"`
	IntervalMs int64 `yaml:"intervalMs"`
}

type ConfigMap map[string]ConfigParameters

type Config struct {
	Flags map[string]string           `yaml:"flags"`
	Init  map[string]ConfigParameters `yaml:"init"`
	Run   map[string]ConfigParameters `yaml:"runtime"`
}

type CmdConfig struct {
	Cmd    string
	Params ConfigParameters
}

func ReadConfig(flags Flags) (config []CmdConfig, init []string, fl Flags) {
	cfgfile, err := ioutil.ReadFile(flags.ConfigFileName)
	if err != nil {
		log.Fatal("Error reading config: ", err.Error())
	}

	cfg := Config{}

	err = yaml.Unmarshal(cfgfile, &cfg)
	if err != nil {
		log.Fatal("Error parsing config: ", err.Error())
	}

	sortedkeys := make([]string, 0, len(cfg.Run))
	initcmds := make([]string, 0, len(cfg.Init))

	for k := range cfg.Run {
		sortedkeys = append(sortedkeys, k)
	}
	for k := range cfg.Init {
		initcmds = append(initcmds, k)
	}

	sort.Strings(sortedkeys)

	for _, k := range sortedkeys {
		config = append(config, CmdConfig{k, cfg.Run[k]})
	}

	flags.update(cfg.Flags)

	return config, initcmds, flags
}
