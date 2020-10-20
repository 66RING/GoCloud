package oss

import (
	"GoCloud/config"
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var ossCli *oss.Client

// TODO 可能的并发安全改造
func Client() *oss.Client {
	if ossCli != nil {
		return ossCli
	}
	ossCli, err := oss.New(config.OSSEndpoint, config.OSSAccessKeyID, config.OSSAccessKeySecret)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	return ossCli
}

// 获取bucket存储空间
func Bucket() *oss.Bucket {
	cli := Client()
	if cli != nil {
		bucket, err := cli.Bucket(config.OSSBucket)
		if err != nil {
			fmt.Println("Create Bucket ERROR: ", err.Error())
			return nil
		}
		return bucket
	} else {
		return nil
	}
}

func DownloadUrl(objName string) string {
	signedURL, err := Bucket().SignURL(objName, oss.HTTPGet, 3600)
	if err != nil {
		fmt.Println("Get object ERROR", err.Error())
		return ""
	}
	return signedURL
}
