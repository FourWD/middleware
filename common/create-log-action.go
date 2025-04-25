package common

import (
	"github.com/FourWD/middleware/orm"
	"github.com/google/uuid"
)

func CreateLogAction(userID string, remark string) error {
	logAction := new(orm.LogAction)
	logAction.ID = uuid.NewString()
	logAction.UserID = userID
	logAction.Remark = remark

	return Database.Create(&logAction).Error
}
