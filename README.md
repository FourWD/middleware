รีวิวรอบใหม่ — สถานะหลังการ refactor
🟢 Build / Vet / Tests
ผ่านทั้ง 3 packages — infra, kit, common (kafka ไม่มี caller, model, orm, permission ไม่มี test)

โครงสร้างปัจจุบัน
Package ไฟล์ บทบาท
infra 67 Framework + lifecycle + clients (DB/Redis/Mongo/Firebase/PubSub/Storage/Mail) + middleware stack
common 38 Business helpers + thin bridges to infra (legacy compat)
kit 41 Stateless utilities
kafka 3 Orphan — ไม่มี caller ใน repo, ไม่ผูกกับ infra
model, orm, permission — Schema/types — OK
✅ สิ่งที่ดี (เพิ่งทำเสร็จ)
Single source of truth — env → LoadCommonConfig → cfg ไหลทุกที่
Logger — รวม LOG_LEVEL, APP_VERSION เข้า CommonConfig (ไม่อ่าน env หลายจุด)
APP_ID — รวม APP_NAME ⇄ APP_ID เป็นค่าเดียว (auto override จาก GAE_SERVICE บน GAE)
JWT_ISSUER — auto = APP_ID (ไม่ต้อง env)
3-tier Rate Limit — Strict/Default/Skip ผ่าน route group
Worker API — graceful lifecycle + log + bounded shutdown
GAE auto-restart — registerGAEVersionCheck (auto only on GAE)
DB pool metrics — Prometheus collector lazy scrape
Mongo / Firebase / PubSub — รวมเป็น singletons ใน infra; common bridge เหลือแค่ DB
Pubsub v1 → v2 — รวมเป็น v2 อย่างเดียว
JWT v4 → v5 — direct dep เป็น v5 (v4 indirect via Firebase SDK เท่านั้น)
viper หาย 100% — ใน common ไม่มี import แล้ว
common.MigrateInfra ผอม — เหลือแค่ bridge Database/DatabaseSql
common/mongo.go หาย — ใช้ infra.Mongo ตรง
🔴 ที่ยังมีอยู่ใน common — ควรพิจารณาต่อ
Duplicate กับ infra (full)
ไฟล์ สถานะ คำแนะนำ
common/connect_mysql_database.go ซ้ำ infra.OpenDB ลบได้ ถ้าไม่มี caller
common/connect_postgres_database.go เหมือนกัน ลบได้
common/database_dns.go ซ้ำ BuildMySQLDSN/BuildPostgresDSN ลบได้
common/init_database_connection_pool.go infra auto pool ลบได้
common/database.go ConnectDatabaseViper, ConnectDatabaseMySqlGoogle legacy (var Database/Sql ยังต้องเก็บ) ลบ function, เก็บ var
common/mail.go ซ้ำ infra.MailClient ลบได้
common/zip_file.go ซ้ำ kit.zip ตรวจ — ถ้าซ้ำ ลบ
common/download_file.go ซ้ำ kit/gcs ลบได้
Partial duplicate (overlap แต่ semantic ต่าง)
common infra สถานะ
AuthenticationMiddleware (jwt.go) bearer.go + jwt.go + require_role.go API ต่าง — เลือกใช้ตัวใดตัวหนึ่ง
GenerateJWTToken, DecodeJWT TokenManager.GenerateTokens / Validate common simple, infra full-featured
JWTClaims struct infra.Claims 2 type ใกล้เคียง
Fiber response (FiberCustom/OK/Error) WriteSuccess, WriteErrorEnvelope format ต่างกัน — ขัดแย้ง
IsJwtValid, BlacklistJwtToken infra.BlacklistStore (Mongo + Redis impls) infra สมบูรณ์กว่า
🟡 Code smell ที่เหลือ

1. kafka package — orphan
   ไม่มี caller ใน repo
   ไม่ผูกกับ infra (ไม่มี client config / lifecycle)
   ลบทิ้ง หรือ integrate เข้า infra
2. common/get_app_engine_log.go — admin tool, ใช้ที่เดียว?
   อ่าน Cloud Logging API
   ถ้าไม่มี caller → ลบ
3. ไฟล์ใน common ที่ทำหน้าที่ "factory" สำหรับ external service (business logic)
   discord.go, otp.go, payment_2c2p*.go, upload*.go — business specific ควรอยู่ที่ project ปลายทาง ไม่ใช่ middleware กลาง
   ระยะยาว: ย้ายเข้า project ปลายทาง
4. common/get_noti_token.go, common/create_log_action.go — business
   พึ่ง common.Database + ORM model
   ผูกแน่นกับ business → ควรอยู่ project ปลายทาง
5. common/check_role_permission\*.go — RBAC
   มี infra/require_role.go แล้ว — ตรวจว่าซ้ำหรือคนละ approach
6. go.mod — ตรวจ unused deps (โอกาส go mod tidy อีกรอบ)
   🟠 ปัญหาเชิงสถาปัตยกรรม
   ปัจจุบัน — common เป็น "junk drawer"
   มี 3 หมวดปนกัน:

A. Bridge helpers (MigrateInfra, log wrappers) → keep
B. Generic utilities (Encrypt/Decrypt, FiberPaginatedQuery) → ย้ายไป kit หรือ infra
C. Business logic (OTP/Payment/Upload/Discord) → ย้ายไป project ปลายทาง
ทางออก
Phase 1 (ระยะกลาง) — ย้าย B → kit/infra:

Encrypt/Decrypt → kit/crypto.go
FiberDisableXFrame/FiberNoSniff → infra/security_headers.go
Discord\* → kit/discord.go (มีอยู่แล้ว) ตรวจ overlap
Phase 2 (ระยะยาว) — ย้าย C ออกจาก middleware repo:

ทุกธุรกิจเฉพาะ: payment, otp, upload, get_noti_token → project repo
เหลือ middleware ที่เป็น generic infrastructure ล้วน
คำถามสำหรับวางแผนต่อ
kafka — ใช้จริงที่ไหน? ถ้าใช้: integrate เข้า infra (config + client + lifecycle) ถ้าไม่: ลบ
Business code (otp, payment, upload, discord, get_app_engine_log) — ตั้งใจให้อยู่ middleware ตลอดไป หรือจะย้ายไป project repo?
AuthenticationMiddleware vs infra.bearer + require_role — ใช้ตัวไหน? consolidate ไหม?
Fiber response 2 format — รวมเป็นอันเดียวไหม?
Quick wins ที่ทำได้ทันที (ไม่ break anything)
ลบ 7 ไฟล์ DB legacy (connect_mysql, connect_postgres, database_dns, init_database_connection_pool, ConnectDatabaseViper เฉพาะ function) — 30 นาที
ลบ common/mail.go (ซ้ำ infra) — 5 นาที
ตรวจ kafka ลบหรือ integrate — 10 นาที
ตรวจ common/zip_file.go ↔ kit/zip.go — 5 นาที
go mod tidy รอบใหม่ — 1 นาที
อยากให้ทำ Quick wins ทั้งหมดไหมครับ?
