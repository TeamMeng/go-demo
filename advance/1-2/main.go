package main

import (
	"fmt"
	"net/http"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 在请求进入实际处理函数前打印访问日志。
		fmt.Println("request start", r.RequestURI)
		next.ServeHTTP(w, r)
		// 在处理完成后打印结束日志，便于观察一次请求的完整生命周期。
		fmt.Println("request finished", r.Body)
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/students", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// 当前接口只允许 GET，请求成功时返回一个固定的学生示例数据。
		if r.Method == http.MethodGet {
			fmt.Fprintf(w, `{"id": 1, "name": "ZhangSan", "age": 20}`)
			return
		}
		// 其他 HTTP 方法统一返回 405。
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// 将日志中间件包裹到路由处理器外层后再启动服务。
	handler := loggingMiddleware(mux)

	fmt.Println("server is starting on :8080")

	if err := http.ListenAndServe(":8080", handler); err != nil {
		fmt.Println("server start failed:", err)
	}
}
