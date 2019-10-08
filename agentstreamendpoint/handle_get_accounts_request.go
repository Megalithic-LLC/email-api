package agentstreamendpoint

import (
	"github.com/docktermj/go-logger/logger"
	"github.com/on-prem-net/email-api/agentstreamendpoint/emailproto"
	"github.com/on-prem-net/email-api/model"
)

func (self *AgentStream) handleGetAccountsRequest(requestId uint64, getAccountsReq emailproto.GetAccountsRequest) {
	logger.Tracef("AgentStream:handleGetAccountsRequest(%d)", requestId)

	accounts := []model.Account{}
	searchFor := &model.Account{}
	if err := self.endpoint.db.Where(searchFor).Find(&accounts).Error; err != nil {
		logger.Errorf("Failed loading accounts: %v", err)
		self.SendErrorResponse(requestId, err)
		return
	}

	self.SendGetAccountsResponse(requestId, accounts)
}
