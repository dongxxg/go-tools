

-- 添加 lib 目录到 package.path（假设 str.lua 在 lib/str/ 目录下）
package.path = package.path .. ";./lib/?.lua;./lib/?/init.lua"

-- 引入模块
local str = require "str.str"

-- 使用 split 函数
local input = "apple|banana|cherry"
local result = str.split(input, "|")

-- 打印结果
for i, v in ipairs(result) do
    print(i, v)
end