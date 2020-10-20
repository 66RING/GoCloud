package main

import (
	"GoCloud/handler"
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/suc", handler.UploadSucHandler)
	http.HandleFunc("/file/fastupload", handler.HTTPInterceptor(handler.TryFastUploadHandler))
	http.HandleFunc("/file/meta", handler.GetFileMetaHandler)
	http.HandleFunc("/file/download", handler.DownloadHandler)
	http.HandleFunc("/file/update", handler.FileMetaUpdateHandler)
	http.HandleFunc("/file/delete", handler.FileDeleteHandler)
	http.HandleFunc("/file/query", handler.FileQueryHandler)
	http.HandleFunc("/file/downloadurl", handler.DownloadURLhandler)

	http.HandleFunc("/user/signup", handler.SignupHandler)
	http.HandleFunc("/user/signin", handler.SignInHandler)
	http.HandleFunc("/user/info", handler.UserInfoHandler)

	http.HandleFunc("/file/mpupload/init", handler.InitMultipartUploadHandler)
	http.HandleFunc("/file/mpupload/uppart", handler.UploadPartHandler)
	http.HandleFunc("/file/mpupload/complete", handler.CompleteUploadHandler)

	err := http.ListenAndServe(":8088", nil)
	if err != nil {
		fmt.Println("Fail to start server ", err.Error())
	}
}
