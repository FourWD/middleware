package common

import (
	"fmt"
)

type DNS struct {
	Username string
	Password string
	Database string
	IP       string
	Instance string
	IsMySql5 bool
}

func CreateMySqlDSN(dsn DNS) string {
	var protocol string
	setting := "?charset=utf8mb4&parseTime=True"

	if dsn.IsMySql5 {
		setting = "?charset=utf8&parseTime=True"
	}

	if dsn.Instance != "" {
		protocol = fmt.Sprintf("unix(/cloudsql/%s)", dsn.Instance)
	} else {
		protocol = fmt.Sprintf("tcp(%s:3306)", dsn.IP)
		setting += "&loc=Local"
	}
	Log("DB_CREATE_DSN", map[string]interface{}{"database": dsn.Database}, "")
	return fmt.Sprintf("%s:%s@%s/%s%s", dsn.Username, dsn.Password, protocol, dsn.Database, setting)
}

func CreatePostgresDSN(dsn DNS) string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", dsn.IP, dsn.Username, dsn.Password, dsn.Database, 5432)
}
