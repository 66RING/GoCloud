package ceph

import (
	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
)

var cephConn *s3.S3

func GetCephConnection() *s3.S3 {
	if cephConn != nil {
		return cephConn
	}
	// 初始化ceph
	aws.Auth{
		AccessKey: "",
		SecretKey: "",
	}

	// 创建s3类型的连接

}
