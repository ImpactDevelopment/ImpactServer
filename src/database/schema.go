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
		CREATE TABLE IF NOT EXISTS users (
			user_id  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
			email TEXT UNIQUE,
			password_hash TEXT,
			created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT, /* unix seconds */

			mc_uuid UUID UNIQUE,
			legacy_premium BOOL NOT NULL DEFAULT TRUE, /* list this mc uuid in the premium list for 4.7 and below */

			discord_id TEXT UNIQUE,

			pepsi BOOL NOT NULL DEFAULT FALSE,
			staff BOOL NOT NULL DEFAULT FALSE,
			developer BOOL NOT NULL DEFAULT FALSE
		);
	`)
	if err != nil {
		log.Println("Unable to create users table")
		return err
	}

	return nil
}
