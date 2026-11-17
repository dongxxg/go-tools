package db

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetCoalesceAssignmentByDialect 配合 gorm 的 clause.OnConflict 使用, 在ON DUPLICATE KEY UPDATE 中使用COALESCE函数, 并抹平不同数据库之间差异
// 1. 在冲突更新的情况下, 如果新值为null则使用原值, 2. 新值不为null或记录不存在的情况下, 使用新值
// 防止有些字段本不是null但冲突更新的时候被设置成了null, 比如:
// 有个记录A在数据库中存在且A.a不为null, 但是新增的时候不知道该数据是否存在, 所以默认被当成一条新纪录新增(新增的结构体中A.a初始值是null),
// 新增的时候由于冲突, 走了冲突更新的逻辑, 错误地把A.a更新了null值进去(表中A.a的值可能不是null)
// 所以这里使用 COALESCE 函数实现, 不同数据库之间使用不同表达式实现
// 这里只针对特殊字段使用, 如果某些场景需要更新字段为null值, 请勿使用该函数, 请直接使用 clause.AssignmentColumns, clause.Assignments 等函数
// 参数:
// 1. dialect: 数据库类型, 目前支持 mysql, 达梦, 其他类型返回空
// 2. tableName: 表名, 用于达梦数据库的表达式中
// 3. columns: 需要更新的字段, 默认使用COALESCE函数, 冲突情况下如果字段为null则使用原值, 如果字段不为null则使用新值
func GetCoalesceAssignmentByDialect(dialect, tableName string, columns ...string) clause.Set {

	if len(columns) == 0 || len(tableName) == 0 {
		return clause.Set{}
	}
	assignmentMap := make(map[string]any, len(columns))

	switch dialect {
	case "mysql":
		for _, column := range columns {
			assignmentMap[column] = gorm.Expr(
				fmt.Sprintf("COALESCE(VALUES(%s), %s)", column, column),
			)
		}
	case "dameng":
		for _, column := range columns {
			assignmentMap[column] = gorm.Expr(
				fmt.Sprintf("COALESCE(excluded.%s, %s.%s)", column, tableName, column),
			)
		}
	default:
		return clause.Set{}
	}
	return clause.Assignments(assignmentMap)
}
