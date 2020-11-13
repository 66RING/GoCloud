package middleware

import (
	"io"
	"io/ioutil"
	"net/http"
)

type middleware func(http.HandlerFunc) http.HandlerFunc

type Router struct {
	middlewareChain []middleware
	router          map[string]http.HandlerFunc
}

func New() *Router {
	return &Router{
		router: make(map[string]http.HandlerFunc),
	}
}

func (r *Router) Use(m middleware) {
	r.middlewareChain = append(r.middlewareChain, m)
}

func (r *Router) merge(h http.HandlerFunc) http.HandlerFunc {
	mergeHandler := h
	l := len(r.middlewareChain)
	if l > 0 {
		for i := l - 1; i >= 0; i-- {
			mergeHandler = r.middlewareChain[i](mergeHandler)
		}
	}
	return mergeHandler
}

func (r *Router) ANY(route string, h http.HandlerFunc) {
	mergeHandler := r.merge(h)
	r.router[route] = mergeHandler
	http.HandleFunc(route, mergeHandler)
}

func (r *Router) GET(route string, h http.HandlerFunc) {
	mergeHandler := r.merge(h)
	mergeHandler = func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				//返回一个静态页面
				w.WriteHeader(http.StatusMethodNotAllowed) // 405
				data, err := ioutil.ReadFile("./static/view/index.html")
				if err != nil {
					io.WriteString(w, "inter err")
					return
				}
				io.WriteString(w, string(data))
			} else {
				h(w, r)
			}
		}
	}(h)
	r.router[route] = mergeHandler
	http.HandleFunc(route, mergeHandler)
}

func (r *Router) POST(route string, h http.HandlerFunc) {
	mergeHandler := r.merge(h)
	mergeHandler = func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				//返回一个静态页面
				w.WriteHeader(http.StatusMethodNotAllowed) // 405
				data, err := ioutil.ReadFile("./static/view/index.html")
				if err != nil {
					io.WriteString(w, "inter err")
					return
				}
				io.WriteString(w, string(data))
			} else {
				h(w, r)
			}
		}
	}(h)
	r.router[route] = mergeHandler
	http.HandleFunc(route, mergeHandler)
}
