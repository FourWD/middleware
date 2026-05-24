package common

import (
	"database/sql"
	"testing"
)

func TestFormatDataType_WithLength(t *testing.T) {
	got := formatDataType("varchar", sql.NullInt64{Int64: 255, Valid: true})
	if got != "varchar (255)" {
		t.Fatalf("got %q want %q", got, "varchar (255)")
	}
}

func TestFormatDataType_WithoutLength(t *testing.T) {
	got := formatDataType("int", sql.NullInt64{Valid: false})
	if got != "int" {
		t.Fatalf("got %q want %q", got, "int")
	}
}

func TestFinalizeTable_ComputesTotalAndMD5(t *testing.T) {
	in := tableInfo{
		TableName: "users",
		ColumnList: []columnInfo{
			{Name: "id", Type: "varchar (36)"},
			{Name: "email", Type: "varchar (255)"},
		},
	}
	out := finalizeTable(in)

	if out.TotalColumn != 2 {
		t.Fatalf("TotalColumn: got %d want 2", out.TotalColumn)
	}
	if out.Md5 == "" {
		t.Fatal("Md5 should be computed")
	}
	if len(out.Md5) != 32 {
		t.Fatalf("Md5 length: got %d want 32 (hex md5)", len(out.Md5))
	}
}

func TestFinalizeTable_DeterministicMD5(t *testing.T) {
	in := tableInfo{
		TableName:  "users",
		ColumnList: []columnInfo{{Name: "id", Type: "varchar (36)"}},
	}
	a := finalizeTable(in)
	b := finalizeTable(in)
	if a.Md5 != b.Md5 {
		t.Fatalf("MD5 not deterministic: %s vs %s", a.Md5, b.Md5)
	}
}

func TestFinalizeTable_DifferentSchemaDifferentMD5(t *testing.T) {
	a := finalizeTable(tableInfo{
		TableName:  "users",
		ColumnList: []columnInfo{{Name: "id", Type: "varchar (36)"}},
	})
	b := finalizeTable(tableInfo{
		TableName:  "users",
		ColumnList: []columnInfo{{Name: "id", Type: "varchar (40)"}},
	})
	if a.Md5 == b.Md5 {
		t.Fatal("schema drift should produce different MD5")
	}
}

func TestFinalizeTable_EmptyColumnList(t *testing.T) {
	in := tableInfo{TableName: "empty", ColumnList: []columnInfo{}}
	out := finalizeTable(in)
	if out.TotalColumn != 0 {
		t.Fatalf("got %d want 0", out.TotalColumn)
	}
	if out.Md5 == "" {
		t.Fatal("Md5 should still compute for empty table")
	}
}
