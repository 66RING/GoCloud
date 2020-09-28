## 开发日志/技术总结 

### 技术栈

#### 持久化
    * 不单单保存在内存中，否则每次重启数据都会丢失
    * 使用数据库(这里使用mySQL)进行文件持久化

    
#### 集群
    * 单机搭建集群可以使用docker来启动几个mySQL容器


### API使用

#### io

- `io.Create("path")`, 打开一个文件流
- `io.Copy(newFile, sourceFile) fileSize, error`


#### http包

```go
func handlerFunc(w http.ResponseWriter, r *http.Request) {}

http.HandleFunc("/router/path", handlerFunc)

err := http.ListenAndServe(":port", nil)
```

- 使用`io.WriteString(w, "inter err")`往(http)流里写入
- `r.Method`可查看request类型


##### 文件信息查询

`r *http.Request`会包含请求的所有信息，如form表单等

- `r.ParseForm()`解析表单
    * 通过`r.Form["key"][0]`等操作可以获取表单信息


##### 文件下载

想要浏览器识别文件并下，需要设置header，如:

```go
w.Header().Set("Content-Type", "application/octect-stream")
w.Header().Set("Content-Disposition", "attachment;filename=\""+fm.FileName+"\"")
```


##### 文件更新删除

- 更新/删除系统中的数据
    * 删除操作需要考虑线程安全问题
- 更新/删除内存中的数据


#### Time

time.Now().Format("2006-01-02 15:04:05"), go里面一个特殊时间点

