package common

import (
	"fmt"
	"time"

	"github.com/FourWD/middleware/model"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func Send(c *fiber.Ctx) error {
	type Payload struct {
		AppID           string `json:"app_id"`
		SendToUserToken string `json:"send_to_user_token"`
	}
	type NotiPayload struct {
		NotificationTitle string `json:"notification_title"`
		NotificationText  string `json:"notification_text"`
	}
	payload := new(Payload)
	err := c.BodyParser(payload)
	if err != nil {
		return FiberError(c, "1002", "review your payload")
	}

	app, err := GetNotificationApp(payload.AppID)

	if err != nil {
		FiberError(c, "1002", "review your payload")
	}

	notificationPayload := new(NotiPayload)
	notificationPayload.NotificationTitle = "OmakaseCar Auction"
	notificationPayload.NotificationText = "รอบประมูล TIL วันที่ 20/11/2023 กำลังจะเริ่ม"

	notification := new(model.LogNotificationRequest)

	notification.ID = uuid.NewString()
	notification.CreatedAt = time.Now()
	notification.AppID = app.ID
	notification.AppID = payload.AppID
	notification.SendToUserToken = payload.SendToUserToken
	notification.IsSend = false
	notification.Payload = fmt.Sprintf(`{
		"notification_title":"%s",
		"notification_text":"%s"
	}`, notificationPayload.NotificationTitle, notificationPayload.NotificationText)
	notification.SendDate = NilDate()
	Database.Save(notification)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": 1, "message": "success", "data": "jsonBody"})
}

func GetNotificationApp(appID string) (model.AppNotification, error) {
	app := new(model.AppNotification)
	app.AppKey = "999999999"
	app.AppSecret = "f51sdgs1d8fsgdf85g"
	return *app, nil
}
