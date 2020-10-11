package handler

import (
	rPool "GoCloud/cache/redis"
	dblayer "GoCloud/db"
	"GoCloud/util"
	"fmt"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

// 初始化信息
type MultipartUploadInfo struct {
	FileHash    string
	FileSize    int
	UploadID    string
	ChunckSize  int
	ChunckCount int
}

// 初始化分块上传
func InitMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 解析用户请求
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "params invalid", nil).JSONBytes())
		return
	}

	// 获取redis连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 生成分块上传信息
	upInfo := MultipartUploadInfo{
		FileHash:    filehash,
		FileSize:    filesize,
		UploadID:    username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunckSize:  5 * 1024 * 1024, // 5M
		ChunckCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
	}

	// 将初始化信息写入redis缓存
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "chunkcount", upInfo.ChunckCount)
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "filehash", upInfo.FileHash)
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "filesize", upInfo.FileSize)
	// 也可通过hmset批量

	// 将初始化信息返回给客户端
	w.Write(util.NewRespMsg(0, "OK", upInfo).JSONBytes())
}

// 上传文件
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	// 解析参数
	r.ParseForm()
	// username := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index")

	// 获取redis连接池
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 获取文件句柄，以存储分块内容
	fpath := "/tmp/" + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(fpath), 0744)

	fd, err := os.Create(fpath)
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Upload part faild", nil).JSONBytes())
		return
	}
	defer fd.Close()

	buf := make([]byte, 1024*1024)
	for {
		n, err := r.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}

	// 更新redis缓存
	rConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)

	// 响应
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

// 通知上传合并(上传完成)
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 解析参数
	r.ParseForm()
	upid := r.Form.Get("uploadid")
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize := r.Form.Get("filesize")
	filename := r.Form.Get("filename")

	// 获取redis连接池
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 获取uploadid以查询redis并判断是否所有分块上传完成
	data, err := redis.Values(rConn.Do("HGETALL", "MP_"+upid))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "complete upload failed", nil).JSONBytes())
		return
	}
	totalCount := 0
	chunkCount := 0
	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			chunkCount++
		}
	}

	if totalCount != chunkCount {
		w.Write(util.NewRespMsg(-2, "invalid request", nil).JSONBytes())
		return
	}

	// TODO 合并

	// 更新数据库
	fsize, _ := strconv.Atoi(filesize)
	dblayer.OnFileUploadFinished(filehash, filename, int64(fsize), "fileaddr(TODO)")
	dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(fsize))

	// 响应
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}
