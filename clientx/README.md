`clientx` 是一个Go语言实现的增强型HTTP客户端库，基于Go标准库`net/http`包进行封装，提供了更简洁、更强大的HTTP请求功能。

## 功能特点

- 支持所有HTTP方法（GET, POST, PUT, DELETE, HEAD, OPTIONS, CONNECT, TRACE）
- 内置请求重试机制和可自定义的退避策略
- 支持上下文（context）控制请求超时和取消
- 提供函数式配置选项，使用更灵活
- 支持常见数据格式：JSON、表单、多部分表单（文件上传）
- 并发安全的客户端管理
- 可自定义HTTP客户端配置
- 支持中间件机制，可灵活扩展请求处理流程
- 自定义错误类型，提供更详细的错误信息


## 使用示例

### 基本用法

#### 发送GET请求

```go
import (
    "context"
    "fmt"
    "github.com/chihqiang/gox/clientx"
)

func main() {
    ctx := context.Background()
    resp, err := clientx.Get(ctx, "https://api.example.com/users")
    if err != nil {
        fmt.Printf("请求失败: %v\n", err)
        return
    }
    defer resp.Body.Close()
    
    // 处理响应...
}
```

#### 发送POST请求

```go
import (
    "context"
    "fmt"
    "github.com/chihqiang/gox/clientx"
)

func main() {
    ctx := context.Background()
    data := []byte(`{"name":"张三","age":30}`)
    
    resp, err := clientx.Post(ctx, "https://api.example.com/users", data, 
        clientx.WithHeaders(map[string]string{
            "Content-Type": "application/json",
        })
    )
    if err != nil {
        fmt.Printf("请求失败: %v\n", err)
        return
    }
    defer resp.Body.Close()
    
    // 处理响应...
}
```

### 高级用法

#### 使用JSON请求

```go
import (
    "context"
    "fmt"
    "github.com/chihqiang/gox/clientx"
)

func main() {
    ctx := context.Background()
    user := struct {
        Name string `json:"name"`
        Age  int    `json:"age"`
    }{
        Name: "张三",
        Age:  30,
    }
    
    resp, err := clientx.PostJSON(ctx, "https://api.example.com/users", user)
    if err != nil {
        fmt.Printf("请求失败: %v\n", err)
        return
    }
    defer resp.Body.Close()
    
    // 处理响应...
}
```

#### 发送表单数据

```go
import (
    "context"
    "fmt"
    "net/url"
    "github.com/chihqiang/gox/clientx"
)

func main() {
    ctx := context.Background()
    form := url.Values{}
    form.Add("username", "admin")
    form.Add("password", "123456")
    
    resp, err := clientx.PostForm(ctx, "https://api.example.com/login", form)
    if err != nil {
        fmt.Printf("请求失败: %v\n", err)
        return
    }
    defer resp.Body.Close()
    
    // 处理响应...
}
```

#### 文件上传

```go
import (
    "context"
    "fmt"
    "github.com/chihqiang/gox/clientx"
)

func main() {
    ctx := context.Background()
    
    // 打开文件
    file, err := clientx.OpenFile("avatar", "/path/to/avatar.jpg")
    if err != nil {
        fmt.Printf("打开文件失败: %v\n", err)
        return
    }
    
    // 准备上传数据
    uploadData := clientx.UploadFields{
        Fields: map[string]string{
            "username": "admin",
            "desc":     "用户头像",
        },
        Files: []clientx.File{file},
    }
    
    // 发送多部分表单请求
    resp, err := clientx.PostMForm(ctx, "https://api.example.com/upload", uploadData)
    if err != nil {
        fmt.Printf("上传失败: %v\n", err)
        return
    }
    defer resp.Body.Close()
    
    // 处理响应...
}
```

#### 自定义请求配置

```go
import (
    "context"
    "fmt"
    "time"
    "github.com/chihqiang/gox/clientx"
)

func main() {
    ctx := context.Background()
    
    // 设置自定义重试策略和请求头
    resp, err := clientx.Get(ctx, "https://api.example.com/users",
        clientx.WithRetries(5), // 设置最大重试次数为5
        clientx.WithHeaders(map[string]string{
            "Authorization": "Bearer token123",
            "User-Agent":    "MyApp/1.0",
        }),
        clientx.WithBackoff(func(attempt int) time.Duration {
            // 自定义退避函数：线性增长
            return time.Duration(attempt*1000) * time.Millisecond
        }),
    )
    if err != nil {
        fmt.Printf("请求失败: %v\n", err)
        return
    }
    defer resp.Body.Close()
    
    // 处理响应...
}
```

#### 使用上下文控制超时

```go
import (
    "context"
    "fmt"
    "time"
    "github.com/chihqiang/gox/clientx"
)

func main() {
    // 创建一个5秒超时的上下文
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    resp, err := clientx.Get(ctx, "https://api.example.com/slow-endpoint")
    if err != nil {
        fmt.Printf("请求失败: %v\n", err)
        return
    }
    defer resp.Body.Close()
    
    // 处理响应...
}
```

#### 自定义HTTP客户端

```go
import (
    "crypto/tls"
    "net/http"
    "time"
    "github.com/chihqiang/gox/clientx"
)

func main() {
    // 创建自定义HTTP客户端
    customClient := &http.Client{
        Transport: &http.Transport{
            MaxIdleConns:        200,
            IdleConnTimeout:     120 * time.Second,
            TLSHandshakeTimeout: 15 * time.Second,
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: false, // 生产环境不应该跳过验证
            },
        },
        Timeout: 30 * time.Second,
    }
    
    // 设置为默认客户端
    clientx.SetClient(customClient)
    
    // 后续所有请求都会使用这个自定义客户端
    // ...
}
```

#### 使用中间件

```go
import (
    "context"
    "fmt"
    "log"
    "time"
    "net/http"
    "github.com/chihqiang/gox/clientx"
)

func main() {
    // 创建一个日志中间件
    logMiddleware := func(next func(*http.Request) (*http.Response, error)) func(*http.Request) (*http.Response, error) {
        return func(req *http.Request) (*http.Response, error) {
            start := time.Now()
            log.Printf("开始请求: %s %s", req.Method, req.URL.String())
            
            resp, err := next(req)
            
            duration := time.Since(start)
            if err != nil {
                log.Printf("请求失败: %s %s, 错误: %v, 耗时: %v", req.Method, req.URL.String(), err, duration)
            } else {
                log.Printf("请求完成: %s %s, 状态码: %d, 耗时: %v", req.Method, req.URL.String(), resp.StatusCode, duration)
            }
            
            return resp, err
        }
    }
    
    ctx := context.Background()
    
    // 使用中间件发送请求
    resp, err := clientx.Get(ctx, "https://api.example.com/users",
        clientx.WithMiddleware(logMiddleware),
    )
    if err != nil {
        fmt.Printf("请求失败: %v\n", err)
        return
    }
    defer resp.Body.Close()
    
    // 处理响应...
}
```

#### 处理自定义HTTP错误

```go
import (
    "context"
    "fmt"
    "github.com/chihqiang/gox/clientx"
)

func main() {
    ctx := context.Background()
    
    resp, err := clientx.Get(ctx, "https://api.example.com/nonexistent")
    if err != nil {
        // 检查是否为自定义HTTP错误
        if httpErr, ok := err.(*clientx.HTTPError); ok {
            fmt.Printf("HTTP错误: 状态码=%d, 方法=%s, URL=%s, 响应体=%s\n", 
                httpErr.StatusCode, httpErr.Method, httpErr.URL, string(httpErr.Body))
        } else {
            fmt.Printf("其他错误: %v\n", err)
        }
        return
    }
    defer resp.Body.Close()
    
    // 处理响应...
}
```

## API参考

### 核心函数

#### 通用请求函数

```go
func Request(ctx context.Context, method, url string, body []byte, opts ...OptionFunc) (*http.Response, error)
```

#### HTTP方法函数

```go
func Get(ctx context.Context, url string, opts ...OptionFunc) (*http.Response, error)
func Post(ctx context.Context, url string, body []byte, opts ...OptionFunc) (*http.Response, error)
func Put(ctx context.Context, url string, body []byte, opts ...OptionFunc) (*http.Response, error)
func Delete(ctx context.Context, url string, body []byte, opts ...OptionFunc) (*http.Response, error)
func Head(ctx context.Context, url string, opts ...OptionFunc) (*http.Response, error)
func Options(ctx context.Context, url string, body []byte, opts ...OptionFunc) (*http.Response, error)
func Connect(ctx context.Context, url string, body []byte, opts ...OptionFunc) (*http.Response, error)
func Trace(ctx context.Context, url string, body []byte, opts ...OptionFunc) (*http.Response, error)
```

#### 数据格式专用函数

```go
func PostJSON(ctx context.Context, url string, payload any, opts ...OptionFunc) (*http.Response, error)
func PostForm(ctx context.Context, url string, form url.Values, opts ...OptionFunc) (*http.Response, error)
func PostMForm(ctx context.Context, url string, data UploadFields, opts ...OptionFunc) (*http.Response, error)
```

#### 文件处理函数

```go
func OpenFile(fieldName, filename string) (File, error)
```

#### 客户端管理函数

```go
func SetClient(client *http.Client)
func GetClient() *http.Client
```

### 配置选项

```go
func WithRetries(n int) OptionFunc
func WithBackoff(f BackoffFunc) OptionFunc
func WithHeaders(h map[string]string) OptionFunc
func WithForceRetry() OptionFunc
func WithMiddleware(mw Middleware) OptionFunc
func WithMaxIdleConns(n int) OptionFunc
func WithMaxConnsPerHost(n int) OptionFunc
func WithIdleConnTimeout(d time.Duration) OptionFunc
func WithTimeout(timeout time.Duration) OptionFunc
```

### 数据结构

```go
// Option 包含HTTP请求的配置选项
type Option struct {
    Retries     int
    Backoff     BackoffFunc
    Headers     map[string]string
    ForceRetry  bool
    Middlewares []Middleware
}

// BackoffFunc 定义重试间隔的计算函数
type BackoffFunc func(attempt int) time.Duration

// Middleware 定义中间件类型
type Middleware func(next func(*http.Request) (*http.Response, error)) func(*http.Request) (*http.Response, error)

// HTTPError 自定义HTTP错误类型
type HTTPError struct {
    StatusCode int
    Method     string
    URL        string
    Body       []byte
    Err        error
}

// File 表示上传的文件
type File struct {
    FieldName string
    FileName  string
    File      io.Reader
}

// UploadFields 包含多部分表单的字段和文件
type UploadFields struct {
    Fields map[string]string
    Files  []File
}
```

## 默认配置

- 默认重试次数：3次
- 默认退避策略：指数退避，初始500ms，最大30s
- 默认超时时间：10秒
- 默认TLS配置：跳过证书验证（注意：生产环境应修改此配置）
- 默认会自动重试的HTTP方法：GET, HEAD, PUT, DELETE
- 默认中间件链：空（无中间件）

## 注意事项

1. 始终记得关闭响应体`resp.Body`，以避免资源泄漏
2. 在生产环境中，建议自定义TLS配置，不要跳过证书验证
3. 对于包含敏感数据的请求，确保使用HTTPS
4. 使用上下文（context）来控制长请求的超时和取消
5. 对于非幂等的HTTP方法（如POST），默认不会自动重试，除非使用`WithForceRetry()`选项
6. 中间件的执行顺序是后进先出（LIFO），即最后添加的中间件最先执行
7. 自定义错误类型`HTTPError`提供了更详细的错误信息，包括状态码、URL、方法和响应体内容

## 依赖

- Go标准库：`net/http`, `context`, `crypto/tls`, `io`, `bytes`, `encoding/json`, `mime/multipart`, `sync`, `time`, `runtime`等