package common

import (
	"github.com/FourWD/middleware/orm"
	"github.com/google/uuid"
)

func CreateLogAction(userID string, remark string, remarkKey string) error {
	logAction := new(orm.LogAction)
	logAction.ID = uuid.NewString()
	logAction.UserID = userID
	logAction.Remark = remark
	logAction.RemarkKey = remarkKey

	logData := map[string]interface{}{
		"data": logAction,
	}
	Log("CreateLogAction", logData, logAction.ID)

	return Database.Create(&logAction).Error
}
