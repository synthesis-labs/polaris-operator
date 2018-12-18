package controller

import (
	"github.com/synthesis-labs/polaris-operator/pkg/controller/polarisstack"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, polarisstack.Add)
}
