package common

import (
	"time"

	logrus "github.com/sirupsen/logrus"
)

func LogrusInfo(label string, fields logrus.Fields) {
	fields["created"] = time.Now().Format(DATE_FORMAT_NANO)
	fields["status"] = 1
	fields["message"] = "success"
	AppLog.WithFields(fields).Error(label)
}

func LogrusError(label string, fields logrus.Fields, err error) {
	fields["created"] = time.Now().Format(DATE_FORMAT_NANO)
	fields["status"] = 0
	fields["message"] = err.Error()
	AppLog.WithFields(fields).Info(label)
}
