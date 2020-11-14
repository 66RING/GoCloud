package handler

import (
	dblayer "GoCloud/db"
	"GoCloud/meta"
	"GoCloud/store/oss"
	"GoCloud/util"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// 接受文件流及存储到本地目录
	file, head, err := r.FormFile("file")
	if err != nil {
		fmt.Println("fail to get data", err.Error())
		return
	}
	defer file.Close()

	fileMeta := meta.FileMeta{
		FileName: head.Filename,
		Location: "/home/ring/tem/GoCloud/" + head.Filename,
		UploadAt: time.Now().Format("2006-01-02 15:04:05"), // go里面一个特殊时间点
	}

	newFile, err := os.Create(fileMeta.Location)
	if err != nil {
		fmt.Println("fail to create file, ", err.Error())
		return
	}
	defer newFile.Close()

	// target, source
	fileMeta.FileSize, err = io.Copy(newFile, file)
	if err != nil {
		fmt.Println("fail to save date, ", err.Error())
		return
	}

	// 计算较大文件哈希耗时较长，影响上传速度和用户体验。可以抽离出来做微服务，然后进行异步处理
	newFile.Seek(0, 0)
	fileMeta.FileSha1 = util.FileSha1(newFile)

	// TODO
	// 写入ceph 或 写入OSS

	ossPath := "oss/" + fileMeta.FileSha1
	// err = oss.Bucket().PutObject(ossPath, file)
	err = oss.Bucket().PutObjectFromFile(ossPath, fileMeta.Location)
	if err != nil {
		fmt.Println(err.Error())
		w.Write([]byte("Upload to oss failed"))
		return
	}

	// 更新到数据库同时更新到内存
	meta.UpdateFileMeta(fileMeta)

	fileMeta.Location = ossPath
	ok := meta.UpdateFileMetaDB(fileMeta)
	if !ok {
		fmt.Println("更新文件到数据库失败")
	}

	// 写唯一文件表同时写用户文件表 //
	r.ParseForm()
	username := r.Form.Get("username")
	suc := dblayer.OnUserFileUploadFinished(username, fileMeta.FileSha1,
		fileMeta.FileName, fileMeta.FileSize)
	if suc {
		// http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
	} else {
		w.Write([]byte("Upload Failed."))
	}
}

// Upload success
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished")
}

// 获取文件信息
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form["filehash"][0]
	// fMeta := meta.GetFileMeta(filehash)
	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(fMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// 查询批量的文件元信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	username := r.Form.Get("username")
	userFiles, err := dblayer.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(userFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fsha1 := r.Form.Get("filehash")
	fm := meta.GetFileMeta(fsha1)

	f, err := os.Open(fm.Location)
	if err != nil {
		fmt.Println("Download from local fail: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	fmt.Println("Downloading from local...")
	// 文件很小就直接全部加载到内存了
	data, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println("io error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("Content-Disposition", "attachment;filename=\""+fm.FileName+"\"")

	w.Write(data)
}

func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	opType := r.Form.Get("op")
	fileSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")

	if opType != "0" {
		w.WriteHeader(http.StatusForbidden) // 403
		return
	}

	curFileMeta := meta.GetFileMeta(fileSha1)
	curFileMeta.FileName = newFileName
	meta.UpdateFileMeta(curFileMeta)

	data, err := json.Marshal(curFileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileSha1 := r.Form.Get("filehash")
	username := r.Form.Get("username")
	// 删除先取出位置，在系统中删除
	fMeta := meta.GetFileMeta(fileSha1)
	os.Remove(fMeta.Location)
	ok := dblayer.OnUserFileDelete(username, fileSha1)
	if !ok {
		fmt.Println("Delete from tbl_user_file fail")
	}

	meta.RemoveFileMeta(fileSha1)

	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp.JSONBytes())
}

func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// 1. 解析请求参数
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))

	// 2. 从文件表中查询相同hash的文件记录
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := util.RespMsg{
			Code: -1,
			Msg:  err.Error(),
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 3. 查不到记录则秒传失败
	if fileMeta.FileName == "" {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败, 文件名为空",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 4. 上传过则将文件写入用户文件表，返回成功
	suc := dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(resp.JSONBytes())
		return
	} else {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，写入数据库失败",
		}
		w.Write(resp.JSONBytes())
	}
}

func DownloadURLhandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	row, _ := dblayer.GetFileMeta(filehash)

	// TODO 判断文件存在oss还是ceph

	signedURL := oss.DownloadUrl(row.FileAddr.String)
	fmt.Println("Downloading from oss...")
	w.Write([]byte(signedURL))
}
