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

type CmdConfig struct {
	Cmd    string
	Params ConfigParameters
}

func ReadConfig(path string) (config []CmdConfig) {
	cfgfile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Error reading config: ", err.Error())
	}

	cfg := make(ConfigMap)

	err = yaml.Unmarshal(cfgfile, &cfg)
	if err != nil {
		log.Fatal("Error parsing config: ", err.Error())
	}

	sortedkeys := make([]string, 0, len(cfg))

	for k := range cfg {
		sortedkeys = append(sortedkeys, k)
	}

	sort.Strings(sortedkeys)

	for _, k := range sortedkeys {
		config = append(config, CmdConfig{k, cfg[k]})
	}

	return config
}
