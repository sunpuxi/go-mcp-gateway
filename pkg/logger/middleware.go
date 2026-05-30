package logger

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

// responseWriter 包装 http.ResponseWriter，捕获状态码，并透传 Flusher 以支持 SSE
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// Flush 实现 http.Flusher，支持 SSE 等流式传输
// 嵌入接口类型不会自动提升具体值的额外方法，必须显式代理
func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// ChiMiddleware 返回一个 Chi 兼容的 HTTP 日志中间件
// 每个请求自动注入 trace_id，记录 method、path、status、duration
func ChiMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 优先从请求头提取 trace_id，否则生成新 ID
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// 注入 context 并设置响应头
		ctx := WithTraceID(r.Context(), traceID)
		r = r.WithContext(ctx)
		w.Header().Set("X-Trace-ID", traceID)

		// 包装 ResponseWriter 以捕获状态码
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rw, r)

		// 请求完成后记录
		Log.Infow("request",
			"trace_id", traceID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_addr", r.RemoteAddr,
		)
	})
}
