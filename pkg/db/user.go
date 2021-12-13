package db

import (
	"fmt"
	"database/sql"

	log "github.com/sirupsen/logrus"

	"github.com/platform9/fast-path/pkg/objects"
)


// AddUser adds user to database
func (q *Querier) AddUser(user *objects.User) error {
	tx, err := q.handle.Begin()
	if err != nil {
		return err
	}

	stmtIns, err := tx.Prepare("INSERT INTO users(name, email, space) values(?, ?, ?)")

	if err != nil {
		return err
	}

	defer stmtIns.Close()

	if _, err = stmtIns.Exec(user.Name, user.Email, user.Space); err != nil {
		log.Error(err, ": Error inserting ", user.Name)
		return err
	}

	return tx.Commit()
}

// RemoveUser removes user from database based on email
func (q *Querier) RemoveUserByEmail(user *objects.User) error {
	tx, err := q.handle.Begin()
	if err != nil {
		return err
	}

	stmtIns, err := tx.Prepare("DELETE FROM users WHERE email=?")

	if err != nil {
		fmt.Println(err)
		return err
	}

	defer stmtIns.Close()
	if _, err = stmtIns.Exec(user.Email); err != nil {
		log.Error(err, ": Error deleting ", user.Email)
		return err
	}

	return tx.Commit()
}

// RemoveUser removes user from database based on name
func (q *Querier) RemoveUserByName(user *objects.User) error {
        tx, err := q.handle.Begin()
        if err != nil {
                return err
        }

        stmtIns, err := tx.Prepare("DELETE FROM users WHERE name=?")

        if err != nil {
                fmt.Println(err)
                return err
        }

        defer stmtIns.Close()
        if _, err = stmtIns.Exec(user.Name); err != nil {
                log.Error(err, ": Error deleting ", user.Name)
                return err
        }

        return tx.Commit()
}


func NullStrToStr(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	} else {
		return ""
	}
}

// GetUsers returns a list of users from database
func (q *Querier) GetUsers(users *[]objects.User) error {
	tx, err := q.handle.Begin()
	if err != nil {
		return err
	}

	rows, err := tx.Query("SELECT id, name, email, space FROM users")

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, email, space sql.NullString
		var id int
		if err = rows.Scan(&id, &name, &email, &space); err != nil {
			return err
		}
		*users = append(*users, objects.User{
			ID:       id,
			Name:     NullStrToStr(name),
			Email:    NullStrToStr(email),
			Space:    NullStrToStr(space),
		})
	}

	return tx.Commit()
}


// GetUserByName returns a user given userName
func (q *Querier) GetUserByName(userName string, user *objects.User) error {
	tx, err := q.handle.Begin()
	if err != nil {
		return err
	}

	rows, err := tx.Query("SELECT id, email, space FROM users WHERE name=?", userName)

	if err != nil {
		return err
	}

	found := false
	if rows.Next() {
		var email sql.NullString
		var space sql.NullString
		var id int
		err = rows.Scan(&id, &email, &space)

		if err != nil {
			tx.Rollback()
			return err
		}

		found = true
		*user = objects.User{
			ID:      id,
			Name:    userName,
			Email:   NullStrToStr(email),
			Space:   NullStrToStr(space),
		}
	}

	if !found {
		tx.Rollback()
		return nil
	}

	rows.Close()

	return tx.Commit()
}


// GetUserByEmail returns a user given userEmail
func (q *Querier) GetUserByEmail(userEmail string, user *objects.User) error {
        tx, err := q.handle.Begin()
        if err != nil {
                return err
        }

        rows, err := tx.Query("SELECT id, name, space FROM users WHERE name=?", userEmail)

        if err != nil {
                return err
        }

        found := false
        if rows.Next() {
                var name sql.NullString
                var space sql.NullString
                var id int
                err = rows.Scan(&id, &name, &space)

                if err != nil {
                        tx.Rollback()
                        return err
                }

                found = true
                *user = objects.User{
                        ID:      id,
                        Name:    NullStrToStr(name),
                        Email:   userEmail,
                        Space:   NullStrToStr(space),
                }
        }

        if !found {
                tx.Rollback()
                return nil
        }

        rows.Close()

        return tx.Commit()
}
