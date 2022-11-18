package main

// TableStatus defines mysql table status we needed
type TableStatus struct {
	Name      string
	Rows      uint64
	Collation string
	Comment   string
}

// ColumnsStatement defines mysql columns statement we needed
type ColumnsStatement struct {
	Field   string
	Type    string
	Null    string
	Key     string
	Comment string
	Default string
	Extra   string
}

// JavaField defines POJO members
type JavaField struct {
	JavaType    string
	Field       string
	Comment     string
	PackageName string
	IsPri       bool
}
