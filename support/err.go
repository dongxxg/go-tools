package support

// 统一错误处理方式，停止程序逻辑，进而执行延迟函数。
// 避免每个函数都要处理err，出现大量的if...else...
func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}
