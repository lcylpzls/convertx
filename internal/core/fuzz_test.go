package core

import "testing"

// FuzzToInt 验证 string → int 转换对任意输入不 panic。
func FuzzToInt(f *testing.F) {
	seeds := []string{
		"42", "-1", "0", "abc", "1.5", "99999999999999999999",
		" 42 ", "0x1F", "t", "", "true", "+7", "1e3", "中文",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		_, _ = To[int](s)
	})
}

// FuzzToTime 验证 string → time 转换对任意输入不 panic。
func FuzzToTime(f *testing.F) {
	seeds := []string{
		"2026-08-17T10:00:00Z", "2026-08-17 10:00:00", "bad", "",
		"17 Aug 26 10:00 UTC", "2026-08-17", "2026-08-17T10:00:00.123456+08:00",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		_, _ = ToTime(s)
	})
}
