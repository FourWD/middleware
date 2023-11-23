package common

import (
	"github.com/FourWD/middleware/model"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func OtpRequest(payload model.OtpRequestPayload, db gorm.DB) (model.OtpResult, error) {

	result, errUpload := otpRequestToServer(payload, viper.GetString("app_id"), viper.GetString("token.upload"))
	if errUpload != nil {
		return result, errUpload
	}

	// SAVE TO LOG

	// logFile := new(middlewareOrm.File)
	// logFile.ID = result.ID
	// logFile.BucketName = getBucketName(viper.GetString("app_id"))
	// logFile.Cdn = result.Cdn
	// logFile.FileName = result.FileName
	// logFile.Extension = result.Extension
	// logFile.Path = result.Path
	// logFile.FullPath = result.FullPath
	// err := db.Save(&logFile)
	// if err.Error != nil {
	// 	PrintError("error save file", "tb file")
	// }
	return result, nil
}

func otpRequestToServer(payload model.OtpRequestPayload, appID string, token string) (model.OtpResult, error) {
	result := new(model.OtpResult)

	// 	jsonData, err := json.Marshal(payload)

	// 	if err != nil {
	// 		fmt.Println("there was an error with the JSON", err.Error())
	// 		return *result, err
	// 	} else {
	// 		client := &http.Client{}
	// 		uploadUrl := "https://pakwan-service-dot-fourwd.as.r.appspot.com/api/v1/upload/"
	// 		req, err := http.NewRequest("POST", uploadUrl, bytes.NewBuffer(jsonData))
	// 		if err != nil {
	// 			fmt.Println(err)
	// 			return *result, err
	// 		}
	// 		req.Header.Add("Content-Type", "application/json")
	// 		req.Header.Add("Authorization", "Bearer "+token)

	// 		response, err := client.Do(req)
	// 		if err != nil {
	// 			fmt.Println(err)
	// 			return *result, err
	// 		}
	// 		defer response.Body.Close()
	// 		if err != nil {
	// 			fmt.Println("there was an error with the request", err.Error())
	// 			return *result, err
	// 		} else {
	// 			body, _ := io.ReadAll(response.Body)
	// 			var resp model.ApiResponse
	// 			err := json.Unmarshal(body, &resp)
	// 			if err != nil {
	// 				return *result, err
	// 			}
	// 			// // Unmarshal the JSON string into a MenuItem struct
	// 			// errUnmars := json.Unmarshal([]byte(body), &r)
	// 			// if errUnmars != nil {
	// 			// 	fmt.Println("Error:", err)
	// 			// }
	// 			// result.ID = resp.Data.ID
	// 			// result.Cdn = resp.Data.Cdn
	// 			// result.Extension = resp.Data.Extension
	// 			// result.FileName = resp.Data.FileName
	// 			// result.Path = resp.Data.Path
	// 			// result.FullPath = resp.Data.FullPath
	// 		}
	// 	}

	return *result, nil
}
