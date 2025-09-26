package clientx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

// Post 封装普通 POST 请求
// ctx: 上下文，可控制超时或取消
// url: 请求 URL
// body: 请求体字节数据
// opts: 可选配置（重试、Headers、Middleware 等）
func Post(ctx context.Context, url string, body []byte, opts ...OptionFunc) (*http.Response, error) {
	return Request(ctx, http.MethodPost, url, body, opts...)
}

// PostJSON 发送 JSON 请求
// 自动序列化 payload 并设置 Content-Type 为 application/json
func PostJSON(ctx context.Context, url string, payload any, opts ...OptionFunc) (*http.Response, error) {
	// 序列化 JSON
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("JSON serialization failed: %w", err)
	}
	// 设置 Content-Type: application/json
	headerOpt := WithHeaders(map[string]string{
		"Content-Type": "application/json",
	})
	// 调用基础 Post 方法发送请求
	return Post(ctx, url, data, append(opts, headerOpt)...)
}

// PostForm 发送表单请求 application/x-www-form-urlencoded
func PostForm(ctx context.Context, url string, form url.Values, opts ...OptionFunc) (*http.Response, error) {
	// 设置 Content-Type
	headerOpt := WithHeaders(map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
	// 将表单编码成 URL 查询字符串形式
	return Post(ctx, url, []byte(form.Encode()), append(opts, headerOpt)...)
}

// File 定义单个上传文件结构
type File struct {
	FieldName string    // 表单字段名
	FileName  string    // 文件名，上传给服务器
	File      io.Reader // 文件内容，可是 *os.File 或其它 io.Reader
}

// FormData 封装多文件和表单字段
type FormData struct {
	Fields map[string]string // 普通表单字段
	Files  []File            // 文件列表
}

// OpenFile 打开本地文件并返回 File 对象
// fieldName: 表单字段名
// filename: 本地文件路径
// 返回可直接传给 PostMForm 的 File 对象
func OpenFile(fieldName, filename string) (File, error) {
	file, err := os.Open(filename)
	if err != nil {
		return File{}, err
	}
	return File{
		FieldName: fieldName,
		FileName:  filename,
		File:      file,
	}, nil
}

// PostMForm 支持 multipart/form-data 上传，包括文件和表单字段
func PostMForm(ctx context.Context, url string, data FormData, opts ...OptionFunc) (*http.Response, error) {
	// 校验：必须至少有文件或表单字段
	if (data.Files == nil || len(data.Files) == 0) && (data.Fields == nil || len(data.Fields) == 0) {
		return nil, fmt.Errorf("upload failed: Files and Fields cannot be empty at the same time")
	}

	// 缓冲区存放 multipart 数据
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 写入文件部分
	for _, f := range data.Files {
		// 创建 form file
		part, err := writer.CreateFormFile(f.FieldName, f.FileName)
		if err != nil {
			return nil, fmt.Errorf("create form file %s failed: %w", f.FileName, err)
		}
		// 拷贝文件内容到 multipart
		if _, err := io.Copy(part, f.File); err != nil {
			return nil, fmt.Errorf("copy file %s failed: %w", f.FileName, err)
		}
		// 上传完成后关闭文件（如果实现了 io.Closer）
		if closer, ok := f.File.(io.Closer); ok {
			_ = closer.Close()
		}
	}

	// 写入普通表单字段
	if data.Fields != nil {
		for k, v := range data.Fields {
			if err := writer.WriteField(k, v); err != nil {
				return nil, fmt.Errorf("write form field %s failed: %w", k, err)
			}
		}
	}

	// 关闭 writer，生成 multipart 边界
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer failed: %w", err)
	}

	// 设置 Content-Type 为 multipart/form-data
	headerOpt := WithHeaders(map[string]string{
		"Content-Type": writer.FormDataContentType(),
	})

	// 调用 Post 发送请求
	return Post(ctx, url, buf.Bytes(), append(opts, headerOpt)...)
}
