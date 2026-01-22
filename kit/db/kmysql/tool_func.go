package kmysql

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"strings"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/util"
)

// QueryAll 查询返回所有字段
func QueryAll(ctx kctx.Ctx, db *sql.DB, query string, mapKey string) (map[string][]interface{}, error) {
	mapKey = strings.ToLower(mapKey)
	rows, err := db.Query(query)
	if err != nil {
		return nil, errors.Wrapf(err, "db query")
	}
	defer rows.Close()

	names, err := rows.Columns()
	if err != nil {
		return nil, errors.Wrapf(err, "asdfasf")
	}
	out := make(map[string][]interface{})

	for rows.Next() {
		res3 := make([]interface{}, len(names), len(names))
		resPtr := make([]interface{}, len(names), len(names))
		for j := 0; j < len(res3); j++ {
			resPtr[j] = &res3[j]
			//resPtr = append(resPtr, &res3[j])
		}
		if err := rows.Scan(resPtr...); err != nil {
			fmt.Println(err)
			continue
		}
		temp := make([]interface{}, 0, len(names))

		var mapKeyVal string
		for i := 0; i < len(names); i++ {
			if mapKeyVal == "" {
				n := strings.ToLower(names[i])
				if n == mapKey {
					mapKeyVal = util.ConvertToString(res3[i])
				}
			}

			if strings.Contains(names[i], "time") {
				t, err := util.ParseToTime(util.ConvertToString(res3[i]))
				if err != nil {
					temp = append(temp, "null")
					continue
				}
				temp = append(temp, t.Format("2006-01-02 15:04:05"))
				continue
			}
			ss := util.ConvertToString(res3[i])
			if strings.Contains(ss, "'") {
				ss = strings.ReplaceAll(ss, "'", "\\'")
				temp = append(temp, ss)
			} else {
				temp = append(temp, res3[i])
			}
		}
		if len(temp) == 0 || mapKeyVal == "" {
			continue
		}
		out[mapKeyVal] = temp
	}
	return out, nil
}
