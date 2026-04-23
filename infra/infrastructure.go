package infra

import (
	"context"
	"fmt"
	"strings"
)

// InfraClients bundles all the clients built by initInfrastructure so the
// return type does not balloon with each new component.
type InfraClients struct {
	Databases       Databases
	Redis           *RedisClient
	Mongo           *MongoClient
	MongoMiddleware *MongoClient
	Firebase        *FirebaseClient
	Blacklist       BlacklistStore
	PubSub          *PubSubClient
	Storage         *StorageClient
	Mail            *MailClient
}

func initInfrastructure(
	cfg CommonConfig,
	appLogger *Logger,
	cleanupHooks *[]func(context.Context) error,
) (InfraClients, error) {
	out := InfraClients{}

	// Primary database is opt-out: empty DB_NAME means the service runs
	// without a database (gateway, webhook relay, etc.). Migrations also
	// skip automatically when no DB is available.
	if strings.TrimSpace(cfg.Database.Name) != "" {
		if err := RunMigrations(cfg); err != nil {
			return out, err
		}

		primary, err := OpenDB(cfg.Database, appLogger)
		if err != nil {
			return out, err
		}
		*cleanupHooks = append(*cleanupHooks, func(context.Context) error {
			sqlDB, err := primary.DB()
			if err != nil {
				return err
			}
			return sqlDB.Close()
		})
		out.Databases.Primary = primary
	}

	if cfg.SecondaryDBEnabled {
		secondary, err := OpenDB(cfg.SecondaryDatabase, appLogger)
		if err != nil {
			return out, err
		}
		*cleanupHooks = append(*cleanupHooks, func(context.Context) error {
			sqlDB, err := secondary.DB()
			if err != nil {
				return err
			}
			return sqlDB.Close()
		})
		out.Databases.Secondary = secondary
	}

	if cfg.RedisEnabled {
		redisClient := NewRedisClient(cfg.Redis)
		pingCtx, cancel := context.WithTimeout(context.Background(), BootstrapTimeout)
		defer cancel()
		if err := redisClient.Ping(pingCtx).Err(); err != nil {
			return out, err
		}
		*cleanupHooks = append(*cleanupHooks, func(context.Context) error {
			return redisClient.Close()
		})
		out.Redis = redisClient
	}

	if cfg.MongoEnabled {
		mongoCtx, mongoCancel := context.WithTimeout(context.Background(), BootstrapTimeout)
		defer mongoCancel()
		mc, err := ConnectMongo(mongoCtx, cfg.Mongo)
		if err != nil {
			return out, fmt.Errorf("mongo: %w", err)
		}
		*cleanupHooks = append(*cleanupHooks, mc.Close)
		out.Mongo = mc
	}

	if cfg.MongoMiddlewareEnabled {
		mongoCtx, mongoCancel := context.WithTimeout(context.Background(), BootstrapTimeout)
		defer mongoCancel()
		mc, err := ConnectMongo(mongoCtx, cfg.MongoMiddleware)
		if err != nil {
			return out, fmt.Errorf("mongo middleware: %w", err)
		}
		*cleanupHooks = append(*cleanupHooks, mc.Close)
		out.MongoMiddleware = mc
	}

	if cfg.Firebase.CredentialsFile != "" || cfg.Firebase.NotificationCredentialsFile != "" {
		firebaseCtx, firebaseCancel := context.WithTimeout(context.Background(), BootstrapTimeout)
		defer firebaseCancel()
		fc, err := NewFirebaseClient(firebaseCtx, cfg.Firebase)
		if err != nil {
			return out, fmt.Errorf("firebase: %w", err)
		}
		if fc != nil {
			*cleanupHooks = append(*cleanupHooks, func(context.Context) error {
				return fc.Close()
			})
			out.Firebase = fc
		}
	}

	if cfg.Auth.BlacklistEnabled {
		switch {
		case out.MongoMiddleware != nil:
			out.Blacklist = NewMongoBlacklistStore(out.MongoMiddleware)
		case out.Redis != nil:
			out.Blacklist = NewRedisBlacklistStore(out.Redis)
		default:
			return out, fmt.Errorf("jwt blacklist requires MONGO_MIDDLEWARE_ENABLED or REDIS_ENABLED")
		}
	}

	if cfg.PubSub.Enabled {
		pubSubCtx, pubSubCancel := context.WithTimeout(context.Background(), BootstrapTimeout)
		defer pubSubCancel()
		client, err := NewPubSubClient(pubSubCtx, cfg.PubSub)
		if err != nil {
			return out, fmt.Errorf("pubsub: %w", err)
		}
		*cleanupHooks = append(*cleanupHooks, func(context.Context) error {
			return client.Close()
		})
		out.PubSub = client
	}

	if cfg.Storage.Enabled {
		storageCtx, storageCancel := context.WithTimeout(context.Background(), BootstrapTimeout)
		defer storageCancel()
		client, err := NewStorageClient(storageCtx, cfg.Storage)
		if err != nil {
			return out, fmt.Errorf("storage: %w", err)
		}
		*cleanupHooks = append(*cleanupHooks, func(context.Context) error {
			return client.Close()
		})
		out.Storage = client
	}

	if cfg.Mail.Enabled {
		out.Mail = NewMailClient(cfg.Mail)
	}

	return out, nil
}
