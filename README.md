## TODO 

- 首页
- 注册页面
- USE GIN TO REBUILD

- token验证 handler/user.go IsTokenValid

- 分块上传取消
- 分块上传进度

- metalink, http, ftp, bt三路下载
- mirror brain, 就近服务器获取

## 开发日志/技术总结 

使用go的http包，未使用框架，只为更好理解底层原理

### 杂项

- 每个文件通过哈希在服务器中保存一份
    * 因此需要检测已存在
- 为了方便外部访问往往在方法(包)内部处理好并创建对应的结构体给外部

### 技术栈

#### 持久化

- 不单单保存在内存中，否则每次重启数据都会丢失
- 使用数据库(这里使用mySQL)进行文件持久化
- 数据库科学建表
    * 如果表很长，需要分库分表，然后这些表再关联起来
        + 水平分表
            + 假设分256张表
            + 可以按文件sha值后两位分
            + 但是扩展存在麻烦，要保证旧的文件能哈希到原有的表上，新的文件又得按新的规则扩展到新的表中
        + 垂直分表

    
#### 集群

- 单机搭建集群可以使用docker来启动几个mySQL容器
- 主从节点搭建
    * 1. 找到要作为主节点的binlog信息
        + `show master status;`
    * 2. 配置从节点的master信息，告诉slave从哪里读取binlog
        + `CHANGE MASTER TO MASTER_HOST='your_master_host',MASTER_USER='user',MASTER_PASSWORD='password',MASTER_LOG_FILE='binlog.xxxx',MASTER_LOG_POS=0;`
        + `CHANGE MASTER TO MASTER_HOST='192.168.123.108',MASTER_USER='reader',MASTER_PASSWORD='reader',MASTER_LOG_FILE='binlog.000004',MASTER_LOG_POS=0;`
    * 3. 启动slave模式
        + `start slave;`
        + 查看slave状况：`show slave status\G;`
            + 如果Slave_IO_Running和Slave_SQL_Running都是Yes则正常运行


#### 帐号系统和鉴权

- 用户登录/注册
- 用户Session持久化鉴权
    * 在Session过期，用户退出登录前都有效
        + 基于token，登录获取token后，以后的每次请求都携带token，就知道有权了
        + 基于session，cookie
- 用户资源隔离
    * 有可能云端只储存了一份文件，一个用户删除了他的云文件不会影响到别人的云文件


#### 分段上传和断点续传

- 分段上传
    * 小文件使用分段上传
    * 并行上传，无序传输
        + 合理设置并行数量
- 断点续传


#### 秒传

原理：要上传的文件之前已经有人上传过了。

- 关键点
    * 文件哈希值
    * 用户文件关联
        + 用户文件表
            + 软删除(文件状态)，资源隔离
            + 通过用户文件表链接到唯一文件表，取出存储地址
        + 唯一文件表
            + 哈希比较的是文件的内容，即使文件名改变也不影响，改变文件名只影响用户文件表的引用


#### Redis缓存

存储每个文件已经上传的元信息，序列号、大小、起始位置等。

- 为何redis
    * 文件信息只需在上传期间保留，完成后清除
    * 缓存
- **redis默认是没有密码的**


#### 分布式存储(Ceph)

本地存储完成后，再(异步)迁移到ceph集群中，以后用户读取时直接中ceph集群中读取

- 存储同时利用每个节点的计算能力
- ceph基础组件
    * OSD：用于数据的储存(储存/复制/平衡/恢复等)
    * Monitor：监控集群状态，维护集群的Map表，保证集群数据一致性
    * MDS(mate data server)：保存文件系统的元数据
    * GW：网关服务，兼容AWS(亚马逊)S3术语
        + Region：存储数据所在的地理区域
        + Endpoint：存储服务入口，Web服务入口的url
        + Bucket：存储桶是S3中用于存储对象的容器
        + Object：S3中的存储基本实体，由对象数据和元数据组成
        + Key：是存储桶中对象的唯一标识符
- 异步处理迁移操作
    * 用户只需等待储存到本地节点的时间




#### (TODO)微服务

计算较大文件哈希耗时较长，影响上传速度和用户体验。可以抽离出来**做微服务，然后进行异步处理**


#### 公有云OSS

当技术/示例有限时也不必每个轮子都自己造

- 可能需要的功能
    * 可能需要公共读bucket
        + 一些公开的文件
    * 对象声明周期的管理
        + 定期删除一些文件


#### Tips
- 同时上传相同文件的冲突
    * 思路1
        + 先完成上传先入库
        + 后上传只更新用户文件表，并删除以上传的文件


### API使用

#### io

- `io.Create("path")`, 打开一个文件流
- `io.Copy(newFile, sourceFile) fileSize, error`


#### go操作mysql

- 通过`sql.DB`来管理数据库的连接对象，`var db *sql.DB`
- 通过`sql.Open`来创建一个协程安全的`sql.DB`对象，`db, err := sql.Open(driverName, dataSourceName)`打开数据库
- `db.Ping()`检查活跃
- 一切ok可以将实例对象返回(db)出去


##### 数据库操作

数据库操作和httpHandler分离，handler完成对应操作后调用一下数据库操作即可。

- 预编译sql语句`stmt, err := db.Prepare("SQL Statement")`
    * 执行
        + `ret, err := stmt.Exec(args)`
            + `ret.RowsAffected()`可以检查操作影响的行数
        + 同样也是执行`err = stmt.QueryRow(args).Scan(&value)`把结果赋给value
    * 用完关闭`defer stmt.Close()`
    * 可以防止sql注入攻击


#### http包

```go
func handlerFunc(w http.ResponseWriter, r *http.Request) {}

http.HandleFunc("/router/path", handlerFunc)

err := http.ListenAndServe(":port", nil)
```

- 使用`io.WriteString(w, "inter err")`往(http)流里写入
- `r.Method`可查看request类型
- 响应数据`json.Marshal(v)`后返回给浏览器


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


##### JSON

需要携带的数据变多，结构复杂时，转换成json类型的字节就传输给浏览器

`r, err := json.Marshal(struct)`


##### 拦截器

使用拦截器，验证token。**拦截器原理类似闭包**：传入什么传出什么，但在中间过程做了些验证

```go
func Validator(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if len(username) < 3{
				w.WriteHeader(http.StatusForbidden)
				return
			}
			h(w, r)
		},
	)
}

http.HandleFunc("/user/info", Validator(UserHandler))
```

- 把一些校验逻辑抽成拦截器
    * handler里就可以专心处理业务了
    * 代码复用，避免handler里验证各自为政


#### Time

time.Now().Format("2006-01-02 15:04:05"), go里面一个特殊时间点


#### 加密

##### md5

- `_md5 := md5.New()`，实例一个对象
- `_md5.Write(data)`，(流)写入要加密的数据(`[]byte`)
- `hex.EncodeToString(_md5.Sum([]byte("")))`，计算并返回字符串
