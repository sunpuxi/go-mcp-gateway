package logger

import "context"

type ctxKey string

const traceIDKey ctxKey = "trace_id"

// WithTraceID 将 trace_id 注入 context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// TraceIDFromCtx 从 context 提取 trace_id
func TraceIDFromCtx(ctx context.Context) string {
	if v, ok := ctx.Value(traceIDKey).(string); ok {
		return v
	}
	return ""
}

// InfoCtx 带 trace_id 的信息日志
func InfoCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	args := appendTraceID(ctx, keysAndValues)
	Log.Infow(msg, args...)
}

// ErrorCtx 带 trace_id 的错误日志
func ErrorCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	args := appendTraceID(ctx, keysAndValues)
	Log.Errorw(msg, args...)
}

// WarnCtx 带 trace_id 的警告日志
func WarnCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	args := appendTraceID(ctx, keysAndValues)
	Log.Warnw(msg, args...)
}

// DebugCtx 带 trace_id 的调试日志
func DebugCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	args := appendTraceID(ctx, keysAndValues)
	Log.Debugw(msg, args...)
}

func appendTraceID(ctx context.Context, kvs []interface{}) []interface{} {
	if traceID := TraceIDFromCtx(ctx); traceID != "" {
		return append(kvs, "trace_id", traceID)
	}
	return kvs
}
