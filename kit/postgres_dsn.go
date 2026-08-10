package kit

import "fmt"

// PostgresDSN holds the fields needed to build a Postgres connection string.
type PostgresDSN struct {
	Username string
	Password string
	Database string
	Host     string
	// Port ปล่อยเป็น 0 ได้ จะใช้ 5432 (เพิ่มทีหลัง จึงไม่กระทบ struct literal เดิม)
	Port int
	// Instance = Cloud SQL connection name (project:region:instance)
	// ถ้าตั้งค่าไว้จะชนะ Host แล้วต่อผ่าน unix socket /cloudsql/<instance>
	Instance string
}

// CreatePostgresDSN builds a Postgres DSN string.
//
// Deprecated: ใช้ infra.OpenDB กับ infra.BuildPostgresDSN แทน ตัวนี้เหลือไว้เพื่อความ
// เข้ากันได้ย้อนหลังเท่านั้น และไม่มีผู้เรียกภายในโมดูลแล้ว
func CreatePostgresDSN(dsn PostgresDSN) string {
	host := dsn.Host
	if dsn.Instance != "" {
		host = "/cloudsql/" + dsn.Instance
	}
	port := dsn.Port
	if port == 0 {
		port = 5432
	}
	// sslmode=prefer ไม่ใช่ disable: ลอง TLS ก่อนแล้วค่อยถอย เพื่อให้ต่อ Cloud SQL ที่ตั้ง
	// sslMode=ENCRYPTED_ONLY ได้ด้วย
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=prefer",
		host, dsn.Username, dsn.Password, dsn.Database, port)
}