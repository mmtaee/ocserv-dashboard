package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"gorm.io/gorm"
)

var Migration012 = &gormigrate.Migration{
	ID: "012_remove_uid_add_admins_and_tokens",

	Migrate: func(tx *gorm.DB) error {
		// --- Step 1: Remove uid from ocserv_users ---
		if err := tx.Exec(`
			ALTER TABLE ocserv_users
			DROP COLUMN IF EXISTS uid;
		`).Error; err != nil {
			return err
		}
		logger.Info("migration: removed uid from ocserv_users")

		// --- Step 2: Create administrators table ---
		if err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS administrators (
				id BIGSERIAL PRIMARY KEY,
				username VARCHAR(255) NOT NULL UNIQUE,
				password VARCHAR(255) NOT NULL,
				role VARCHAR(32) NOT NULL DEFAULT 'admin',
				last_login TIMESTAMP WITH TIME ZONE NULL,
				created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
			);
		`).Error; err != nil {
			return err
		}
		logger.Info("migration: created administrators table")

		// --- Step 3: Move data from users to administrators ---
		if err := tx.Exec(`
			INSERT INTO administrators (id, username, password, role, created_at, updated_at)
			SELECT id, username, password, CASE WHEN is_admin THEN 'super' ELSE 'admin' END, created_at, updated_at
			FROM users
			ON CONFLICT DO NOTHING;
		`).Error; err != nil {
			return err
		}
		logger.Info("migration: migrated users to administrators")

		// --- Step 4: Drop old tables user_tokens and users ---
		if err := tx.Exec(`DROP TABLE IF EXISTS user_tokens;`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`DROP TABLE IF EXISTS users;`).Error; err != nil {
			return err
		}
		logger.Info("migration: dropped old users and user_tokens tables")

		// --- Step 5: Update ocserv_users: add owner_admin_id, drop owner ---
		if err := tx.Exec(`
			ALTER TABLE ocserv_users
			ADD COLUMN IF NOT EXISTS owner_admin_id BIGINT,
			DROP COLUMN IF EXISTS owner;
		`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			CREATE INDEX IF NOT EXISTS idx_ocserv_users_owner_admin_id ON ocserv_users(owner_admin_id);
		`).Error; err != nil {
			return err
		}
		logger.Info("migration: updated ocserv_users with owner_admin_id")

		// --- Step 6: Update ocserv_groups: add owner_admin_id, drop owner ---
		if err := tx.Exec(`
			ALTER TABLE ocserv_groups
			ADD COLUMN IF NOT EXISTS owner_admin_id BIGINT,
			DROP COLUMN IF EXISTS owner;
		`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			CREATE INDEX IF NOT EXISTS idx_ocserv_groups_owner_admin_id ON ocserv_groups(owner_admin_id);
		`).Error; err != nil {
			return err
		}
		logger.Info("migration: updated ocserv_groups with owner_admin_id")

		// --- Step 7: Update telegram_packages: add owner_admin_id ---
		if err := tx.Exec(`
			ALTER TABLE telegram_packages
			ADD COLUMN IF NOT EXISTS owner_admin_id BIGINT;
		`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			CREATE INDEX IF NOT EXISTS idx_telegram_packages_owner_admin_id ON telegram_packages(owner_admin_id);
		`).Error; err != nil {
			return err
		}
		logger.Info("migration: updated telegram_packages with owner_admin_id")

		// --- Step 8: Update telegram_requests: add owner_admin_id ---
		if err := tx.Exec(`
			ALTER TABLE telegram_requests
			ADD COLUMN IF NOT EXISTS owner_admin_id BIGINT;
		`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			CREATE INDEX IF NOT EXISTS idx_telegram_requests_owner_admin_id ON telegram_requests(owner_admin_id);
		`).Error; err != nil {
			return err
		}
		logger.Info("migration: updated telegram_requests with owner_admin_id")

		// --- Step 9: Create administrator_tokens table ---
		if err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS administrator_tokens (
				id BIGSERIAL PRIMARY KEY,
				administrator_id BIGINT NOT NULL,
				token TEXT NOT NULL,
				created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
				expire_at TIMESTAMP WITH TIME ZONE NOT NULL
			);
		`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			CREATE INDEX IF NOT EXISTS idx_administrator_tokens_administrator_id ON administrator_tokens(administrator_id);
		`).Error; err != nil {
			return err
		}
		logger.Info("migration: created administrator_tokens table")

		// --- Step 10: Add foreign key constraint ---
		if err := tx.Exec(`
			ALTER TABLE administrator_tokens
			ADD CONSTRAINT fk_administrator_tokens_administrator
			FOREIGN KEY (administrator_id)
			REFERENCES administrators(id)
			ON DELETE CASCADE
			ON UPDATE CASCADE;
		`).Error; err != nil {
			return err
		}
		logger.Info("migration: added foreign key to administrator_tokens")

		logger.Info("migration 012 complete successfully")
		return nil
	},

	Rollback: func(tx *gorm.DB) error {
		// --- Rollback Step 10 ---
		if err := tx.Exec(`
			ALTER TABLE administrator_tokens
			DROP CONSTRAINT IF EXISTS fk_administrator_tokens_administrator;
		`).Error; err != nil {
			return err
		}

		// --- Rollback Step 9 ---
		if err := tx.Exec(`
			DROP INDEX IF EXISTS idx_administrator_tokens_administrator_id;
		`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`DROP TABLE IF EXISTS administrator_tokens;`).Error; err != nil {
			return err
		}

		// --- Rollback Step 8 ---
		if err := tx.Exec(`
			DROP INDEX IF EXISTS idx_telegram_requests_owner_admin_id;
		`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			ALTER TABLE telegram_requests
			DROP COLUMN IF EXISTS owner_admin_id;
		`).Error; err != nil {
			return err
		}

		// --- Rollback Step 7 ---
		if err := tx.Exec(`
			DROP INDEX IF EXISTS idx_telegram_packages_owner_admin_id;
		`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			ALTER TABLE telegram_packages
			DROP COLUMN IF EXISTS owner_admin_id;
		`).Error; err != nil {
			return err
		}

		// --- Rollback Step 6 ---
		if err := tx.Exec(`
			DROP INDEX IF EXISTS idx_ocserv_groups_owner_admin_id;
		`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			ALTER TABLE ocserv_groups
			ADD COLUMN IF NOT EXISTS owner VARCHAR(32) DEFAULT '',
			DROP COLUMN IF EXISTS owner_admin_id;
		`).Error; err != nil {
			return err
		}

		// --- Rollback Step 5 ---
		if err := tx.Exec(`
			DROP INDEX IF EXISTS idx_ocserv_users_owner_admin_id;
		`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			ALTER TABLE ocserv_users
			ADD COLUMN IF NOT EXISTS owner VARCHAR(16) DEFAULT '',
			DROP COLUMN IF EXISTS owner_admin_id;
		`).Error; err != nil {
			return err
		}

		// --- Rollback Step 4 ---
		if err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS users (
				id BIGSERIAL PRIMARY KEY,
				uid VARCHAR(26) NOT NULL UNIQUE,
				username VARCHAR(16) NOT NULL UNIQUE,
				password VARCHAR(64) NOT NULL,
				is_admin BOOLEAN DEFAULT FALSE,
				salt VARCHAR(8) NOT NULL,
				last_login TIMESTAMP NULL,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
		`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS user_tokens (
				id BIGSERIAL PRIMARY KEY,
				user_id BIGINT NOT NULL,
				uid TEXT NOT NULL UNIQUE,
				token TEXT,
				created_at TIMESTAMP,
				expire_at TIMESTAMP,
				CONSTRAINT fk_user_tokens_user
					FOREIGN KEY(user_id)
					REFERENCES users(id)
					ON DELETE CASCADE
			);
		`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_user_tokens_user_id ON user_tokens(user_id);`).Error; err != nil {
			return err
		}

		// --- Rollback Step 3 ---
		if err := tx.Exec(`
			INSERT INTO users (id, username, password, is_admin, created_at, updated_at)
			SELECT id, username, password, CASE WHEN role = 'super' THEN TRUE ELSE FALSE END, created_at, updated_at
			FROM administrators
			ON CONFLICT DO NOTHING;
		`).Error; err != nil {
			return err
		}

		// --- Rollback Step 2 ---
		if err := tx.Exec(`DROP TABLE IF EXISTS administrators;`).Error; err != nil {
			return err
		}

		// --- Rollback Step 1 ---
		if err := tx.Exec(`
			ALTER TABLE ocserv_users
			ADD COLUMN IF NOT EXISTS uid CHAR(26) NOT NULL;
		`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			CREATE UNIQUE INDEX IF NOT EXISTS idx_ocserv_users_uid ON ocserv_users(uid);
		`).Error; err != nil {
			return err
		}

		return nil
	},
}
