# stringx

`stringx` 是一个 Go 语言实现的字符串处理工具包，提供了常用的字符串处理函数，如敏感信息隐藏、字符串分割等功能。

## 功能特点

- 敏感信息隐藏（手机号、邮箱、身份证号、银行卡号）
- 字符串分割与去重
- 简洁易用的API设计
- 基于Go标准库，仅依赖少量第三方库

## 文件结构

- `hide.go`: 提供敏感信息隐藏相关函数
- `split.go`: 提供字符串分割相关函数

## 敏感信息隐藏函数

### HidePhone
隐藏手机号，保留前三位和后四位。

```go
func HidePhone(phone string) string
```

**参数:**
- `phone`: 手机号字符串

**返回值:**
- 隐藏处理后的手机号字符串

**示例:**
```go
import "github.com/chihqiang/gox/stringx"

phone := "13812345678"
hiddenPhone := stringx.HidePhone(phone)
// hiddenPhone = "138****5678"
```

### HideEmail
隐藏邮箱，保留@前两位，@后不变。

```go
func HideEmail(email string) string
```

**参数:**
- `email`: 邮箱字符串

**返回值:**
- 隐藏处理后的邮箱字符串

**示例:**
```go
import "github.com/chihqiang/gox/stringx"

email := "user@example.com"
hiddenEmail := stringx.HideEmail(email)
// hiddenEmail = "us****@example.com"
```

### HideIDCard
隐藏身份证号，保留前六位和后四位。

```go
func HideIDCard(id string) string
```

**参数:**
- `id`: 身份证号字符串

**返回值:**
- 隐藏处理后的身份证号字符串

**示例:**
```go
import "github.com/chihqiang/gox/stringx"

idCard := "110101199001011234"
hiddenIDCard := stringx.HideIDCard(idCard)
// hiddenIDCard = "110101****1234"
```

### HideBankCard
隐藏银行卡号，保留前四位和后四位。

```go
func HideBankCard(card string) string
```

**参数:**
- `card`: 银行卡号字符串

**返回值:**
- 隐藏处理后的银行卡号字符串

**示例:**
```go
import "github.com/chihqiang/gox/stringx"

bankCard := "6222021234567890123"
hiddenBankCard := stringx.HideBankCard(bankCard)
// hiddenBankCard = "6222****0123"
```

### Hide（内部函数）
通用的字符串隐藏函数，可自定义前缀保留长度、后缀保留长度和掩码字符。

```go
func Hide(s string, prefix, suffix int, mask rune) string
```

**参数:**
- `s`: 要处理的字符串
- `prefix`: 保留的前缀长度
- `suffix`: 保留的后缀长度
- `mask`: 用于隐藏的掩码字符

**返回值:**
- 隐藏处理后的字符串

## 字符串分割函数

### Split
分割字符串并去除每个部分的空白字符，忽略空字符串。

```go
func Split(s, sep string) []string
```

**参数:**
- `s`: 要分割的字符串
- `sep`: 分隔符

**返回值:**
- 分割后的非空字符串数组

**示例:**
```go
import "github.com/chihqiang/gox/stringx"

input := "a,b,,c,d  "
result := stringx.Split(input, ",")
// result = []string{"a", "b", "c", "d"}
```

### SplitUniq
分割字符串数组中的每个字符串，并合并结果，去除重复项。

```go
func SplitUniq(ss []string, sep string) []string
```

**参数:**
- `ss`: 字符串数组
- `sep`: 分隔符

**返回值:**
- 去重后的分割结果数组

**示例:**
```go
import "github.com/chihqiang/gox/stringx"

inputs := []string{"a,b,c", "b,c,d"}
result := stringx.SplitUniq(inputs, ",")
// result = []string{"a", "b", "c", "d"}
```

## 依赖

- `github.com/samber/lo`: 提供数组去重功能
- Go标准库 `strings`

## 注意事项

- 敏感信息隐藏函数假设输入格式基本符合要求，不会进行严格的格式验证
- SplitUniq 函数依赖第三方库 samber/lo 提供去重功能