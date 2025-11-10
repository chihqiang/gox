# httpx

HTTP响应处理工具包，提供了简洁、高效的HTTP响应格式化和发送功能，支持统一的响应格式和错误处理。

## 功能特点

- **统一响应格式**：提供标准的响应结构，包含业务代码、消息和数据
- **多格式支持**：支持JSON和XML两种常用数据格式的响应处理
- **自定义错误处理**：内置业务错误类型，便于统一错误返回格式
- **性能优化**：使用缓冲区池减少内存分配，提高响应性能
- **类型安全**：利用Go泛型提供类型安全的响应处理


## 核心概念

### 基础响应结构

httpx定义了统一的响应结构`BaseResponse`，包含三个主要字段：
- `Code`：业务状态码（非HTTP状态码）
- `Msg`：业务消息
- `Data`：业务数据（可选）

对于XML响应，提供了`BaseXmlResponse`结构，包含XML头部信息和基础响应内容。

### 业务错误

通过`CodeMsg`自定义错误类型，实现了`error`接口，可以方便地在业务层定义和传递错误信息。

### 状态码常量

定义了常用的业务状态码常量：
- `BusinessCodeOK` (0)：表示成功状态
- `BusinessMsgOk` ("ok")：默认成功消息
- `BusinessCodeError` (-1)：表示错误状态

## 使用示例

### 1. JSON响应

```go
import (
    "net/http"
    "github.com/chihqiang/gox/httpx"
)

// 处理函数示例
func handleUserInfo(w http.ResponseWriter, r *http.Request) {
    // 准备响应数据
    userInfo := map[string]interface{}{
        "id":   1,
        "name": "张三",
        "age":  30,
    }
    
    // 发送JSON响应，自动封装为标准格式
    // 返回: {"code":0, "msg":"ok", "data":{"id":1, "name":"张三", "age":30}}
    httpx.JsonResponse(w, userInfo)
}
```

### 2. 错误响应

```go
import (
    "net/http"
    "github.com/chihqiang/gox/httpx"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
    // 验证请求参数
    if r.URL.Query().Get("id") == "" {
        // 创建业务错误
        err := httpx.NewCodeMsg(40001, "缺少id参数")
        // 发送错误响应
        // 返回: {"code":40001, "msg":"缺少id参数"}
        httpx.JsonResponse(w, err)
        return
    }
    
    // 处理请求...
}
```

### 3. 自定义状态码

```go
import (
    "net/http"
    "github.com/chihqiang/gox/httpx"
)

func handleResource(w http.ResponseWriter, r *http.Request) {
    // 检查资源是否存在
    resourceExists := false
    
    if !resourceExists {
        // 设置HTTP 404状态码
        httpx.NotFound(w)
        // 发送错误信息
        httpx.JsonResponse(w, httpx.NewCodeMsg(40400, "资源不存在"))
        return
    }
    
    // 处理请求...
}
```

### 4. XML响应

```go
import (
    "net/http"
    "github.com/chihqiang/gox/httpx"
)

func handleXmlRequest(w http.ResponseWriter, r *http.Request) {
    // 准备XML响应数据
    product := struct {
        ID    int    `xml:"id"`
        Name  string `xml:"name"`
        Price float64 `xml:"price"`
    }{
        ID:    1001,
        Name:  "示例产品",
        Price: 99.99,
    }
    
    // 发送XML响应，自动封装为标准格式
    httpx.XmlResponse(w, product)
    // 输出: <?xml version="1.0" encoding="UTF-8"?><xml><code>0</code><msg>ok</msg><data><id>1001</id><name>示例产品</name><price>99.99</price></data></xml>
}
```

## 最佳实践

1. **统一错误处理**：使用`CodeMsg`类型定义应用程序中的业务错误，保持错误格式一致性

2. **适当设置HTTP状态码**：结合业务响应和HTTP状态码，提供更清晰的API语义

3. **使用类型安全的响应函数**：优先使用`JsonResponse`和`XmlResponse`函数，它们会自动处理标准响应格式

4. **错误包装**：处理错误时保留原始错误信息，便于调试和问题排查
