package main

import (
	"fmt"
	"strings"
)

func parseJavaImportsAndFields(javaFields []JavaField) (importCodes, fieldCodes string) {
	maxTypeStringLen := 0
	maxFieldStringLen := 0
	for _, v := range javaFields {
		if len(v.JavaType) > maxTypeStringLen {
			maxTypeStringLen = len(v.JavaType)
		}
		if len(v.Field) > maxFieldStringLen {
			maxFieldStringLen = len(v.Field)
		}
	}

	importCodes, fieldCodes = "", ""
	for _, v := range javaFields {
		if v.PackageName != "" && !strings.Contains(importCodes, v.PackageName) {
			importCodes += fmt.Sprintf("import %s;\n", v.PackageName)
		}
		if v.Field == "id" || v.Field == "ctime" || v.Field == "mtime" {
			continue
		}
		fieldCodes += fmt.Sprintf("  private %s %s;%s// %s\n", v.JavaType+strings.Repeat(" ", maxTypeStringLen-len(v.JavaType)), v.Field, strings.Repeat(" ", maxFieldStringLen-len(v.Field)+1), v.Comment)
	}
	importCodes = strings.TrimSuffix(importCodes, "\n")
	fieldCodes = strings.TrimSuffix(fieldCodes, "\n")
	return
}
