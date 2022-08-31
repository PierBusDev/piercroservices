package data

import (
	"context"
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

type Models struct {
	User User
}

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	Password  string    `json:"-"`
	Active    int       `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

const timeout = time.Second * 4

var db *sql.DB

// New creates and returns an instance of type Models containing all the subtype needed for the application.
func New(dbPool *sql.DB) Models {
	db = dbPool
	return Models{
		User: User{},
	}
}

//GetAll returns a slice of all the users in the database, sorted by last name
func (u *User) GetAll() ([]*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	query := `select id, email, first_name, last_name, password, user_active, created_at, updated_at from users order by last_name`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Password, &user.Active, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Println("Error while trying to scan a row in GetAll(): ", err)
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

//GetByEmail returns a single user based on the email passed as a parameter
func (u *User) GetByEmail(email string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	query := `select id, email, first_name, last_name, password, user_active, created_at, updated_at from users where email = $1`
	var user User
	row := db.QueryRowContext(ctx, query, email)

	err := row.Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Password, &user.Active, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

//GetByID returns a single user based on the id passed as a parameter
func (u *User) GetByID(id int) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	query := `select id, email, first_name, last_name, password, user_active, created_at, updated_at from users where id = $1`
	var user User
	row := db.QueryRowContext(ctx, query, id)

	err := row.Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Password, &user.Active, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

//Update updates a single user using the information stored in the receiver u
func (u *User) Update() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	query := `update users set email = $1, first_name = $2, last_name = $3, user_active = $4, updated_at = $5 where id = $6` //note not updating password
	_, err := db.ExecContext(ctx, query, u.Email, u.FirstName, u.LastName, u.Active, time.Now(), u.ID)
	if err != nil {
		return err
	}

	return nil
}

//Delete deletes one user based on the id stored in the receiver u
func (u *User) Delete() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	query := `delete from users where id = $1`
	_, err := db.ExecContext(ctx, query, u.ID)
	if err != nil {
		return err
	}

	return nil
}

//DeleteById delete one user based on the id passed as a parameter
func (u *User) DeleteById(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	query := `delete from users where id = $1`
	_, err := db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

//Insert inserts a new user into the db based in the User object passed as a parameter
//returns the id of the new user
func (u *User) Insert(user User) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	query := `insert into users (email, first_name, last_name, password, user_active, created_at, updated_at) values ($1, $2, $3, $4, $5, $6, $7) returning id`
	var id int
	err = db.QueryRowContext(ctx, query, user.Email, user.FirstName, user.LastName, hashedPwd, user.Active, time.Now(), time.Now()).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

//ResetPassword is used to reset the password of a user
func (u *User) ResetPassword(newPwd string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	query := `update users set password = $1, updated_at = $2 where id = $3`
	_, err = db.ExecContext(ctx, query, hashedPwd, time.Now(), u.ID)
	if err != nil {
		return err
	}

	return nil
}

//PasswordMatches uses bcrypt to compare user password passed as parameter to the hash stored in the db
//if they match it returns true, else it returns false
func (u *User) PasswordMatches(password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil //password is incorrect
		default:
			return false, err //unknown error
		}
	}

	return true, nil
}
