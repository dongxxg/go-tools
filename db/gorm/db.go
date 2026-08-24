package db

import (
	"bytes"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"text/template"
)

const (
	INT       = "int"
	VARCHAR   = "varchar"
	CHAR      = "char"
	TIMESTAMP = "timestamp"
	DATE      = "date"
	TEXT      = "text"
	DECIMAL   = "decimal"
)

func GetColumnType(t string) string {

	switch t {
	case "string":
		return VARCHAR
	case "int":
	case "int8":
	case "int16":
	case "int32":
	case "int64":
		return INT
	case "float32":
	case "float64":
		return DECIMAL
	case "time":
		return TIMESTAMP
	default:
		return CHAR
	}
	return CHAR
}

type Base struct {
	id int8 `gorm:"column:id" json:"id"`
}

type Column struct {
	Name         string `json:"name"`
	TypeName     string `json:"typeName"`
	Length       int    `json:"length"`
	IsPrimaryKey bool   `json:"isPrimaryKey"`
}

func openDb(name string) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s&parseTime=%s&loc=%s",
		"root",
		"A123456z",
		"localhost",
		name,
		"utf8mb4",
		"true",
		"Local")

	config := &mysql.Config{
		DSN:                       dsn,
		SkipInitializeWithVersion: true,
	}
	db, _ := gorm.Open(mysql.Dialector{
		config,
	})
	return db
}

func GetData(sourceDbName string, runSql string) []map[string]interface{} {
	db := openDb(sourceDbName)
	result := new([]map[string]interface{})
	tx := db.Raw(runSql)
	tx.Scan(result)
	return *result
}

func ShowTable(table string) string {
	db := openDb("from_db")
	result := new(string)  //header
	result1 := new(string) //content
	tx := db.Raw(fmt.Sprintf("show create table %s;", table))
	r, _ := tx.Rows()
	defer r.Close()
	if r.Next() {
		e := r.Scan(result, result1)
		if e != nil {
			fmt.Println(e)
		}
	}
	return *result1
}

func InsertData(data []map[string]interface{}, destDbName string, destTableName string) {
	db := openDb(destDbName)
	tx := db.Exec(fmt.Sprintf("select * from %s limit 1"), destTableName)
	if tx.Error != nil {
		createSql := ShowTable(destTableName)
		exec := db.Exec(createSql)
		if exec.Error != nil {
			fmt.Println("create table faile and error msg", exec.Error)
			return
		}
	}
	//准备插入sql
	insertSql := "insert into %s (%s) values %s"
	var columns []string = make([]string, 0)
	var allValues [][]interface{} = make([][]interface{}, 10)
	for index, res := range data {
		values := make([]interface{}, 0)
		if index == 0 {
			for key, _ := range res {
				columns = append(columns, key)
				//values = append(values, value)
			}
		}
		//处理map乱序的问题
		values = sortKey(columns, res)
		allValues[index] = values
	}
	columnStr := strings.Join(columns, ",")
	t := template.Must(template.New("values").Funcs(template.FuncMap{"lastRow": DealLastRow}).
		Funcs(template.FuncMap{"lastCol": DealLastData}).Funcs(template.FuncMap{
		"dealData": DealData,
	}).Parse(valueSql))
	buf := &bytes.Buffer{}
	t.Execute(buf, struct {
		AllValues [][]interface{}
	}{allValues})
	sql := fmt.Sprintf(insertSql, destTableName, columnStr, strings.TrimSpace(buf.String()))
	fmt.Println(sql)
	tx = db.Begin()
	reDb := tx.Exec(sql)
	if reDb.Error != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}

}

func sortKey(keys []string, data map[string]interface{}) (values []interface{}) {
	for _, key := range keys {
		if data[key] == nil {
			data[key] = "NULL"
		}
		values = append(values, data[key])
	}
	return
}
