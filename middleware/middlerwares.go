package middleware

import "net/http"

func Cors(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")  //允许访问所有域
		w.Header().Add("Access-Control-Allow-Headers", "*") //header的类型
		h(w, r)
	}
}
