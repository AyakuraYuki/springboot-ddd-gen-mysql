package main

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"strings"
)

const (
	sqlShowTableStatus = "SHOW TABLE STATUS LIKE '%s'"
	sqlShowFullColumns = "SHOW FULL COLUMNS FROM %s"
)

// connectToDB inits mDB to connect to the Database.
func connectToDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, host, port, schemaName)
	fmt.Printf("dsn: %s\n\n", dsn)
	if mDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{}); err != nil {
		panic(err)
	}
}

// readTablePrototype returns table status and columns by fetching MySQL.
func readTablePrototype() (*TableStatus, []ColumnsStatement) {
	tableStatus := &TableStatus{}
	if err = mDB.Raw(fmt.Sprintf(sqlShowTableStatus, tableName)).First(&tableStatus).Error; err != nil {
		panic(err)
	}
	columnsStatements := make([]ColumnsStatement, 0)
	if err = mDB.Raw(fmt.Sprintf(sqlShowFullColumns, tableName)).Find(&columnsStatements).Error; err != nil {
		panic(err)
	}
	return tableStatus, columnsStatements
}

// parseJavaFields returns Java fields by analysing table info
func parseJavaFields(columns []ColumnsStatement) (javaFields []JavaField) {
	javaFields = make([]JavaField, 0)
	for _, v := range columns {
		javaType, packageName := getStructType(v)
		f := JavaField{
			JavaType:    javaType,
			Field:       camelCase(v.Field),
			Comment:     v.Comment,
			PackageName: packageName,
			IsPri:       strings.ToUpper(v.Key) == "PRI",
		}
		javaFields = append(javaFields, f)
	}
	return
}

// getStructType returns Java type by DB type
func getStructType(s ColumnsStatement) (string, string) {
	types := strings.ToLower(s.Type)
	if strings.Contains(types, "varchar") ||
		strings.Contains(types, "char") ||
		strings.Contains(types, "text") ||
		strings.Contains(types, "json") {
		return "String", ""
	}
	if strings.Contains(types, "bigint") {
		return "Long", ""
	}
	if strings.Contains(types, "tinyint") ||
		strings.Contains(types, "integer") ||
		strings.Contains(types, "int") {
		return "Integer", ""
	}
	if strings.Contains(types, "decimal") ||
		strings.Contains(types, "numeric") ||
		strings.Contains(types, "double") ||
		strings.Contains(types, "float") ||
		strings.Contains(types, "real") {
		return "BigDecimal", "java.math.BigDecimal"
	}
	if strings.Contains(types, "bit") ||
		strings.Contains(types, "boolean") {
		return "Boolean", ""
	}
	if strings.Contains(types, "binary") {
		return "byte[]", ""
	}
	if strings.Contains(types, "date") ||
		strings.Contains(types, "time") {
		return "LocalDateTime", "java.time.LocalDateTime"
	}
	return "String", ""
}

// tryRemoveTablePrefix will remove table prefix by the following rules:
// 1. tb_table -> table;
// 2. t_table -> table;
// 3. r_relation -> relation.
func tryRemoveTablePrefix(s string) string {
	if len(s) == 0 {
		return s
	}
	if strings.HasPrefix(s, "tb_") {
		return strings.TrimPrefix(s, "tb_")
	}
	if strings.HasPrefix(s, "t_") {
		return strings.TrimPrefix(s, "t_")
	}
	if strings.HasPrefix(s, "r_") {
		return strings.TrimPrefix(s, "r_")
	}
	return s
}

// isASCIILower returns true if character is in ASCII range 'a - z', false if out of this range.
func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

// camelCase makes underscore string to camel case string.
func camelCase(s string) string {
	var b []byte
	var wasUnderscore bool
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '_' {
			if wasUnderscore && isASCIILower(c) {
				c -= 'a' - 'A'
			}
			b = append(b, c)
		}
		wasUnderscore = c == '_'
	}
	return string(b)
}

// firstUpCase makes the first character to upper case.
func firstUpCase(str string) string {
	if len(str) == 0 {
		return str
	}
	if !isASCIILower(str[0]) {
		return str
	}
	c := str[0]
	c -= 'a' - 'A'
	b := []byte(str)
	b[0] = c
	return string(b)
}

func writeFile(path, filename, codes string) {
	if _, err = os.Stat(path); os.IsNotExist(err) {
		if err = os.MkdirAll(path, os.ModePerm); err != nil {
			panic(err)
		}
	}
	_ = os.WriteFile(filename, []byte(codes), 0666)
}
