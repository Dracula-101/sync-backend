package base

import "sync-backend/arch/network"

type baseController struct {
	MessageSender
	network.BaseController
}

func NewBaseController(basePath string) BaseController {
	return &baseController{
		MessageSender:  NewMessageSender(),
		BaseController: network.NewBaseController(basePath),
	}
}
