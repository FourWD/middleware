package common

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Database *gorm.DB
var DatabaseSql *sql.DB

type DNS struct {
	Username string
	Password string
	Database string
	IP       string
	Instance string
}

func CreateDSN(isGCP bool, dsn DNS) string {
	var protocol string
	setting := "?charset=utf8mb4&parseTime=True"
	if isGCP {
		protocol = fmt.Sprintf("unix(/cloudsql/%s)", dsn.Instance)
	} else {
		protocol = fmt.Sprintf("tcp(%s:3306)", dsn.IP)
		setting += "&loc=Local"
	}
	log.Printf("Database: %s", dsn.Database)
	return fmt.Sprintf("%s:%s@%s/%s%s", dsn.Username, dsn.Password, protocol, dsn.Database, setting)
}

func ConnectDatabase(dns string) error {
	var err error

	Database, err = gorm.Open(mysql.Open(dns+"&loc=Asia%2FBangkok"), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})

	if err != nil {
		PrintError(`Gorm`, `Connection Error !`)
		panic(err)
	}

	DatabaseSql, err = sql.Open("mysql", dns+"&loc=Asia%2FBangkok")
	if err != nil {
		PrintError(`DB Mysql`, `Connection Error !`)
		panic(err)
	}

	timeZone := "Asia/Bangkok"
	Database.Raw("SET time_zone=?", timeZone)
	DatabaseSql.Exec("SET time_zone=?", timeZone)

	log.Println("CONNECT DB GOOGLE SUCCESS")

	return nil
}

func ConnectDatabaseMySqlGoogle(DNS DNS) (*sql.DB, error) {
	isGCP := true
	if App.Env == "local" {
		isGCP = false
	}

	dsn := CreateDSN(isGCP, DNS)

	database, err := sql.Open("mysql", dsn+"&loc=Asia%2FBangkok")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	timeZone := "Asia/Bangkok"
	database.Exec("SET time_zone=?", timeZone)

	log.Println("CONNECT DB MYSQL SUCCESS")

	return database, nil
}

func ConnectDatabaseViper() error {
	dns := DNS{
		Username: viper.GetString("database.username"),
		Password: viper.GetString("database.password"),
		Database: viper.GetString("database.database"),
		IP:       viper.GetString("database.ip"),
		Instance: viper.GetString("database.instance"),
	}

	// isGCP := false
	// if viper.GetString("production") == "true" {
	// 	isGCP = true
	// }

	isGCP := true
	if App.Env == "local" {
		isGCP = false
	}

	return ConnectDatabase(CreateDSN(isGCP, dns))
}

func DBCreate(requestID string, model interface{}) error {
	data, _ := toMap(model)
	logData := map[string]interface{}{
		"data": data,
	}

	err := Database.Create(model).Error

	if err != nil {
		logData["status"] = "error"
		logData["error"] = err
		LogError("DBCreate", logData, requestID)
	} else {
		logData["status"] = "success"
		Log("DBCreate", logData, requestID)
	}

	return err
}

func DBUpdate(requestID string, model interface{}) error {
	data, _ := toMap(model)
	logData := map[string]interface{}{
		"data": data,
	}

	err := Database.Updates(model).Error

	if err != nil {
		logData["status"] = "error"
		logData["error"] = err
		LogError("DBUpdate", logData, requestID)
	} else {
		logData["status"] = "success"
		Log("DBUpdate", logData, requestID)
	}

	return err
}

func DBUpdateField(requestID string, model any, id string, updateData map[string]interface{}) error {
	for key, value := range updateData {
		if floatValue, ok := value.(float64); ok {
			updateData[key] = parseToFloat(fmt.Sprintf("%.6f", floatValue))
		}
	}

	logData := map[string]interface{}{
		"data": updateData,
	}

	err := Database.Model(model).Where("id = ?", id).Updates(updateData).Error

	if err != nil {
		logData["status"] = "error"
		logData["error"] = err
		LogError("DBUpdateField", logData, requestID)
	} else {
		logData["status"] = "success"
		Log("DBUpdateField", logData, requestID)
	}

	return err
}

func DBDelete(requestID string, model any, id string, DeletedBy string) error {
	updateData := map[string]interface{}{}
	updateData["deleted_at"] = time.Now()
	updateData["deleted_by"] = DeletedBy
	return DBUpdateField(requestID, model, id, updateData)
}

func parseToFloat(str string) float64 {
	parsedValue, err := strconv.ParseFloat(str, 64)
	if err != nil {
		// fmt.Println("Error parsing float:", err)
		return 0
	}
	return parsedValue
}

func toMap(v interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}
	return result, nil
}
