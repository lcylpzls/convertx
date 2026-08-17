package core

import (
	"reflect"
	"sync"
)

// FieldMeta 描述结构体的单个字段元信息。
type FieldMeta struct {
	// Name 是字段名。
	Name string
	// Index 是字段的嵌套索引（匿名字段展开后，用于 FieldByIndex）。
	Index []int
	// Type 是字段类型。
	Type reflect.Type
	// TagName 是 tag 指定的映射名（tag 值为 "-" 时为空且 Omit 为 true）。
	TagName string
	// Omit 表示 tag 为 "-"，转换时忽略该字段。
	Omit bool
}

// metaCacheKey 是结构体元信息缓存键：类型 + tag 名。
// tagName 可通过 WithTagName 选项变化，必须纳入键。
type metaCacheKey struct {
	t       reflect.Type
	tagName string
}

// metaCache 缓存结构体元信息，避免每次转换重复反射解析。
// 返回的 []FieldMeta 为共享缓存实例，调用方只读使用、不得修改。
var metaCache sync.Map // map[metaCacheKey][]FieldMeta

// metaMaxDepth 是匿名字段展开的最大深度：超限时嵌入字段按普通字段处理，
// 防御异常深层的类型嵌套（正常代码远低于该值）。
const metaMaxDepth = 1000

// parseStructMeta 反射解析结构体的字段元信息，带缓存。
// 匿名（嵌入）字段自动展开并提升；指针型嵌入字段按指向的结构体展开。
func parseStructMeta(t reflect.Type, tagName string) []FieldMeta {
	key := metaCacheKey{t: t, tagName: tagName}
	if cached, ok := metaCache.Load(key); ok {
		return cached.([]FieldMeta)
	}
	fields := parseStructMetaUncached(t, tagName)
	actual, _ := metaCache.LoadOrStore(key, fields)
	return actual.([]FieldMeta)
}

// parseStructMetaUncached 执行无缓存的反射解析。
// visited 记录当前递归路径上的类型，防止匿名自嵌入/互嵌指针导致无限递归；
// 兄弟字段嵌入同一类型不受影响（返回前从 visited 移除）。
func parseStructMetaUncached(t reflect.Type, tagName string) []FieldMeta {
	var fields []FieldMeta
	visited := make(map[reflect.Type]bool)
	var visit func(rt reflect.Type, prefix []int)
	visit = func(rt reflect.Type, prefix []int) {
		if visited[rt] {
			return
		}
		visited[rt] = true
		defer delete(visited, rt)
		for i := 0; i < rt.NumField(); i++ {
			sf := rt.Field(i)
			// 匿名字段：递归展开嵌入结构体。
			if sf.Anonymous {
				ft := sf.Type
				if ft.Kind() == reflect.Ptr {
					ft = ft.Elem()
				}
				if ft.Kind() == reflect.Struct && len(prefix) < metaMaxDepth {
					idx := append(append([]int{}, prefix...), i)
					visit(ft, idx)
					continue
				}
			}
			tag := sf.Tag.Get(tagName)
			// 未导出字段不参与转换（避免反射读取 panic）。
			if sf.PkgPath != "" {
				continue
			}
			fields = append(fields, FieldMeta{
				Name:    sf.Name,
				Index:   append(append([]int{}, prefix...), i),
				Type:    sf.Type,
				TagName: tag,
				Omit:    tag == "-",
			})
		}
	}
	visit(t, nil)
	return fields
}

// joinFieldPath 拼接嵌套字段路径（如 "User" + "Address" → "User.Address"）。
func joinFieldPath(base, name string) string {
	if base == "" {
		return name
	}
	return base + "." + name
}
