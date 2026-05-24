package common

import (
	"github.com/FourWD/middleware/infra"
	"github.com/FourWD/middleware/orm"
	"github.com/google/uuid"
)

func CreateLogAction(userID string, remark string, remarkKey string) error {
	logAction := orm.LogAction{
		ID:        uuid.NewString(),
		UserID:    userID,
		Remark:    remark,
		RemarkKey: remarkKey,
	}

	infra.AppLog.Event("LOG_ACTION_CREATE", map[string]any{
		"action_id":  logAction.ID,
		"user_id":    userID,
		"remark_key": remarkKey,
	}, logAction.ID,
		infra.WithComponent(infra.ComponentDB),
		infra.WithOperation("create"),
		infra.WithLogKind(infra.LogKindBusiness),
		infra.WithField("table", "log_actions"))

	return Database.Create(&logAction).Error
}
