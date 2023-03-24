package main

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Watch []Watch `toml:"watch"`
}

type Watch struct {
	Path    string `toml:"path"`
	Prefix  string `toml:"prefix"`
	HookUrl string `toml:"hook_url"`
}

func loadConfig(p string) *Config {
	tml, err := ioutil.ReadFile(p)
	if err != nil {
		panic(err)
	}
	var conf Config
	err = toml.Unmarshal(tml, &conf)
	if err != nil {
		panic(err)
	}

	return &conf
}
