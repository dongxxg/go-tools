package support

import (
	"fmt"
	"testing"
)

func TestDefault(t *testing.T) {
	var old1 *int64
	var new1 *int64
	var tmp1 int64 = 0
	old1 = &tmp1

	var t1 int64 = 1
	fmt.Println("IsNilOrDefaultValue(int64(0))", IsNilOrDefaultValue(*old1))
	fmt.Println("IsNilOrDefaultValue(int64(1))", IsNilOrDefaultValue(t1))
	fmt.Println("IsNilOrDefaultValue(*int64(0))", IsNilOrDefaultValue(old1))
	fmt.Println("IsNilOrDefaultValue(*int64(1))", IsNilOrDefaultValue(&t1))
	fmt.Println("IsNilOrDefaultValue(*int64(nil))", IsNilOrDefaultValue(new1))
	fmt.Println("CompareValue(*int64(0), *int64(nil))", CompareValue(old1, new1))
	fmt.Println("CompareValue(*int64(nil), *int64(0))", CompareValue(new1, old1))

	fmt.Println()

	var old2 *string
	var new2 *string
	var tmp2 string = "1"
	old2 = &tmp2

	var t2 string = ""
	fmt.Println("IsNilOrDefaultValue(string(1))", IsNilOrDefaultValue(*old2))
	fmt.Println("IsNilOrDefaultValue(string(\"\"))", IsNilOrDefaultValue(t2))
	fmt.Println("IsNilOrDefaultValue(*string(1))", IsNilOrDefaultValue(old2))
	fmt.Println("IsNilOrDefaultValue(*string(\"\"))", IsNilOrDefaultValue(&t2))
	fmt.Println("IsNilOrDefaultValue(*string(nil))", IsNilOrDefaultValue(new2))
	fmt.Println("CompareValue(*string(1), *string(nil))", CompareValue(old2, new2))
	fmt.Println("CompareValue(*string(nil), *string(1))", CompareValue(new2, old2))
	fmt.Println()

	var old3 *bool
	var new3 *bool
	var tmp3 bool = true
	old3 = &tmp3

	var t3 bool = false
	fmt.Println("IsNilOrDefaultValue(bool(true))", IsNilOrDefaultValue(*old3))
	fmt.Println("IsNilOrDefaultValue(bool(false))", IsNilOrDefaultValue(t3))
	fmt.Println("IsNilOrDefaultValue(*bool(true))", IsNilOrDefaultValue(old3))
	fmt.Println("IsNilOrDefaultValue(*bool(false))", IsNilOrDefaultValue(&t3))
	fmt.Println("IsNilOrDefaultValue(*bool(nil))", IsNilOrDefaultValue(new3))
	fmt.Println("CompareValue(*bool(true), *bool(nil))", CompareValue(old3, new3))
	fmt.Println("CompareValue(*bool(nil), *bool(true))", CompareValue(new3, old3))
	fmt.Println()

	var old4 *float64
	var new4 *float64
	var tmp4 float64 = 2.323
	old4 = &tmp4

	var t4 float64 = 0
	fmt.Println("IsNilOrDefaultValue(float64(2.323))", IsNilOrDefaultValue(*old4))
	fmt.Println("IsNilOrDefaultValue(float64(0))", IsNilOrDefaultValue(t4))
	fmt.Println("IsNilOrDefaultValue(*float64(2.323))", IsNilOrDefaultValue(old4))
	fmt.Println("IsNilOrDefaultValue(*float64(0))", IsNilOrDefaultValue(&t4))
	fmt.Println("IsNilOrDefaultValue(*float64(nil))", IsNilOrDefaultValue(new4))
	fmt.Println("CompareValue(*float64(2.323), *float64(nil))", CompareValue(old4, new4))
	fmt.Println("CompareValue(*float64(nil), *float64(2.323))", CompareValue(new4, old4))
	fmt.Println()

	var old5 *[]byte
	var new5 *[]byte
	var tmp5 []byte = []byte("1")
	old5 = &tmp5

	var t5 []byte
	fmt.Println("IsNilOrDefaultValue([]byte(1))", IsNilOrDefaultValue(*old5))
	fmt.Println("IsNilOrDefaultValue([]byte", IsNilOrDefaultValue(t5))
	fmt.Println("IsNilOrDefaultValue(*[]byte(1))", IsNilOrDefaultValue(old5))
	fmt.Println("IsNilOrDefaultValue(*[]byte", IsNilOrDefaultValue(&t5))
	fmt.Println("IsNilOrDefaultValue(*[]byte(nil))", IsNilOrDefaultValue(new5))
	fmt.Println("CompareValue(*[]byte(1), *[]byte(nil))", CompareValue(old5, new5))
	fmt.Println("CompareValue(*[]byte(nil), *[]byte(1))", CompareValue(new5, old5))
	fmt.Println()

}
