package common

import logrus "github.com/sirupsen/logrus"

func LogrusInfo(label string, fields logrus.Fields) {
	fields["status"] = 1
	fields["message"] = "success"
	logrus.WithFields(fields).Error(label)
}

func LogrusError(label string, fields logrus.Fields, err error) {
	fields["status"] = 0
	fields["message"] = err.Error()
	logrus.WithFields(fields).Info(label)
}
