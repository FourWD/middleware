package upload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	model "github.com/FourWD/middleware/model"
)

func Upload(u model.UploadPayload) (model.UploadResult, error) {

	r := new(model.UploadResult)

	//start uploading
	uploadUrl := "https://pakwan-service-dot-fourwd.as.r.appspot.com/api/v1/upload/upload"
	type Body struct {
		BucketName string `json:"bucket_name"`
		Path       string `json:"path"`
		FileName   string `json:"filename"`
		FileBase64 string `json:"file_base_64"`
	}

	type Response struct {
		CDN      string `json:"cdn"`
		FileName string `json:"file_name"`
		FileType string `json:"file_type"`
		Path     string `json:"path"`
		FullPath string `json:"full_path"`
	}

	type ApiResponse struct {
		Status  int      `json:"status"`
		Message string   `json:"message"`
		Data    Response `json:"data"`
	}
	var resp ApiResponse
	base64 := strings.Replace(u.FileBase64, "data:image/png;base64,", "", -1)
	base64 = strings.Replace(base64, "data:image/jpeg;base64,", "", -1)
	base64 = strings.Replace(base64, "data:image/jpg;base64,", "", -1)
	bodyInstance := Body{
		BucketName: "fourwd-auction",
		Path:       "uploads",
		FileName:   u.Filename + ".jpeg",
		FileBase64: base64,
	}

	jsonData, err := json.Marshal(bodyInstance)

	if err != nil {
		fmt.Println("there was an error with the JSON", err.Error())
	} else {
		client := &http.Client{}
		req, err := http.NewRequest("POST", uploadUrl, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer PA3KBCDIORypzCzD2fQdaqyLUHpPoM60BEaeP68O1GXmbP7dF0hyOBed9ZRcr6ti")

		response, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		defer response.Body.Close()
		if err != nil {
			fmt.Println("there was an error with the request", err.Error())
		} else {
			body, _ := ioutil.ReadAll(response.Body)
			err := json.Unmarshal(body, &resp)
			if err != nil {
				return *r, err
			}
		}
	}

	//end upload

	// call pakwan upload

	return *r, nil
}
