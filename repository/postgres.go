package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	pg_conf "github.com/joegasewicz/pg-conf"
)

type ToDo struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type DB struct {
	db *sql.DB
}

func New() (*DB, error) {
	conf := pg_conf.PostgresConfig{
		PGHost:     os.Getenv("PG_HOST"),
		PGUser:     os.Getenv("PG_USER"),
		PGPort:     os.Getenv("PG_PORT"),
		PGPassword: os.Getenv("PG_PASS"),
		PGSSLMode:  "disable",
	}
	var err error
	db, err := sql.Open("postgres", conf.GetPostgresConnStr())
	if err != nil {
		return nil, fmt.Errorf("can't open db: %w", err)
	}
	pErr := db.Ping()
	if pErr != nil {
		log.Fatal(pErr)
	}
	_, err = db.Exec("DROP DATABASE IF EXISTS todo_list")
	if err != nil {
		return nil, fmt.Errorf("couldn't drop a database: %w", err)
	}
	_, err = db.Exec("CREATE DATABASE todo_list")
	if err != nil {
		return nil, fmt.Errorf("couldn't create a database: %w", err)
	}
	conf.PGDatabase = "todo_list"
	newDB, err := sql.Open("postgres", conf.GetPostgresConnStr())
	if err != nil {
		return nil, fmt.Errorf("can't open db: %w", err)
	}
	if err := newDB.Ping(); err != nil {
		newDB.Close() // Закрываем соединение, если произошла ошибка
		return nil, fmt.Errorf("can't ping new db: %w", err)
	}
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS items (
		ID SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL
	);
`
	_, err = newDB.Exec(createTableQuery)
	if err != nil {
		newDB.Close()
		return nil, fmt.Errorf("can't create table: %w", err)
	}

	return &DB{db: newDB}, nil
}

func (pg *DB) CreateToDo(name string) error {
	_, err := pg.db.Exec("INSERT INTO items (name) VALUES ($1)", name)
	if err != nil {
		return fmt.Errorf("failed to create todo: %w", err)
	}
	return nil
}

func (pg *DB) UpdateTodo(id int, name string) error {
	_, err := pg.db.Exec("UPDATE items SET name = $1 WHERE id = $2", name, id)
	if err != nil {
		return fmt.Errorf("failed to update todo: %w", err)
	}
	return nil
}
func (pg *DB) DeleteTodo(id int) error {
	_, err := pg.db.Exec("DELETE FROM items WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}
	return nil
}
func (pg *DB) ListToDo() ([]ToDo, error) {
	var todos []ToDo

	rows, err := pg.db.Query("SELECT * FROM items;")
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var todo ToDo

		if err := rows.Scan(&todo.Id, &todo.Name); err != nil {
			return nil, fmt.Errorf("failed to parse rows %w", err)
		}
		todos = append(todos, todo)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("data incomlete: %w", err)
	}
	return todos, nil
}
