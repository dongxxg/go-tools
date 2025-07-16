package util

func SliceIntEquals(a []int, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, va := range a {
		if va != b[i] {
			return false
		}
	}
	return true
}

func SliceIntHasPrefix(a []int, b []int) bool {
	if len(a) < len(b) {
		return false
	}
	for i, vb := range b {
		if a[i] != vb {
			return false
		}
	}
	return true
}

func SliceStringIndexOf(stringSlice []string, e string) int {
	for i, str := range stringSlice {
		if str == e {
			return i
		}
	}
	return -1
}

// SetDeleteString 字符串集合删除
// @setString 元素唯一
func SetDeleteString(setString []string, e string) []string {
	if i := SliceStringIndexOf(setString, e); i >= 0 {
		setString = append(setString[:i], setString[i+1:]...)
	}
	return setString
}

// SetAddString 字符串集合添加
// @setString 元素唯一
func SetAddString(setString []string, e string) []string {
	if i := SliceStringIndexOf(setString, e); i < 0 {
		setString = append(setString, e)
	}
	return setString
}

// 包含
func StrContains(strings []string, s string) bool {
	return SliceStringIndexOf(strings, s) > -1
}
