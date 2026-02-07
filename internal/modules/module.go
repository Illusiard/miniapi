package modules

import "github.com/Illusiard/miniapi/internal/caps"

type Module interface {
	Name() string
	Register(s caps.Setup) error
}

type Spec struct {
	Module      Module
	WithStore   bool
	Description string
	Version     string
}
