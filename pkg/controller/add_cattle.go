package controller

import (
	"github.com/chechiachang/cattle-operator/pkg/controller/cattle"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, cattle.Add)
}
