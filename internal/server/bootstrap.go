package server

import (
	"database/sql"
	"log"
)

func bootstrap(db *sql.DB) {
	createUserTable(db)
	createPasswordsTable(db)
}

func createUserTable(db *sql.DB) {
	if _, userTableErr := db.Exec(
		`create table if not exists users (
    		id integer primary key autoincrement,
    		login text unique not null,
    		password text not null
		)`,
	); userTableErr != nil {
		log.Fatalln(userTableErr)
	}
}

func createPasswordsTable(db *sql.DB) {
	if _, passwordsTableErr := db.Exec(
		`create table if not exists passwords (
    		id integer primary key autoincrement,
    		user_id integer not null,
    		login text not null,
    		password text not null,
    		description text,
    		foreign key(user_id) references users(id)
		)`,
	); passwordsTableErr != nil {
		log.Fatalln(passwordsTableErr)
	}
}
