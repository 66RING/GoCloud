package handler

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		//返回上传的html页面
		data, err := ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "inter err")
			return
		}
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		// 接受文件流及存储到本地目录
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Println("fail to get data", err.Error())
			return
		}
		defer file.Close()
		newfile, err := os.Create("/tmp/" + head.Filename)
		if err != nil {
			fmt.Println("fail to create file, ", err.Error())
			return
		}
		defer newfile.Close()

		// target, source
		_, err = io.Copy(newfile, file)
		if err != nil {
			fmt.Println("fail to save date, ", err.Error())
			return
		}

		http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
	}
}

// Upload success
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished")
}
