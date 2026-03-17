package util

import (
	"math/rand"
	"strconv"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GenerateRandomInt 生成随机整数
func GenerateRandomInt(min, max int) int {
	return min + rand.Intn(max-min+1)
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// StringToInt 字符串转整数
func StringToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// StringToInt64 字符串转int64
func StringToInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

// IntToString 整数转字符串
func IntToString(i int) string {
	return strconv.Itoa(i)
}

// Int64ToString int64转字符串
func Int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

// Float64ToString float64转字符串
func Float64ToString(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

// StringToFloat64 字符串转float64
func StringToFloat64(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// Now 获取当前时间字符串
func Now() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// Date 获取当前日期字符串
func Date() string {
	return time.Now().Format("2006-01-02")
}

// Timestamp 获取当前时间戳
func Timestamp() int64 {
	return time.Now().Unix()
}

// MilliTimestamp 获取当前毫秒时间戳
func MilliTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatDate 格式化日期
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// ParseTime 解析时间字符串
func ParseTime(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
}

// ParseDate 解析日期字符串
func ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, time.Local)
}

// GetCurrentTimestamp 获取当前时间戳（秒）
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}
