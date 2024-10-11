package common

import logrus "github.com/sirupsen/logrus"

func LogrusInfo(label string, field logrus.Fields) {
	field["status"] = 1
	field["message"] = "success"
	logrus.WithFields(field).Error(label)
}

func LogrusError(label string, field logrus.Fields, err error) {
	field["status"] = 0
	field["message"] = err.Error()
	logrus.WithFields(field).Info(label)
}
