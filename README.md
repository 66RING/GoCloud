## 开发日志/技术总结 

### http包

```go
func handlerFunc(w http.ResponseWriter, r *http.Request) {}

http.HandleFunc("/router/path", handlerFunc)

err := http.ListenAndServe(":port", nil)
```

- 使用`io.WriteString(w, "inter err")`往(http)流里写入
- `r.Method`可查看request类型

