package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"go-login-auth/models"
	"os"
)

var (
	DB    *sql.DB
	dbErr error
)

func ConnectPSQL() error {
	host := os.Getenv("POSTGRESQL_HOST")
	port := os.Getenv("POSTGRESQL_PORT")
	user := os.Getenv("POSTGRESQL_USER")
	password := os.Getenv("POSTGRESQL_PASS")
	dbName := os.Getenv("POSTGRESQL_DBNAME")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable", host, port, user, password, dbName)

	DB, dbErr = sql.Open("postgres", psqlInfo)
	if dbErr != nil {
		return dbErr
	}
	return nil
}

func CheckUser(email, passwordHash string) uint {
	var id uint = 0
	DB.QueryRow("SELECT id FROM users WHERE email = $1 AND password = $2", email, passwordHash).Scan(&id)
	return id
}

func GetUserById(id uint) (models.User, error) {
	stmt, err := DB.Prepare("SELECT name, surname, email FROM users WHERE id = $1")
	if err != nil {
		fmt.Println(err)
		return models.User{}, err
	}
	var user models.User
	user.Id = id
	err2 := stmt.QueryRow(id).Scan(&user.Name, &user.Surname, &user.Email)
	if err2 != nil {
		fmt.Println(err2)
		return models.User{}, err2
	} else {
		return user, nil
	}
}
