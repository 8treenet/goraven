package controller

import (
	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/ping", &PingController{})
	})
}

// PingController
type PingController struct {
	Worker freedom.Worker
}

// Get
func (controller *PingController) Get() {
	controller.Worker.IrisContext().WriteString("pong")
}
