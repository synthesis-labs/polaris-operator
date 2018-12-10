package controller

import (
	"github.com/synthesis-labs/polaris-operator/pkg/controller/polarissourcerepository"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, polarissourcerepository.Add)
}
