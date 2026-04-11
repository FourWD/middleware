package infra

import (
	"context"
	"fmt"
)

func initInfrastructure(
	cfg CommonConfig,
	appLogger *Logger,
	cleanupHooks *[]func(context.Context) error,
) (Databases, *RedisClient, *MongoClient, *FirebaseClient, BlacklistStore, *PubSubClient, *StorageClient, *MailClient, error) {
	if err := RunMigrations(cfg); err != nil {
		return Databases{}, nil, nil, nil, nil, nil, nil, nil, err
	}

	primary, err := OpenDB(cfg.Database, appLogger)
	if err != nil {
		return Databases{}, nil, nil, nil, nil, nil, nil, nil, err
	}
	*cleanupHooks = append(*cleanupHooks, func(context.Context) error {
		sqlDB, err := primary.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	})

	dbs := Databases{Primary: primary}

	if cfg.SecondaryDBEnabled {
		secondary, err := OpenDB(cfg.SecondaryDatabase, appLogger)
		if err != nil {
			return Databases{}, nil, nil, nil, nil, nil, nil, nil, err
		}
		*cleanupHooks = append(*cleanupHooks, func(context.Context) error {
			sqlDB, err := secondary.DB()
			if err != nil {
				return err
			}
			return sqlDB.Close()
		})
		dbs.Secondary = secondary
	}

	var redisClient *RedisClient
	if cfg.RedisEnabled {
		redisClient = NewRedisClient(cfg.Redis)
		pingCtx, cancel := context.WithTimeout(context.Background(), BootstrapTimeout)
		defer cancel()
		if err := redisClient.Ping(pingCtx).Err(); err != nil {
			return Databases{}, nil, nil, nil, nil, nil, nil, nil, err
		}
		*cleanupHooks = append(*cleanupHooks, func(context.Context) error {
			return redisClient.Close()
		})
	}

	var mongoClient *MongoClient
	if cfg.MongoEnabled {
		mongoCtx, mongoCancel := context.WithTimeout(context.Background(), BootstrapTimeout)
		defer mongoCancel()
		mc, err := ConnectMongo(mongoCtx, cfg.Mongo)
		if err != nil {
			return Databases{}, nil, nil, nil, nil, nil, nil, nil, fmt.Errorf("mongo: %w", err)
		}
		*cleanupHooks = append(*cleanupHooks, mc.Close)
		mongoClient = mc
	}

	var firebaseClient *FirebaseClient
	if cfg.Firebase.CredentialsFile != "" || cfg.Firebase.NotificationCredentialsFile != "" {
		firebaseCtx, firebaseCancel := context.WithTimeout(context.Background(), BootstrapTimeout)
		defer firebaseCancel()
		fc, err := NewFirebaseClient(firebaseCtx, cfg.Firebase)
		if err != nil {
			return Databases{}, nil, nil, nil, nil, nil, nil, nil, fmt.Errorf("firebase: %w", err)
		}
		if fc != nil {
			*cleanupHooks = append(*cleanupHooks, func(context.Context) error {
				return fc.Close()
			})
			firebaseClient = fc
		}
	}

	var blacklistStore BlacklistStore
	if cfg.Auth.BlacklistEnabled {
		switch {
		case mongoClient != nil:
			blacklistStore = NewMongoBlacklistStore(mongoClient)
		case redisClient != nil:
			blacklistStore = NewRedisBlacklistStore(redisClient)
		default:
			return Databases{}, nil, nil, nil, nil, nil, nil, nil, fmt.Errorf("jwt blacklist requires mongo or redis")
		}
	}

	var pubSubClient *PubSubClient
	if cfg.PubSub.Enabled {
		pubSubCtx, pubSubCancel := context.WithTimeout(context.Background(), BootstrapTimeout)
		defer pubSubCancel()
		client, err := NewPubSubClient(pubSubCtx, cfg.PubSub)
		if err != nil {
			return Databases{}, nil, nil, nil, nil, nil, nil, nil, fmt.Errorf("pubsub: %w", err)
		}
		*cleanupHooks = append(*cleanupHooks, func(context.Context) error {
			return client.Close()
		})
		pubSubClient = client
	}

	var storageClient *StorageClient
	if cfg.Storage.Enabled {
		storageCtx, storageCancel := context.WithTimeout(context.Background(), BootstrapTimeout)
		defer storageCancel()
		client, err := NewStorageClient(storageCtx, cfg.Storage)
		if err != nil {
			return Databases{}, nil, nil, nil, nil, nil, nil, nil, fmt.Errorf("storage: %w", err)
		}
		*cleanupHooks = append(*cleanupHooks, func(context.Context) error {
			return client.Close()
		})
		storageClient = client
	}

	var mailClient *MailClient
	if cfg.Mail.Enabled {
		mailClient = NewMailClient(cfg.Mail)
	}

	return dbs, redisClient, mongoClient, firebaseClient, blacklistStore, pubSubClient, storageClient, mailClient, nil
}
