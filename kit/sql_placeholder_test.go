package kit

import "testing"

func TestToPostgresPlaceholders(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"none", "SELECT 1", "SELECT 1"},
		{"single", "SELECT * FROM t WHERE id = ?", "SELECT * FROM t WHERE id = $1"},
		{"multiple", "SELECT * FROM t WHERE a = ? AND b = ?", "SELECT * FROM t WHERE a = $1 AND b = $2"},
		{
			"question inside literal",
			"SELECT * FROM t WHERE note LIKE '%?%' AND id = ?",
			"SELECT * FROM t WHERE note LIKE '%?%' AND id = $1",
		},
		{
			"escaped quote inside literal",
			"SELECT 'it''s ? here', ? FROM t",
			"SELECT 'it''s ? here', $1 FROM t",
		},
		{
			"quoted identifier",
			`SELECT "we?rd" FROM t WHERE id = ?`,
			`SELECT "we?rd" FROM t WHERE id = $1`,
		},
		{
			"mysql backtick identifier",
			"SELECT `we?rd` FROM t WHERE id = ?",
			"SELECT `we?rd` FROM t WHERE id = $1",
		},
		{
			"line comment",
			"SELECT 1 -- what? really\nWHERE id = ?",
			"SELECT 1 -- what? really\nWHERE id = $1",
		},
		{
			"block comment",
			"SELECT /* a ? b */ 1 WHERE id = ?",
			"SELECT /* a ? b */ 1 WHERE id = $1",
		},
		{
			"nested block comment",
			"SELECT /* a /* ? */ b ? */ 1 WHERE id = ?",
			"SELECT /* a /* ? */ b ? */ 1 WHERE id = $1",
		},
		{
			"dollar quoted body",
			"SELECT $$what? no$$, ? FROM t",
			"SELECT $$what? no$$, $1 FROM t",
		},
		{
			"tagged dollar quote",
			"SELECT $tag$what? no$tag$, ? FROM t",
			"SELECT $tag$what? no$tag$, $1 FROM t",
		},
		{
			"escaped question is jsonb operator",
			"SELECT * FROM t WHERE data ?? 'key' AND id = ?",
			"SELECT * FROM t WHERE data ? 'key' AND id = $1",
		},
		{
			"existing positional left alone",
			"SELECT * FROM t WHERE a = $1 AND b = ?",
			"SELECT * FROM t WHERE a = $1 AND b = $1",
		},
		{
			"numbering continues past skipped regions",
			"SELECT ? /* ? */ , ? , '?' , ?",
			"SELECT $1 /* ? */ , $2 , '?' , $3",
		},
		{"unterminated literal", "SELECT 'oops ? ", "SELECT 'oops ? "},
		{"unterminated block comment", "SELECT /* oops ? ", "SELECT /* oops ? "},
		{"empty", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ToPostgresPlaceholders(tc.in); got != tc.want {
				t.Fatalf("\n got: %s\nwant: %s", got, tc.want)
			}
		})
	}
}
