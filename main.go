// MIT License
//
// # Copyright (c) 2022 Ayakura Yuki
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package main

// Author: AyakuraYuki
// Github: https://github.com/AyakuraYuki/springboot-ddd-gen-mysql
// Great salute to https://github.com/wthsjy/wt-gen-mysql-proto

import (
	"flag"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	sqlShowTableStatus = "SHOW TABLE STATUS LIKE '%s'"
	sqlShowFullColumns = "SHOW FULL COLUMNS FROM %s"
)

var (
	err        error
	mDB        *gorm.DB
	host       string
	port       int
	username   string
	password   string
	schemaName string
	tableName  string
	domainName string
)

func main() {
	flag.StringVar(&username, "u", "root", "用户名")
	flag.StringVar(&password, "p", "root", "密码")
	flag.StringVar(&host, "h", "localhost", "主机名，默认 localhost")
	flag.IntVar(&port, "P", 3306, "端口号，默认 3306")
	flag.StringVar(&schemaName, "d", "db_local", "数据库名")
	flag.StringVar(&tableName, "t", "tb_user", "表名")
	flag.StringVar(&domainName, "D", "user", "领域名")
	flag.Parse()

	// fetch table info
	connectToDB()
	tableStatus, ddlms := readTablePrototype()

	// parse class members
	javaFields := parseJavaFields(ddlms)

	genPO(tableStatus, javaFields)
	genMapper(tableStatus)
	genRepository(tableStatus)
}

type TableStatus struct {
	Name      string
	Rows      uint64
	Collation string
	Comment   string
}

type DDLM struct {
	Field   string
	Type    string
	Null    string
	Key     string
	Comment string
	Default string
	Extra   string
}

type JavaField struct {
	JavaType    string
	Field       string
	Comment     string
	PackageName string
	IsPri       bool
}

// parseJavaFields returns Java fields by analysing table info
func parseJavaFields(ddlms []DDLM) (javaFields []JavaField) {
	javaFields = make([]JavaField, 0)
	for _, v := range ddlms {
		javaType, packageName := getStructType(v)
		f := JavaField{
			JavaType:    javaType,
			Field:       v.Field,
			Comment:     v.Comment,
			PackageName: packageName,
			IsPri:       strings.ToUpper(v.Key) == "PRI",
		}
		javaFields = append(javaFields, f)
	}
	return
}

func genJavadoc(className string, tableStatus *TableStatus) string {
	return fmt.Sprintf(`/**
 * %s - %s
 *
 * <p>table_name: %s</p>
 *
 * @author ddd-gen-mysql
 * @date %s
 */`, className, tableStatus.Comment, tableStatus.Name, time.Now().String())
}

func genPO(tableStatus *TableStatus, javaFields []JavaField) {
	entityName := tryRemoveTablePrefix(tableStatus.Name)
	className := fmt.Sprintf("%sPo", firstUpCase(camelCase(entityName)))

	importCodes, fieldCodes := "", ""
	for _, v := range javaFields {
		if v.PackageName != "" && !strings.Contains(importCodes, v.PackageName) {
			importCodes += fmt.Sprintf("import %s;\n", v.PackageName)
		}
		if v.Field == "id" || v.Field == "ctime" || v.Field == "mtime" {
			continue
		}
		fieldCodes += fmt.Sprintf("  private %s %s; // %s\n", v.JavaType, v.Field, v.Comment)
	}
	importCodes = strings.TrimSuffix(importCodes, "\n")
	fieldCodes = strings.TrimSuffix(fieldCodes, "\n")
	codes := fmt.Sprintf(`package com.mahuafm.phoenix.%s.infrastructure.persistence.po;

import com.baomidou.mybatisplus.annotation.TableName;
import com.mahuafm.phoenix.util.infrastructure.persistence.po.base.BaseAutoIdPo;
%s
import lombok.Data;
import lombok.EqualsAndHashCode;

%s
@Data
@EqualsAndHashCode(callSuper = true)
@TableName("%s")
public class %s extends BaseAutoIdPo {

%s

}
`, domainName, importCodes, genJavadoc(className, tableStatus), tableStatus.Name, className, fieldCodes)

	path := fmt.Sprintf("./%s", filepath.Join("ddd-code-gen", domainName, "infrastructure", "persistence", "po"))
	filename := fmt.Sprintf("./%s/%s.java", path, className)
	if _, err = os.Stat(path); os.IsNotExist(err) {
		if err = os.MkdirAll(path, os.ModePerm); err != nil {
			panic(err)
		}
	}
	_ = os.WriteFile(filename, []byte(codes), 0666)
	fmt.Printf("%s: %s\n", className, filename)
}

func genMapper(tableStatus *TableStatus) {
	entityName := tryRemoveTablePrefix(tableStatus.Name)
	className := fmt.Sprintf("%sMapper", firstUpCase(camelCase(entityName)))
	poClassName := fmt.Sprintf("%sPo", firstUpCase(camelCase(entityName)))

	codes := fmt.Sprintf(`package com.mahuafm.phoenix.%s.infrastructure.persistence.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mahuafm.phoenix.points.infrastructure.persistence.po.PointsConfigPo;

%s
public interface %s extends BaseMapper<%s> {}
`, domainName, genJavadoc(className, tableStatus), className, poClassName)

	path := fmt.Sprintf("./%s", filepath.Join("ddd-code-gen", domainName, "infrastructure", "persistence", "mapper"))
	filename := fmt.Sprintf("./%s/%s.java", path, className)
	if _, err = os.Stat(path); os.IsNotExist(err) {
		if err = os.MkdirAll(path, os.ModePerm); err != nil {
			panic(err)
		}
	}
	_ = os.WriteFile(filename, []byte(codes), 0666)
	fmt.Printf("%s: %s\n", className, filename)
}

func genRepository(tableStatus *TableStatus) {
	entityName := tryRemoveTablePrefix(tableStatus.Name)
	className := fmt.Sprintf("%sRepository", firstUpCase(camelCase(entityName)))
	mapperClassName := fmt.Sprintf("%sMapper", firstUpCase(camelCase(entityName)))
	mapperFieldName := fmt.Sprintf("%sMapper", camelCase(entityName))

	codes := fmt.Sprintf(`package com.mahuafm.phoenix.%s.infrastructure.repository;

import com.mahuafm.phoenix.%s.infrastructure.persistence.mapper.PointsConfigMapper;
import javax.annotation.Resource;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Repository;

%s
@Repository
@Slf4j
public class %s {

  @Resource
  private %s %s;

}
`, domainName, domainName, genJavadoc(className, tableStatus), className, mapperClassName, mapperFieldName)

	path := fmt.Sprintf("./%s", filepath.Join("ddd-code-gen", domainName, "infrastructure", "repository"))
	filename := fmt.Sprintf("./%s/%s.java", path, className)
	if _, err = os.Stat(path); os.IsNotExist(err) {
		if err = os.MkdirAll(path, os.ModePerm); err != nil {
			panic(err)
		}
	}
	_ = os.WriteFile(filename, []byte(codes), 0666)
	fmt.Printf("%s: %s\n", className, filename)
}

// connectToDB inits mDB to connect to the Database.
func connectToDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, host, port, schemaName)
	fmt.Printf("dsn: %s\n\n", dsn)
	if mDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{}); err != nil {
		panic(err)
	}
}

// readTablePrototype returns table status and columns by fetching MySQL.
func readTablePrototype() (*TableStatus, []DDLM) {
	tableStatus := &TableStatus{}
	if err = mDB.Raw(fmt.Sprintf(sqlShowTableStatus, tableName)).First(&tableStatus).Error; err != nil {
		panic(err)
	}
	ddlms := make([]DDLM, 0)
	if err = mDB.Raw(fmt.Sprintf(sqlShowFullColumns, tableName)).Find(&ddlms).Error; err != nil {
		panic(err)
	}
	return tableStatus, ddlms
}

// getStructType returns Java type by DB type
func getStructType(s DDLM) (string, string) {
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
