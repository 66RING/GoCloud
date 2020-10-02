package handler

import (
	dblayer "GoCloud/db"
	"GoCloud/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	pwdSalt = "<G:~1"
)

// 处理用户注册请求
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data, err := ioutil.ReadFile("")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
		return
	}
	r.ParseForm()
	username := r.Form.Get("username")
	passwd := r.Form.Get("password")

	if len(username) < 3 || len(passwd) < 5 {
		w.Write([]byte("Invalid parameter"))
		return
	}

	// 密码加密
	encPasswd := util.Sha1([]byte(passwd + pwdSalt))
	suc := dblayer.UserSignup(username, encPasswd)
	if suc {
		w.Write([]byte("SUCCESS"))
	} else {
		w.Write([]byte("FAIL"))
	}
	return
}

// 登录接口
func SignInHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	passwd := r.Form.Get("password")

	encPasswd := util.Sha1([]byte(passwd + pwdSalt))

	// 1. 校验密码
	pwdChecked := dblayer.UserSignin(username, encPasswd)
	if !pwdChecked {
		w.Write([]byte("FAILED"))
		return
	}

	// 2. 校验凭证，如token
	token := GenToken(username)
	upRes := dblayer.UpdateToken(username, token)
	if !upRes {
		w.Write([]byte("FAILED"))
		return
	}

	// 3. 成功后的操作：重定向到首页
	// 返回url，跳转留给客户端操作
	w.Write([]byte("http://" + r.Host + "/static/view/index.html"))
}

func GenToken(username string) string {
	// 40位字符: md5(username + timestamp + tokenSalt) + timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}