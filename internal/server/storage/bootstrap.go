package storage

import (
	"database/sql"
	"log"
)

// Bootstrap creates database tables if they are not created yet
func Bootstrap(db *sql.DB) {
	createUserTable(db)
	createPasswordsTable(db)
	createTextsTable(db)
	createCardsTable(db)
	createBinTable(db)
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
		`create table if not exists password_pairs (
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

func createTextsTable(db *sql.DB) {
	if _, passwordsTableErr := db.Exec(
		`create table if not exists texts (
    		id integer primary key autoincrement,
    		user_id integer not null,
    		name text not null,
    		text text not null,
    		description text,
    		foreign key(user_id) references users(id)
		)`,
	); passwordsTableErr != nil {
		log.Fatalln(passwordsTableErr)
	}
}

func createCardsTable(db *sql.DB) {
	if _, cardsTableErr := db.Exec(
		`create table if not exists cards(
    		id integer primary key autoincrement,
    		user_id integer not null,
    		name text not null,
    		number text not null,
    		valid_through_month integer not null,
    		valid_through_year integer not null,
    		cvv integer not null,
    		description text,
            foreign key(user_id) references users(id)
		)`,
	); cardsTableErr != nil {
		log.Fatalln(cardsTableErr)
	}
}

func createBinTable(db *sql.DB) {
	if _, createBinTableErr := db.Exec(
		`create table if not exists bins(
    		id integer primary key autoincrement,
    		user_id integer not null,
    		name text not null,
    		data blob not null,
    		description text,
    		foreign key(user_id) references users(id)
		)`,
	); createBinTableErr != nil {
		log.Fatalln(createBinTableErr)
	}
}
