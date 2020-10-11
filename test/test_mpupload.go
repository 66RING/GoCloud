package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type Data struct {
	UploadID  string
	ChunkSize int
}

type Resp struct {
	Data Data `json:"data"`
}

func main() {
	// TODO
	username := "ring"
	token := "9ad4499266a68148d9f556022d4c6eb25f7b3399"
	filehash := "656109202328d1f49eccabf530beaf7e5291d1fc"
	filesize := "236356"

	// 请求初始化分块上传接口
	resp, err := http.PostForm(
		"http://localhost:8088/file/mpupload/init",
		url.Values{
			"username": {username},
			"token":    {token},
			"filehash": {filehash},
			"filesize": {filesize},
		})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	// 得到uploadID以及服务端指定的分块大小chunkSize
	// TODO
	var data Data
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println(err.Error)
	}
	uploadID := data.UploadID
	chunkSize := data.ChunkSize

	// uploadID := jsonit.Get(body, "data").Get("UploadID").ToString()
	// chunkSize := jsonit.Get(body, "data").Get("ChunkSize").ToInt()
	fmt.Println("uploadID: %s chunkSize: %d \n", uploadID, chunkSize)

	// 请求分块上传接口
	filename := "/tmp/some/where" // TODO
	tURL := "http://localhost:8088/file/mpupload/uppart?" +
		"username=" + username + "&token=" + token + "&uploadid=" + uploadID
	// TODO : 分块上传函数
	multipartUpload(filename, tURL, chunkSize)

	// 请求分块完成接口
	resp, err = http.PostForm(
		"http://localhost:8088/file/mpupload/complete",
		url.Values{
			"username": {username},
			"token":    {token},
			"filehash": {filehash},
			"filesize": {filesize},
			"filename": {"NAME"},
			"uploadid": {uploadID},
		})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

}
