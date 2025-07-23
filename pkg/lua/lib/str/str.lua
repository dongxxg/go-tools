

local M = {}

-- 自定义字符串分割函数
function M.split(str, delimiter)
    local result = {}
    local from = 1
    local delim_from, delim_to

    while true do
        delim_from, delim_to = string.find(str, delimiter, from)
        if not delim_from then
            break
        end
        table.insert(result, string.sub(str, from, delim_from - 1))
        from = delim_to + 1
    end
    table.insert(result, string.sub(str, from))
    return result
end

return M