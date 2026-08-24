package support

import "reflect"

// IsNilOrDefaultValue 判断是否是默认值, 支持基本类型及其指针
// 如果是指针, 如果是nil直接返回true, 不为nil则判断其指向值是否是默认值
func IsNilOrDefaultValue(value any) bool {
	// 获取该值的类型
	v := reflect.ValueOf(value)

	// 如果是指针且不为nil, 获取其指向的值
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return true
		}
		v = v.Elem()
	}
	// 获取零值 (默认值)
	zeroValue := reflect.Zero(v.Type()).Interface()

	// 如果值等于零值，说明它是默认值
	return reflect.DeepEqual(v.Interface(), zeroValue)
}

// CompareValue 比较两个值，如果新值是默认值，则返回旧值，否则返回新值, 支持基本类型及其指针
func CompareValue[T any](old, new T) T {
	if IsNilOrDefaultValue(new) {
		return old
	}
	return new
}
