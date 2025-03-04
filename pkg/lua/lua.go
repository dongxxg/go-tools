package main

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
)

func main() {
	L := lua.NewState()
	defer L.Close()

	err := L.DoString(regexCmd)
	if err != nil {
		fmt.Println(err)
	}
}

var regexCmd = `
function wildcard_to_regex(wildcard)
    -- 转义所有特殊字符，除了通配符
    wildcard = wildcard:gsub("([%.%+%-%^%$%(%)%[%]%%%$])", "%%%1")
    -- 替换通配符*为.*
    wildcard = wildcard:gsub("%*", ".*")
    -- 替换通配符?为.
    wildcard = wildcard:gsub("%?", ".")
    return "^" .. wildcard .. "$"
end

function path_matches(path, pattern)
    local regex = wildcard_to_regex(pattern)
    print(regex)
    return path:match(regex) ~= nil
end

-- 示例
local path1 = "/hello3/world3"
local pattern1 = "/hello3/world*" -- 匹配example/path/to/中file之后有一个或多个字符，然后是.txt

local path2 = "example/path/to/filea.txt"
local pattern2 = "example/*/file?.txt" -- 匹配example/任意字符/file加一个字符.txt

if path_matches(path1, pattern1) then
    print("Path 1 Match!")
else
    print("Path 1 No match.")
end

if path_matches(path2, pattern2) then
    print("Path 2 Match!")
else
    print("Path 2 No match.")
end
`

var appIdCmd = `
local app1 = {
  appId = "1",
  minute = 3,
}

local fields = {}
fields[1] = app1

local keys = {}

for _, field in ipairs(fields or {}) do
  print("appId: " .. field.appId .. ", minute==" .. field.minute)
end
`
