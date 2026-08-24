

-- 添加 lib 目录到 package.path（假设 str.lua 在 lib/str/ 目录下）
package.path = package.path .. ";./lib/?.lua;./lib/?/init.lua"

local json = require("dkjson")

-- 引入模块
local str = require "str.str"

-- 使用 split 函数
--local input = "apple|banana|cherry"
--local result = str.split(input, "|")

local result = {}

--if err then
--    print(string.format("解析 JSON 失败: %s", err))
--else
--    table.insert(result, obj)
--end

-- 打印结果
--for i, v in ipairs(result) do
--    for k, val in pairs(v) do
--        print(k, val == nil and "nil" or val)
--    end
--end


local s = "{\"devinfo\":3, \"3444\": \"null\", \"34\": {1: null, 2: 34}}"
local obj, pos, err = json.decode(s)
if err then
    print("解析失败:", err)
else
    for k, v in pairs(obj) do
        print(k, v == nil and "nil" or v)
    end
end