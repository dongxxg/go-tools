package main

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
)

func main() {
	L := lua.NewState()
	defer L.Close()
	cmd := `
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
	err := L.DoString(cmd)
	if err != nil {
		fmt.Println(err)
	}
}
