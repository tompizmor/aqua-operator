package controller

import (
	"github.com/niso120b/aqua-operator/pkg/controller/aquaenforcer"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, aquaenforcer.Add)
}
