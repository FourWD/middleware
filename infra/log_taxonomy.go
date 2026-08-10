package infra

// Component identifies the subsystem emitting a log entry. Use these
// constants with WithComponent so dashboards can aggregate by subsystem
// instead of grepping ad-hoc strings.
const (
	ComponentApp        = "app"         // NewApp, App.Run, lifecycle
	ComponentHTTP       = "http"        // request log, server middleware
	ComponentDB         = "db"          // GORM, SQL pool
	ComponentMongo      = "mongo"       // MongoDB ops
	ComponentRedis      = "redis"       // Redis ops
	ComponentAuth       = "auth"        // JWT, blacklist, refresh tokens
	ComponentFirebase   = "firebase"    // Firestore, FCM, Firebase Auth
	ComponentPubSub     = "pubsub"      // Google Pub/Sub
	ComponentStorage    = "storage"     // Google Cloud Storage
	ComponentMail       = "mail"        // Mailgun
	ComponentHTTPClient = "http_client" // outbound HTTP requests
	ComponentCron       = "cron"        // scheduled jobs
	ComponentLeader     = "leader"      // active/standby leader lock
	ComponentHandler    = "handler"     // route handler (service repo)
	ComponentRepo       = "repo"        // repository layer (service repo)
	ComponentPayment    = "payment"     // payment gateway integrations
	ComponentOTP        = "otp"         // SMS OTP provider
	ComponentUpload     = "upload"      // file upload service
)

// LogKind classifies the log entry's purpose. Use these constants with
// WithLogKind so log routing / alerting rules can target the right slice.
const (
	LogKindRequest    = "request"        // HTTP request log
	LogKindBusiness   = "business_event" // domain event (user created, order placed)
	LogKindLifecycle  = "lifecycle"      // boot, shutdown, worker start/stop
	LogKindError      = "error"          // failed operation
	LogKindSecurity   = "security"       // auth failure, suspicious activity
	LogKindDiagnostic = "diagnostic"     // debug, internal state
	LogKindStartup    = "startup"        // boot-time only
	LogKindInfra      = "infrastructure" // infra client connect/disconnect
)
