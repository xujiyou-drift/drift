package controller

import (
	"github.com/xujiyou-drift/drift/pkg/controller/kafka"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, kafka.Add)
}
