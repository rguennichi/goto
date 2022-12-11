package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

const (
	Version = 1
)

type Manifest struct {
	Version      int          `yaml:"version"`
	Servers      Servers      `yaml:"servers"`
	Applications Applications `yaml:"applications"`
}

type Servers map[string]*Server

type Server struct {
	Username     string       `yaml:"username"`
	Port         string       `yaml:"port"`
	Environments Environments `yaml:"environments"`
}

type Environments map[string]*Environment

type Environment struct {
	Hosts []string `yaml:"hosts"`
}

type Applications map[string]*Application

type Application struct {
	Server   Server   `yaml:"server"`
	Username string   `yaml:"username"`
	Path     string   `yaml:"path"`
	Scripts  []Script `yaml:"scripts"`
}

type Script struct {
	Name string `yaml:"name"`
	Exec string `yaml:"exec"`
	Desc string `yaml:"desc"`
}

func Parse(path string) (m *Manifest, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	m = &Manifest{}
	err = yaml.Unmarshal(data, m)
	if err != nil {
		return
	}

	// Validate.
	if m.Version != Version {
		err = fmt.Errorf("incompatible goto config version: file=%v current=%v", m.Version, Version)
	}

	return
}
