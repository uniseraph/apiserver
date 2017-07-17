package handlers_test

import (
	"github.com/zanecloud/apiserver/handlers"
	"testing"
)

func TestSystemAuditHandler(t *testing.T) {
	t.Run("SA=1", func(t *testing.T) {
		req := handlers.GetSystemAuditListRequest{
			StartTime: 1400113570,
			EndTime:   1500113570,
		}

		rsp := handlers.GetSystemAuditListResponse{}

		if err := postTestRequest("logs/list", req, &rsp); err != nil {
			t.Error(err.Error())
		} else {
			t.Log(rsp)
		}
	})
}
