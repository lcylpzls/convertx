package core

import (
	"math"
	"testing"
	"time"

	"github.com/lcylpzls/testx"
)

func TestToTimeStringFormats(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{"RFC3339", "2026-08-17T10:00:00+08:00"},
		{"RFC3339Nano", "2026-08-17T10:00:00.123456+08:00"},
		{"标准时间", "2026-08-17 10:00:00"},
		{"日期", "2026-08-17"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm, err := ToTime(tt.src, WithDefaultLocation(time.UTC))
			testx.NoError(t, err)
			testx.False(t, tm.IsZero())
		})
	}
}

func TestToTimeRFC822(t *testing.T) {
	tm, err := ToTime("17 Aug 26 10:00 UTC")
	testx.NoError(t, err)
	testx.Equal(t, tm.Year(), 2026)
	testx.Equal(t, tm.Month(), time.August)
	testx.Equal(t, tm.Day(), 17)
}

func TestToTimeWithLayout(t *testing.T) {
	tm, err := ToTime("2026/08/17", WithTimeLayout("2006/01/02"))
	testx.NoError(t, err)
	testx.Equal(t, tm.Year(), 2026)

	// 指定格式时仅尝试该格式，不匹配则失败。
	_, err = ToTime("2026-08-17", WithTimeLayout("2006/01/02"))
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrUnknownFormat)
}

func TestToTimeUnknownFormat(t *testing.T) {
	_, err := ToTime("not-a-time")
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrUnknownFormat)
}

func TestToTimeTimezoneOffsetPreserved(t *testing.T) {
	tm, err := ToTime("2026-08-17T10:00:00+08:00")
	testx.NoError(t, err)
	testx.Equal(t, tm.Format(time.RFC3339), "2026-08-17T10:00:00+08:00")
}

func TestToTimeDefaultLocation(t *testing.T) {
	// 无时区字符串默认 time.Local。
	tm, err := ToTime("2026-08-17 10:00:00")
	testx.NoError(t, err)
	testx.Equal(t, tm.Location(), time.Local)

	// WithDefaultLocation 指定 UTC。
	tm, err = ToTime("2026-08-17 10:00:00", WithDefaultLocation(time.UTC))
	testx.NoError(t, err)
	testx.Equal(t, tm.Location(), time.UTC)
}

func TestToTimeTimestampPrecision(t *testing.T) {
	// 秒级：< 1e11。
	sec := int64(1755484800)
	tm, err := ToTime(sec)
	testx.NoError(t, err)
	testx.Equal(t, tm.Unix(), sec)

	// 毫秒级：1e11 ~ 1e14。
	ms := int64(1755484800000)
	tm, err = ToTime(ms)
	testx.NoError(t, err)
	testx.Equal(t, tm.UnixMilli(), ms)

	// 微秒级：1e14 ~ 1e17。
	us := int64(1755484800000000)
	tm, err = ToTime(us)
	testx.NoError(t, err)
	testx.Equal(t, tm.UnixMicro(), us)

	// 纳秒级：≥ 1e17。
	ns := int64(1755484800000000000)
	tm, err = ToTime(ns)
	testx.NoError(t, err)
	testx.Equal(t, tm.UnixNano(), ns)

	// 负时间戳（1970 前）。
	tm, err = ToTime(int64(-1000))
	testx.NoError(t, err)
	testx.Equal(t, tm.Unix(), int64(-1000))
}

func TestToTimeWithTimeUnit(t *testing.T) {
	// 强制指定秒级，跳过推断（大数按秒处理会超出 time 范围，验证强制单位生效）。
	ts := int64(1755484800)
	tm, err := ToTime(ts, WithTimeUnit(UnitSecond))
	testx.NoError(t, err)
	testx.Equal(t, tm.Unix(), ts)
}

func TestToTimeUintSource(t *testing.T) {
	tm, err := ToTime(uint64(1755484800))
	testx.NoError(t, err)
	testx.Equal(t, tm.Unix(), int64(1755484800))
}

func TestToTimeUnsupportedSource(t *testing.T) {
	_, err := ToTime(3.14)
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrConversionFailed)

	_, err = ToTime(true)
	testx.Error(t, err)
}

func TestTimeToString(t *testing.T) {
	tm := time.Date(2026, 8, 17, 10, 0, 0, 0, time.UTC)
	s := TimeToString(tm, time.RFC3339)
	testx.Equal(t, s, "2026-08-17T10:00:00Z")

	// To[string] 默认 RFC3339。
	s, err := To[string](tm)
	testx.NoError(t, err)
	testx.Equal(t, s, "2026-08-17T10:00:00Z")

	// WithTimeLayout 指定格式。
	s, err = To[string](tm, WithTimeLayout("2006-01-02"))
	testx.NoError(t, err)
	testx.Equal(t, s, "2026-08-17")
}

func TestTimeToTimestampConversions(t *testing.T) {
	tm := time.Date(2026, 8, 17, 10, 0, 0, 0, time.UTC)

	// time → int64 默认秒级。
	ts, err := To[int64](tm)
	testx.NoError(t, err)
	testx.Equal(t, ts, tm.Unix())

	// 指定精度。
	ms, err := To[int64](tm, WithTimeUnit(UnitMillisecond))
	testx.NoError(t, err)
	testx.Equal(t, ms, tm.UnixMilli())

	ns, err := To[int64](tm, WithTimeUnit(UnitNanosecond))
	testx.NoError(t, err)
	testx.Equal(t, ns, tm.UnixNano())

	// TimestampToTime / TimeToTimestamp 便捷 API。
	testx.Equal(t, TimestampToTime(ts, UnitSecond).Unix(), ts)
	testx.Equal(t, TimeToTimestamp(tm, UnitSecond), ts)
}

func TestTimeToUintTimestamp(t *testing.T) {
	tm := time.Date(2026, 8, 17, 10, 0, 0, 0, time.UTC)
	ts, err := To[uint64](tm)
	testx.NoError(t, err)
	testx.Equal(t, ts, uint64(tm.Unix()))
}

func TestTimezoneHelpers(t *testing.T) {
	tm := time.Date(2026, 8, 17, 10, 0, 0, 0, time.FixedZone("CST", 8*3600))

	utc := ToUTC(tm)
	testx.Equal(t, utc.Location(), time.UTC)
	testx.Equal(t, utc.Hour(), 2)

	loc := ToLocal(tm)
	testx.Equal(t, loc.Location(), time.Local)

	ny, _ := time.LoadLocation("America/New_York")
	loc2 := ToLocation(tm, ny)
	testx.Equal(t, loc2.Location(), ny)
}

func TestToDuration(t *testing.T) {
	tests := []struct {
		name string
		src  any
		want time.Duration
	}{
		{"字符串", "1h30m", 90 * time.Minute},
		{"毫秒字符串", "500ms", 500 * time.Millisecond},
		{"整数纳秒", int64(1000000000), time.Second},
		{"Duration 原样", time.Minute, time.Minute},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := ToDuration(tt.src)
			testx.NoError(t, err)
			testx.Equal(t, d, tt.want)
		})
	}
}

func TestToDurationErrors(t *testing.T) {
	_, err := ToDuration("not-duration")
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrConversionFailed)

	_, err = ToDuration(true)
	testx.Error(t, err)
}

func TestDurationToString(t *testing.T) {
	s, err := To[string](90 * time.Minute)
	testx.NoError(t, err)
	testx.Equal(t, s, "1h30m0s")
}

func TestToTimeErrorContext(t *testing.T) {
	_, err := ToTime("bad", WithDefaultLocation(time.UTC))
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrUnknownFormat)
}

func TestInferTimeUnit(t *testing.T) {
	testx.Equal(t, inferTimeUnit(99999999999), UnitSecond)
	testx.Equal(t, inferTimeUnit(100000000000), UnitMillisecond)
	testx.Equal(t, inferTimeUnit(99999999999999), UnitMillisecond)
	testx.Equal(t, inferTimeUnit(100000000000000), UnitMicrosecond)
	testx.Equal(t, inferTimeUnit(99999999999999999), UnitMicrosecond)
	testx.Equal(t, inferTimeUnit(100000000000000000), UnitNanosecond)
}

func TestToInt64Overflow(t *testing.T) {
	// uint64 大值转 int64 溢出。
	_, err := ToTime(uint64(math.MaxInt64) + 100)
	testx.Error(t, err)
}
