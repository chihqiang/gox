package stringx

import "strings"

// HidePhone 隐藏手机号，保留前三位和后四位
func HidePhone(phone string) string {
	return Hide(phone, 3, 4, '*')
}

// HideEmail 隐藏邮箱，保留@前两位，@后不变
func HideEmail(email string) string {
	at := strings.Index(email, "@")
	if at <= 2 {
		return email
	}
	// 使用 Hide 函数隐藏 @ 前的字符，只保留前 2 位，后缀长度为 0
	return Hide(email[:at], 2, 0, '*') + email[at:]
}

// HideIDCard 隐藏身份证号，保留前六位和后四位
func HideIDCard(id string) string {
	return Hide(id, 6, 4, '*')
}

// HideBankCard 隐藏银行卡号，保留前四位和后四位
func HideBankCard(card string) string {
	return Hide(card, 4, 4, '*')
}

func Hide(s string, prefix, suffix int, mask rune) string {
	runes := []rune(s)
	length := len(runes)

	// 字符串太短，直接全部用掩码
	if length <= prefix+suffix {
		return strings.Repeat(string(mask), length)
	}
	// 中间固定 4 个掩码
	middle := strings.Repeat(string(mask), 4)
	return string(runes[:prefix]) + middle + string(runes[length-suffix:])
}
