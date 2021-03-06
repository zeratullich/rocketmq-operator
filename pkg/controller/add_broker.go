package controller

import (
	"github.com/zeratullich/rocketmq-operator/pkg/controller/broker"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, broker.Add)
}
