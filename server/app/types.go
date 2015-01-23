package app

import (
	"time"
)

type App struct {
	Id          string
	Description string
	Image       string
	Config      *Config
	Source      *Source
	Created     time.Time
	Updated     time.Time
}

type Config struct {
	NumInstances  int
	ContainerPort int
}

type Code struct {
	Lang string
	Text string
}

type GitRepo struct {
	Url    string
	Branch string
}

type Info struct {
	AppId     string
	Status    string
	Reason    string
	Message   string
	Timestamp time.Time
}

type Source struct {
	Code       *Code
	Dockerfile string
	GitRepo    *GitRepo
	Image      string
}
