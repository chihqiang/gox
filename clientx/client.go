package clientx

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	mu sync.RWMutex
	// 自定义 HTTP 客户端
	defaultClient = &http.Client{
		Transport: &http.Transport{
			// 自动从环境变量读取代理，例如 HTTP_PROXY / HTTPS_PROXY
			Proxy: http.ProxyFromEnvironment,

			// 全局最大空闲连接数，适合高并发场景
			MaxIdleConns: 100,

			// 空闲连接存活时间，超过该时间未使用就关闭，释放资源
			IdleConnTimeout: 90 * time.Second,

			// TLS 握手超时，超过该时间未完成握手就报错
			TLSHandshakeTimeout: 10 * time.Second,

			// HTTP/1.1 Expect: 100-continue 超时时间
			ExpectContinueTimeout: 1 * time.Second,

			// 每个主机的最大空闲连接数，根据 CPU 核心数动态设置
			MaxIdleConnsPerHost: runtime.GOMAXPROCS(0) + 1,

			// TLS 配置
			TLSClientConfig: &tls.Config{
				//跳过证书验证
				InsecureSkipVerify: true,
				// 如果要安全，可以使用自定义 CA：
				// RootCAs: x509.NewCertPool()
			},

			// 是否禁用 Keep-Alive，false 表示启用 TCP 连接复用，提高性能
			DisableKeepAlives: false,
		},

		// 请求超时（包括连接、发送请求、读取响应）
		Timeout: 10 * time.Second,
	}
	// bufferPool 用于大请求体复用
	bufferPool = sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
)

// Middleware 定义中间件类型，可在请求前/后执行逻辑
type Middleware func(next func(*http.Request) (*http.Response, error)) func(*http.Request) (*http.Response, error)

// SetClient 替换全局 HTTP 客户端
func SetClient(client *http.Client) {
	mu.Lock()
	defer mu.Unlock()
	defaultClient = client
}

// GetClient 获取全局 HTTP 客户端
func GetClient() *http.Client {
	mu.RLock()
	defer mu.RUnlock()
	return defaultClient
}

// Option 请求配置结构
type Option struct {
	Retries     int               // 最大重试次数
	Backoff     BackoffFunc       // 重试退避策略
	Headers     map[string]string // 自定义请求头
	ForceRetry  bool              // 是否强制所有方法重试
	Middlewares []Middleware      // 中间件链
}

// BackoffFunc 定义重试退避函数
type BackoffFunc func(attempt int) time.Duration

// 默认指数退避，最大 30s
func defaultBackoff(attempt int) time.Duration {
	d := time.Duration(1<<attempt) * 500 * time.Millisecond
	if d > 30*time.Second {
		return 30 * time.Second
	}
	return d
}

// OptionFunc 函数式配置类型
type OptionFunc func(*Option)

// WithRetries 设置最大重试次数
func WithRetries(n int) OptionFunc {
	return func(o *Option) { o.Retries = n }
}

// WithBackoff 设置退避策略
func WithBackoff(f BackoffFunc) OptionFunc {
	return func(o *Option) { o.Backoff = f }
}

// WithHeaders 添加请求头
func WithHeaders(h map[string]string) OptionFunc {
	return func(o *Option) {
		if o.Headers == nil {
			o.Headers = make(map[string]string)
		}
		for k, v := range h {
			o.Headers[k] = v
		}
	}
}

// WithForceRetry 强制所有方法都进行重试
func WithForceRetry() OptionFunc {
	return func(o *Option) { o.ForceRetry = true }
}

// WithMiddleware 添加中间件
func WithMiddleware(mw Middleware) OptionFunc {
	return func(o *Option) { o.Middlewares = append(o.Middlewares, mw) }
}

// WithTimeout 设置客户端超时时间
func WithTimeout(timeout time.Duration) OptionFunc {
	return func(o *Option) {
		client := GetClient()
		client.Timeout = timeout
	}
}

// WithMaxIdleConns 设置全局最大空闲连接数
func WithMaxIdleConns(n int) OptionFunc {
	return func(o *Option) {
		client := GetClient()
		if t, ok := client.Transport.(*http.Transport); ok {
			t.MaxIdleConns = n
		}
	}
}

// WithMaxConnsPerHost 设置每个主机最大连接数
func WithMaxConnsPerHost(n int) OptionFunc {
	return func(o *Option) {
		client := GetClient()
		if t, ok := client.Transport.(*http.Transport); ok {
			t.MaxConnsPerHost = n
		}
	}
}

// WithIdleConnTimeout 设置空闲连接超时时间
func WithIdleConnTimeout(d time.Duration) OptionFunc {
	return func(o *Option) {
		client := GetClient()
		if t, ok := client.Transport.(*http.Transport); ok {
			t.IdleConnTimeout = d
		}
	}
}

// HTTPError 自定义请求错误类型，包含状态码、方法、URL 和响应体
type HTTPError struct {
	StatusCode int
	Method     string
	URL        string
	Body       []byte
	Err        error
}

// Error 实现 error 接口
func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s %s failed: status %d, body: %q, error: %v", e.Method, e.URL, e.StatusCode, e.Body, e.Err)
	}
	return fmt.Sprintf("%s %s failed: status %d, body: %q", e.Method, e.URL, e.StatusCode, e.Body)
}

// Request 核心请求函数，支持重试、退避、中间件、缓冲池
// ctx: 上下文，可用于取消或设置超时
// method: HTTP 方法，如 GET、POST、PUT 等
// urlStr: 请求 URL
// body: 请求体内容，字节切片
// opts: 可选配置，包括重试次数、退避策略、请求头、中间件等
func Request(ctx context.Context, method, urlStr string, body []byte, opts ...OptionFunc) (*http.Response, error) {
	// 初始化默认选项：重试 3 次，默认退避函数
	options := &Option{
		Retries: 3,
		Backoff: defaultBackoff,
	}
	for _, opt := range opts {
		opt(options) // 应用用户传入的可选配置
	}

	var lastErr error // 记录最后一次错误
	for attempt := 0; attempt <= options.Retries; attempt++ {
		var bodyReader io.Reader
		var buf *bytes.Buffer

		// 对请求体进行优化：
		// 小于 1KB 直接用 bytes.NewReader
		// 大于 1KB 使用缓冲池复用，减少内存分配
		if body != nil {
			if len(body) < 1024 {
				bodyReader = bytes.NewReader(body)
			} else {
				buf = bufferPool.Get().(*bytes.Buffer)
				buf.Reset()
				buf.Write(body)
				bodyReader = buf
			}
		}

		// 创建请求对象，绑定上下文
		req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
		if err != nil {
			if buf != nil {
				bufferPool.Put(buf) // 出错时归还缓冲区
			}
			return nil, err
		}

		// 设置请求头
		for k, v := range options.Headers {
			req.Header.Set(k, v)
		}

		// 构建中间件链
		doFunc := GetClient().Do // 默认 HTTP 请求函数
		for i := len(options.Middlewares) - 1; i >= 0; i-- {
			mw := options.Middlewares[i] // 注意闭包捕获问题
			next := doFunc
			doFunc = func(req *http.Request) (*http.Response, error) {
				return mw(next)(req) // 执行中间件
			}
		}

		// 执行请求
		resp, err := doFunc(req)

		// 请求完成后立即归还缓冲池
		if buf != nil {
			bufferPool.Put(buf)
		}

		// 如果请求成功且状态码 2xx，则直接返回
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		// 读取响应体用于错误信息（最多 512 字节）
		var bodyBytes []byte
		if resp != nil {
			bodyBytes, _ = io.ReadAll(io.LimitReader(resp.Body, 512))
			_ = resp.Body.Close()
		}

		// 4xx 错误直接返回，不重试
		if resp != nil && resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return nil, &HTTPError{
				StatusCode: resp.StatusCode,
				Method:     method,
				URL:        urlStr,
				Body:       bodyBytes,
			}
		}

		// 记录最后一次错误
		lastErr = &HTTPError{
			StatusCode: 0,
			Method:     method,
			URL:        urlStr,
			Body:       bodyBytes,
			Err:        err,
		}

		// 判断是否需要重试
		if attempt < options.Retries {
			// 对 GET/HEAD 方法或 ForceRetry 开启重试
			if options.ForceRetry || strings.ToUpper(method) == http.MethodGet || strings.ToUpper(method) == http.MethodHead {
				backoff := options.Backoff(attempt) // 计算退避时间
				select {
				case <-time.After(backoff):
					// 等待退避时间再重试
				case <-ctx.Done():
					// 上下文取消或超时，返回
					return nil, ctx.Err()
				}
			} else {
				break // 非可重试方法直接退出循环
			}
		}
	}

	// 所有重试失败，返回最后一次错误
	return nil, lastErr
}
