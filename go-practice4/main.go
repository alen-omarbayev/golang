package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type User struct {
	ID      int     `db:"id"`
	Name    string  `db:"name"`
	Email   string  `db:"email"`
	Balance float64 `db:"balance"`
}

func main() {
	dsn := "postgres://postgres:postgres@127.0.0.1:5433/usersdb?sslmode=disable"
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("connect error: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("ping error: %v", err)
	}
	fmt.Println("connected to postgres")



	

	users, err := GetAllUsers(db)
	if err != nil {
		log.Fatalf("GetAllUsers(before): %v", err)
	}
	fmt.Println("All users BEFORE transfer:")
	printUsers(users)

	u1, err := GetUserByID(db, 1)
	if err != nil {
		log.Fatalf("GetUserByID(1): %v", err)
	}
	fmt.Printf("User #1: %+v\n", u1)

	if err := TransferBalance(db, 1, 2, 200); err != nil {
		log.Fatalf("TransferBalance: %v", err)
	}
	fmt.Println("transfer 200 from #1 to #2 committed")

	users, err = GetAllUsers(db)
	if err != nil {
		log.Fatalf("GetAllUsers(after): %v", err)
	}
	fmt.Println("All users AFTER transfer:")
	printUsers(users)
}

func InsertUser(db *sqlx.DB, user User) error {
	const q = `
		INSERT INTO users (name, email, balance)
		VALUES (:name, :email, :balance)
	`
	_, err := db.NamedExec(q, user)
	return err
}

func GetAllUsers(db *sqlx.DB) ([]User, error) {
	var users []User
	const q = `SELECT id, name, email, balance FROM users ORDER BY id`
	if err := db.Select(&users, q); err != nil {
		return nil, err
	}
	return users, nil
}

func GetUserByID(db *sqlx.DB, id int) (User, error) {
	var u User
	const q = `SELECT id, name, email, balance FROM users WHERE id = $1`
	if err := db.Get(&u, q, id); err != nil {
		return User{}, err
	}
	return u, nil
}

func TransferBalance(db *sqlx.DB, fromID int, toID int, amount float64) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	rollback := func(e error) error {
		_ = tx.Rollback()
		return e
	}

	var sender, receiver User
	if err := tx.Get(&sender, `SELECT id, name, email, balance FROM users WHERE id=$1`, fromID); err != nil {
		return rollback(fmt.Errorf("sender not found: %w", err))
	}
	if err := tx.Get(&receiver, `SELECT id, name, email, balance FROM users WHERE id=$1`, toID); err != nil {
		return rollback(fmt.Errorf("receiver not found: %w", err))
	}

	if sender.Balance < amount {
		return rollback(fmt.Errorf("insufficient funds"))
	}

	if _, err := tx.Exec(`UPDATE users SET balance = balance - $1 WHERE id = $2`, amount, fromID); err != nil {
		return rollback(fmt.Errorf("debit failed: %w", err))
	}
	if _, err := tx.Exec(`UPDATE users SET balance = balance + $1 WHERE id = $2`, amount, toID); err != nil {
		return rollback(fmt.Errorf("credit failed: %w", err))
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}
	return nil
}

func printUsers(users []User) {
	for _, u := range users {
		fmt.Printf("  #%d %-10s | %-22s | balance: %.2f\n", u.ID, u.Name, u.Email, u.Balance)
	}
}
