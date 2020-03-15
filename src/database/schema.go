package database

import (
	"log"
)

func initialSetup() {
	err := createTables()
	if err != nil {
		panic(err)
	}
}

func createTables() error {
	_, err := DB.Exec(`
		CREATE EXTENSION IF NOT EXISTS "pgcrypto";
	`)
	if err != nil {
		log.Println("Unable to load pgcrypto extension")
		return err
	}

	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS pending_donations (
			token  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT, -- unix seconds
			paypal_order_id TEXT UNIQUE, -- can be null in case we want to make a "gift card" with no paypal order id attached
			paypal_payer_id TEXT,
			paypal_payer_email TEXT,
			amount INTEGER, -- can be null for the same reason

			used BOOL NOT NULL DEFAULT FALSE,
			used_by UUID REFERENCES users(user_id),
			log_msg_id TEXT
		);
	`)
	if err != nil {
		log.Println("Unable to create pending_donations table")
		return err
	}

	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS password_resets (
			token  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(user_id) ,
			created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT -- unix seconds
		);
	`)
	if err != nil {
		log.Println("Unable to create password_resets table")
		return err
	}

	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			user_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email TEXT UNIQUE,
			password_hash TEXT,
			created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT, -- unix seconds

			mc_uuid UUID UNIQUE,
			discord_id TEXT UNIQUE,

			legacy_enabled BOOL NOT NULL DEFAULT FALSE, -- list this mc uuid in the premium list for 4.7 and below. this determines if you get a cape shown to other users who are using 4.7-
			cape_enabled BOOL NOT NULL DEFAULT TRUE, -- show a cape to others on 4.8+

			legacy BOOL NOT NULL DEFAULT TRUE, -- this defaults to true e.g. for manual inserts. register.go overrides this to false however!
			premium BOOL NOT NULL DEFAULT TRUE,
			pepsi BOOL NOT NULL DEFAULT FALSE,
			spawnmason BOOL NOT NULL DEFAULT FALSE,
			staff BOOL NOT NULL DEFAULT FALSE,
			developer BOOL NOT NULL DEFAULT FALSE
		);
	`)
	if err != nil {
		log.Println("Unable to create users table")
		return err
	}

	// A view allows us to control logical column order
	_, err = DB.Exec(`
		DROP VIEW IF EXISTS users_view;

		CREATE VIEW users_view AS SELECT
			user_id,
			email,
			mc_uuid,
			discord_id,
			password_hash,
			cape_enabled, --TODO invert this to "incognito"
			legacy_enabled,
			legacy,
			STRING_TO_ARRAY(
				CONCAT_WS(',',
					CASE WHEN premium THEN 'premium' END,
					CASE WHEN pepsi THEN 'pepsi' END,
					CASE WHEN spawnmason THEN 'spawnmason' END,
					CASE WHEN staff THEN 'staff' END,
					CASE WHEN developer THEN 'developer' END
				),
				','
			) AS roles
			FROM users;
	`)
	if err != nil {
		log.Println("Unable to create users_view view")
		return err
	}

	_, err = DB.Exec(`
		CREATE OR REPLACE FUNCTION notify_users_updated()
		  RETURNS trigger AS $$
		DECLARE
		BEGIN
		  PERFORM pg_notify('users_updated', '');
		  RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
	`)
	if err != nil {
		log.Println("Unable to create notify_users_updated trigger function")
	}

	_, err = DB.Exec(`
		DROP TRIGGER IF EXISTS users_update_trigger ON users;

		CREATE TRIGGER users_update_trigger
		AFTER INSERT OR UPDATE OR DELETE ON users
		FOR EACH STATEMENT
		EXECUTE PROCEDURE notify_users_updated();
	`)
	if err != nil {
		log.Println("Unable to create users_update_trigger trigger")
	}

	return nil
}
