package server

import (
	"database/sql"
	"log"
)

func bootstrap(db *sql.DB) {
	if _, userTableErr := db.Exec(
		`create table if not exists users (
    		id integer primary key autoincrement,
    		login varchar(255) unique,
    		password varchar(255)
		)`,
	); userTableErr != nil {
		log.Fatalln(userTableErr)
	}
}
