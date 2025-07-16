package util

import (
	"fmt"
	"testing"
)

func TestGetOssRelativePath(t *testing.T) {
	fmt.Println(GetOssRelativePath("https://file.plaso.cn/dev-plaso/liveclass/1866/621wy_1624283684033_devconf0/1_1624283919604.jpeg"))
	fmt.Println(GetOssRelativePath("https://file.plaso.cn/dev-plaso/liveclass/1866/621wy_1624283684033_devconf0/"))
	fmt.Println(GetOssRelativePath("https://file.plaso.c1n1/dev-plaso/liveclass/1866/621wy_1624283684033_devconf0/"))
}

func TestGetRelativePath(t *testing.T) {
	fmt.Println(GetRelativePath("https://file.plaso.cn/dev-plaso/liveclass/1866/621wy_1624283684033_devconf0/1_1624283919604.jpeg"))
	fmt.Println(GetRelativePath("https://file.plaso.cn/dev-plaso/liveclass/1866/621wy_1624283684033_devconf0/"))
	fmt.Println(GetRelativePath("https://file.plaso.c1n1/dev-plaso/liveclass/1866/621wy_1624283684033_devconf0/"))
	fmt.Println(GetRelativePath("1_1624283919604"))
	fmt.Println(GetRelativePath("1_1624283919604.jpg"))
	fmt.Println(GetRelativePath([]int{1, 3, 4}))
}
