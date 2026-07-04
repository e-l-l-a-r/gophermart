package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/e-l-l-a-r/gophermart/internal/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type (
	SqlStorage struct {
		db *sql.DB
	}

	ErrNoData struct {
		logger.TracedError
	}
)

var storage *SqlStorage

func InitSqlStorage(connectionStr string) (*SqlStorage, error) {
	db, err := sql.Open("pgx", connectionStr)
	if err != nil {
		return nil, err
	}

	storage = &SqlStorage{
		db: db,
	}

	return storage, nil
}

func GetSqlStorage() (*SqlStorage, error) {
	if storage == nil {
		return nil, fmt.Errorf("no storage inited")
	}
	return storage, nil
}

func (sqls *SqlStorage) Close() {
	sqls.db.Close()
}

func (sqls *SqlStorage) DoMigrate() error {
	driver, err := postgres.WithInstance(sqls.db, &postgres.Config{})
	if err != nil {
		return logger.NewTracedError("error creating migrations driver: ", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"postgres", driver)
	if err != nil {
		return logger.NewTracedError("error creating migrations: ", err)
	}
	m.Up()
	return nil
}

func (sqls *SqlStorage) AddUser(ctx context.Context, user itfUser) error {

	_, err := logger.ExecuteWithRetry(func(args ...interface{}) (interface{}, error) {
		return sqls.db.ExecContext(ctx, `
			WITH ins_data as(
				INSERT INTO "User" ("Name", "AuthKey")
				VALUES ($1, $2)
				RETURNING "ID"
		    )
			INSERT INTO "Balance" ("UserId", "Current", "Withdrawn") 
			SELECT 
			    "ID", 0, 0 
			FROM ins_data
            `,
			user.GetLogin(), user.GetAuthKey(),
		)
	})
	if err != nil {
		return logger.NewTracedError("Error adding new user: ", err)
	}
	return nil
}

func (sqls *SqlStorage) CreateSession(ctx context.Context, user itfUser) (session string, err error) {
	err = logger.ExecuteWithRetryNoResult(func(args ...interface{}) error {
		return sqls.db.QueryRowContext(ctx, `
			INSERT INTO "Session" ( "UserId" )
	    	SELECT "ID" from "User"
	    	WHERE 
	    	    "Name" = $1
	    	     AND
	    	    "AuthKey" = $2
			RETURNING "SessionKey"
            `,
			user.GetLogin(), user.GetAuthKey(),
		).Scan(&session)
	})
	if err != nil {
		err = logger.NewTracedError("error adding new session: ", err)
		return "", err
	}
	return
}

func (sqls *SqlStorage) GetUserBySession(ctx context.Context, session string) (data UserData, err error) {
	err = logger.ExecuteWithRetryNoResult(func(args ...interface{}) error {
		return sqls.db.QueryRowContext(ctx, `
			WITH session_data as (
			    SELECT "UserId"
			    FROM "Session"
			    WHERE "SessionKey" = $1::uuid
			      AND "ExpiredAt" > now()
			)
			SELECT 
			    u."Name"
				, b."Current"
				, b."Withdrawn"
			FROM session_data sd
			LEFT JOIN "User" u ON sd."UserId" = u."ID"
			LEFT JOIN "Balance" b ON u."ID" = b."UserId"
            `,
			session,
		).Scan(&data.Login, &data.Current, &data.Withdrawn)
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = &ErrNoData{*logger.NewTracedError("No user found: ", err)}
			return
		}
		err = logger.NewTracedError("error searching user data: ", err)
		return
	}

	return
}

func (sqls *SqlStorage) AddNewOrder(ctx context.Context, orderNum string, userNm string) (username string, err error) {
	data, err := logger.ExecuteWithRetry(func(args ...interface{}) (interface{}, error) {
		return sqls.db.QueryContext(ctx, `
			WITH order_check AS (
				SELECT o."ID", u."Name"
				FROM "Order" o
				JOIN "User" u ON o."UserId" = u."ID"
				WHERE o."Number" = $1
			),
			ins_data as (
				INSERT INTO "Order" ("UserId", "Number", "Status", "Accrual", "UploadedAt")
				SELECT u."ID", $1, 0, 0, NOW()
				FROM "User" u
				WHERE u."Name" = $2
				AND NOT EXISTS (SELECT 1 FROM order_check)
				RETURNING *
		    )
			SELECT '' FROM ins_data
			UNION ALL
			SELECT ( SELECT "Name" FROM order_check )
			LIMIT 1
            `,
			orderNum, userNm,
		)
	})
	if err != nil {
		err = logger.NewTracedError("error adding new order: ", err)
		return "", err
	}
	rows := data.(*sql.Rows)
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&username)
	}
	return
}
