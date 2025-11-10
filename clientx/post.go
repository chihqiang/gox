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

// Post wraps a standard POST request
// ctx: Context for timeout control or cancellation
// url: Request URL
// body: Request body byte data
// opts: Optional configurations (retry, Headers, Middleware, etc.)
func Post(ctx context.Context, url string, body []byte, opts ...OptionFunc) (*http.Response, error) {
	return Request(ctx, http.MethodPost, url, body, opts...)
}

// PostJSON sends a JSON request
// Automatically serializes payload and sets Content-Type to application/json
func PostJSON(ctx context.Context, url string, payload any, opts ...OptionFunc) (*http.Response, error) {
	// Serialize JSON
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("JSON serialization failed: %w", err)
	}
	// Set Content-Type: application/json
	headerOpt := WithHeaders(map[string]string{
		"Content-Type": "application/json",
	})
	// Call base Post method to send request
	return Post(ctx, url, data, append(opts, headerOpt)...)
}

// PostForm sends a form request with application/x-www-form-urlencoded content type
func PostForm(ctx context.Context, url string, form url.Values, opts ...OptionFunc) (*http.Response, error) {
	// Set Content-Type
	headerOpt := WithHeaders(map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
	// Encode form into URL query string format
	return Post(ctx, url, []byte(form.Encode()), append(opts, headerOpt)...)
}

// File defines a single uploaded file structure
type File struct {
	FieldName string    // Form field name
	FileName  string    // File name, sent to server
	File      io.Reader // File content, can be *os.File or other io.Reader
}

// FormData wraps multiple files and form fields
type FormData struct {
	Fields map[string]string // Regular form fields
	Files  []File            // File list
}

// OpenFile opens a local file and returns a File object
// fieldName: Form field name
// filename: Local file path
// Returns a File object that can be directly passed to PostMForm
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

// PostMForm supports multipart/form-data uploads, including files and form fields
func PostMForm(ctx context.Context, url string, data FormData, opts ...OptionFunc) (*http.Response, error) {
	// Validation: must have at least files or form fields
	if (data.Files == nil || len(data.Files) == 0) && (data.Fields == nil || len(data.Fields) == 0) {
		return nil, fmt.Errorf("upload failed: Files and Fields cannot be empty at the same time")
	}

	// Buffer to store multipart data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Write file part
	for _, f := range data.Files {
		// Create form file
		part, err := writer.CreateFormFile(f.FieldName, f.FileName)
		if err != nil {
			return nil, fmt.Errorf("create form file %s failed: %w", f.FileName, err)
		}
		// Copy file content to multipart
		if _, err := io.Copy(part, f.File); err != nil {
			return nil, fmt.Errorf("copy file %s failed: %w", f.FileName, err)
		}
		// Close file after upload (if it implements io.Closer)
		if closer, ok := f.File.(io.Closer); ok {
			_ = closer.Close()
		}
	}

	// Write regular form fields
	if data.Fields != nil {
		for k, v := range data.Fields {
			if err := writer.WriteField(k, v); err != nil {
				return nil, fmt.Errorf("write form field %s failed: %w", k, err)
			}
		}
	}

	// Close writer to generate multipart boundary
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer failed: %w", err)
	}

	// Set Content-Type to multipart/form-data
	headerOpt := WithHeaders(map[string]string{
		"Content-Type": writer.FormDataContentType(),
	})

	// Call Post to send request
	return Post(ctx, url, buf.Bytes(), append(opts, headerOpt)...)
}
