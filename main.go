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
	"gorm.io/gorm"
	"path/filepath"
	"strings"
	"time"
)

const genOutputDir = "gen-output"

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

var buildVersion string

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
	tableStatus, columns := readTablePrototype()

	// parse class members
	javaFields := parseJavaFields(columns)

	genPO(tableStatus, javaFields)
	genMapper(tableStatus)
	genRepository(tableStatus)
	genFactory(tableStatus)
	genEntity(tableStatus, javaFields)
	genAppService(tableStatus)
	genAssembler(tableStatus)
}

func genJavadoc(className string, tableStatus *TableStatus) string {
	content := `/**
 * {{className}} - {{tableComment}}
 *
 * <p>table_name: {{tableName}}</p>
 *
 * @author ddd-gen-mysql
 * @date {{timeString}}
 */`
	content = strings.ReplaceAll(content, "{{className}}", className)
	content = strings.ReplaceAll(content, "{{tableComment}}", tableStatus.Comment)
	content = strings.ReplaceAll(content, "{{tableName}}", tableStatus.Name)
	content = strings.ReplaceAll(content, "{{timeString}}", time.Now().String())
	return content
}

func genPO(tableStatus *TableStatus, javaFields []JavaField) {
	entityName := tryRemoveTablePrefix(tableStatus.Name)
	className := fmt.Sprintf("%sPo", firstUpCase(camelCase(entityName)))
	importCodes, fieldCodes := parseJavaImportsAndFields(javaFields)

	codes := `package com.mahuafm.phoenix.{{domainName}}.infrastructure.persistence.po;

import com.baomidou.mybatisplus.annotation.TableName;
import com.mahuafm.phoenix.util.infrastructure.persistence.po.base.BaseAutoIdPo;
{{importCodes}}
import lombok.Data;
import lombok.EqualsAndHashCode;

{{javadoc}}
@Data
@EqualsAndHashCode(callSuper = true)
@TableName("{{tableName}}")
public class {{className}} extends BaseAutoIdPo {

{{fieldCodes}}

}
`
	codes = strings.ReplaceAll(codes, "{{domainName}}", domainName)
	codes = strings.ReplaceAll(codes, "{{javadoc}}", genJavadoc(className, tableStatus))
	codes = strings.ReplaceAll(codes, "{{importCodes}}", importCodes)
	codes = strings.ReplaceAll(codes, "{{tableName}}", tableStatus.Name)
	codes = strings.ReplaceAll(codes, "{{className}}", className)
	codes = strings.ReplaceAll(codes, "{{fieldCodes}}", fieldCodes)

	path := fmt.Sprintf("%s", filepath.Join(genOutputDir, domainName, "infrastructure", "persistence", "po"))
	filename := fmt.Sprintf("./%s/%s.java", path, className)
	writeFile(path, filename, codes)
	fmt.Printf("%s: %s\n", className, filename)
}

func genMapper(tableStatus *TableStatus) {
	entityName := tryRemoveTablePrefix(tableStatus.Name)
	className := fmt.Sprintf("%sMapper", firstUpCase(camelCase(entityName)))
	poClassName := fmt.Sprintf("%sPo", firstUpCase(camelCase(entityName)))

	codes := `package com.mahuafm.phoenix.{{domainName}}.infrastructure.persistence.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mahuafm.phoenix.{{domainName}}.infrastructure.persistence.po.{{poClassName}};

{{javadoc}}
public interface {{className}} extends BaseMapper<{{poClassName}}> {}
`
	codes = strings.ReplaceAll(codes, "{{domainName}}", domainName)
	codes = strings.ReplaceAll(codes, "{{javadoc}}", genJavadoc(className, tableStatus))
	codes = strings.ReplaceAll(codes, "{{poClassName}}", poClassName)
	codes = strings.ReplaceAll(codes, "{{className}}", className)

	path := fmt.Sprintf("%s", filepath.Join(genOutputDir, domainName, "infrastructure", "persistence", "mapper"))
	filename := fmt.Sprintf("./%s/%s.java", path, className)
	writeFile(path, filename, codes)
	fmt.Printf("%s: %s\n", className, filename)
}

func genRepository(tableStatus *TableStatus) {
	entityName := tryRemoveTablePrefix(tableStatus.Name)
	className := fmt.Sprintf("%sRepository", firstUpCase(camelCase(entityName)))
	mapperClassName := fmt.Sprintf("%sMapper", firstUpCase(camelCase(entityName)))
	mapperFieldName := fmt.Sprintf("%sMapper", camelCase(entityName))

	codes := `package com.mahuafm.phoenix.{{domainName}}.infrastructure.repository;

import com.mahuafm.phoenix.{{domainName}}.infrastructure.persistence.mapper.{{mapperClassName}};
import javax.annotation.Resource;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Repository;

{{javadoc}}
@Repository
@Slf4j
public class {{className}} {

  @Resource
  private {{mapperClassName}} {{mapperFieldName}};

}
`
	codes = strings.ReplaceAll(codes, "{{domainName}}", domainName)
	codes = strings.ReplaceAll(codes, "{{javadoc}}", genJavadoc(className, tableStatus))
	codes = strings.ReplaceAll(codes, "{{mapperClassName}}", mapperClassName)
	codes = strings.ReplaceAll(codes, "{{mapperFieldName}}", mapperFieldName)
	codes = strings.ReplaceAll(codes, "{{className}}", className)

	path := fmt.Sprintf("%s", filepath.Join(genOutputDir, domainName, "infrastructure", "repository"))
	filename := fmt.Sprintf("./%s/%s.java", path, className)
	writeFile(path, filename, codes)
	fmt.Printf("%s: %s\n", className, filename)
}

func genFactory(tableStatus *TableStatus) {
	entityName := tryRemoveTablePrefix(tableStatus.Name)
	entityClassName := firstUpCase(camelCase(entityName))
	className := fmt.Sprintf("%sFactory", entityClassName)
	poClassName := fmt.Sprintf("%sPo", entityClassName)

	codes := `package com.mahuafm.phoenix.{{domainName}}.infrastructure.factory;

import com.mahuafm.phoenix.{{domainName}}.domain.{{domainName}}.entity.{{entityClassName}};
import com.mahuafm.phoenix.{{domainName}}.infrastructure.persistence.po.{{poClassName}};
import com.mahuafm.phoenix.util.bean.BeanCopyUtil;
import java.util.Collections;
import java.util.List;
import java.util.Objects;
import java.util.stream.Collectors;
import lombok.extern.slf4j.Slf4j;
import org.springframework.util.CollectionUtils;

{{javadoc}}
@Slf4j
public class {{className}} {

  public static {{entityClassName}} fromPo({{poClassName}} po) {
    if (po == null) {
      return null;
    }
    var entity = BeanCopyUtil.copy(po, {{entityClassName}}.class);
    // TODO extra code to invoke setter
    return entity;
  }

  public static List<{{entityClassName}}> fromPos(List<{{poClassName}}> pos) {
    if (CollectionUtils.isEmpty(pos)) {
      return Collections.emptyList();
    }
    return pos.stream()
        .map({{className}}::fromPo)
        .filter(Objects::nonNull)
        .collect(Collectors.toList());
  }

  public static {{poClassName}} toPo({{entityClassName}} entity) {
    var po = BeanCopyUtil.copy(entity, {{poClassName}}.class);
    // TODO extra code to invoke setter
    return po;
  }

  public static List<{{poClassName}}> toPos(List<{{entityClassName}}> entities) {
    if (CollectionUtils.isEmpty(entities)) {
      return Collections.emptyList();
    }
    return entities.stream()
        .map({{className}}::toPo)
        .filter(Objects::nonNull)
        .collect(Collectors.toList());
  }

}
`
	codes = strings.ReplaceAll(codes, "{{domainName}}", domainName)
	codes = strings.ReplaceAll(codes, "{{javadoc}}", genJavadoc(className, tableStatus))
	codes = strings.ReplaceAll(codes, "{{entityClassName}}", entityClassName)
	codes = strings.ReplaceAll(codes, "{{className}}", className)
	codes = strings.ReplaceAll(codes, "{{poClassName}}", poClassName)

	path := fmt.Sprintf("%s", filepath.Join(genOutputDir, domainName, "infrastructure", "factory"))
	filename := fmt.Sprintf("./%s/%s.java", path, className)
	writeFile(path, filename, codes)
	fmt.Printf("%s: %s\n", className, filename)
}

func genEntity(tableStatus *TableStatus, javaFields []JavaField) {
	entityName := tryRemoveTablePrefix(tableStatus.Name)
	className := firstUpCase(camelCase(entityName))
	importCodes, fieldCodes := parseJavaImportsAndFields(javaFields)

	codes := `package com.mahuafm.phoenix.{{domainName}}.domain.{{domainName}}.entity;

{{importCodes}}
import lombok.Data;

{{javadoc}}
@Data
public class {{className}} {

{{fieldCodes}}

}
`
	codes = strings.ReplaceAll(codes, "{{domainName}}", domainName)
	codes = strings.ReplaceAll(codes, "{{javadoc}}", genJavadoc(className, tableStatus))
	codes = strings.ReplaceAll(codes, "{{className}}", className)
	codes = strings.ReplaceAll(codes, "{{importCodes}}", importCodes)
	codes = strings.ReplaceAll(codes, "{{fieldCodes}}", fieldCodes)

	path := fmt.Sprintf("%s", filepath.Join(genOutputDir, domainName, "domain", domainName, "entity"))
	filename := fmt.Sprintf("./%s/%s.java", path, className)
	writeFile(path, filename, codes)
	fmt.Printf("%s: %s\n", className, filename)
}

func genAppService(tableStatus *TableStatus) {
	entityName := tryRemoveTablePrefix(tableStatus.Name)
	entityClassName := firstUpCase(camelCase(entityName))
	className := fmt.Sprintf("%sAppService", entityClassName)
	repositoryClassName := fmt.Sprintf("%sRepository", entityClassName)
	repositoryFieldName := fmt.Sprintf("%sRepository", camelCase(entityName))

	codes := `package com.mahuafm.phoenix.{{domainName}}.application.service;

import com.mahuafm.phoenix.{{domainName}}.infrastructure.repository.{{repositoryClassName}};
import javax.annotation.Resource;
import org.springframework.stereotype.Service;

{{javadoc}}
@Service
public class {{className}} {

  @Resource
  private {{repositoryClassName}} {{repositoryFieldName}};

}
`
	codes = strings.ReplaceAll(codes, "{{domainName}}", domainName)
	codes = strings.ReplaceAll(codes, "{{javadoc}}", genJavadoc(className, tableStatus))
	codes = strings.ReplaceAll(codes, "{{className}}", className)
	codes = strings.ReplaceAll(codes, "{{repositoryClassName}}", repositoryClassName)
	codes = strings.ReplaceAll(codes, "{{repositoryFieldName}}", repositoryFieldName)

	path := fmt.Sprintf("%s", filepath.Join(genOutputDir, domainName, "application", "service"))
	filename := fmt.Sprintf("./%s/%s.java", path, className)
	writeFile(path, filename, codes)
	fmt.Printf("%s: %s\n", className, filename)
}

func genAssembler(tableStatus *TableStatus) {
	entityName := tryRemoveTablePrefix(tableStatus.Name)
	className := fmt.Sprintf("%sAssembler", firstUpCase(camelCase(entityName)))
	codes := `package com.mahuafm.phoenix.{{domainName}}.application.assembler;

{{javadoc}}
public class {{className}} {
}
`
	codes = strings.ReplaceAll(codes, "{{domainName}}", domainName)
	codes = strings.ReplaceAll(codes, "{{javadoc}}", genJavadoc(className, tableStatus))
	codes = strings.ReplaceAll(codes, "{{className}}", className)

	path := fmt.Sprintf("%s", filepath.Join(genOutputDir, domainName, "application", "assembler"))
	filename := fmt.Sprintf("./%s/%s.java", path, className)
	writeFile(path, filename, codes)
	fmt.Printf("%s: %s\n", className, filename)
}
