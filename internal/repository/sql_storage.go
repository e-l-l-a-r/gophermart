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
	logger.Info("Add user " + user.GetLogin())
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
	err = logger.ExecuteWithRetryNoResult(func(args ...interface{}) error {
		return sqls.db.QueryRowContext(ctx, `
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
		).Scan(&username)
	})
	if err != nil {
		err = logger.NewTracedError("error adding new order: ", err)
		return
	}

	logger.Info("Add order ", orderNum, " to user ", userNm)

	return
}

// GetOrdesList получает список заказов для указанного пользователя с фильтрацией по статусу
//
// Параметры:
// - ctx: Контекст выполнения запроса
// - userNm: Имя пользователя для фильтрации заказов.Если пустая строка, то вернем заказы всех пользователец
// - lim_stt: Статус, по которому бкдем фильтровать заказы. Вернутся только заказы со статусом менее или равным указанному
// - count: Максимальное количество возвращаемых заказов. Если 0, то верутся все
func (sqls *SqlStorage) GetOrdesList(ctx context.Context, userNm string, lim_stt int, count int) (data []OrderData, err error) {
	res, err := logger.ExecuteWithRetry(func(args ...interface{}) (interface{}, error) {
		sql := `
			SELECT 
			    o."Number"
			    , o."Status"
			    , o."Accrual"
			    , o."UploadedAt"
			FROM "Order" o
			JOIN "User" u ON o."UserId" = u."ID"
			WHERE ( 
			    	u."Name" = $1::text
			    	OR $1::text = ''
				)
				AND o."Status" <= $2::int
		`

		// Если запрос по пользователю, то возвращаем заказыначиная с самого свежего,
		// для проверки статусов берем в первую очередь самые старые заказы
		if userNm == "" {
			sql += `
				ORDER BY o."UploadedAt" DESC
			`
		} else {
			sql += `
				ORDER BY o."UploadedAt" ASC
			`
		}

		if count > 0 {
			data = make([]OrderData, 0, count)
			sql += `
				LIMIT ` + fmt.Sprint(count)
		}

		return sqls.db.QueryContext(ctx, sql, userNm, lim_stt)
	})

	if err != nil {
		err = logger.NewTracedError("error adding new order: ", err)
		return
	}
	rows := res.(*sql.Rows)
	defer rows.Close()

	logger.Info("Gef orders with filter: User: ", userNm, " Status: ", lim_stt, " Count: ", count)

	for rows.Next() {
		var order OrderData
		rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.Uploaded)
		data = append(data, order)
		logger.Info("Order: ", order.Number, " Status: ", order.Status, " Accrual: ", order.Accrual)
	}
	return
}

func (sqls *SqlStorage) UpdOrderData(ctx context.Context, orderNum string, orserStt int, accrual float64) error {
	_, err := logger.ExecuteWithRetry(func(args ...interface{}) (interface{}, error) {
		return sqls.db.ExecContext(ctx, `
			WITH updater as(
				UPDATE "Order"
				SET 
				    "Status" = $1::int,
					"Accrual" = $2::float8
				WHERE
				       "Number" = $3::text
				RETURNING 
				       "UserId", "Accrual"
			)
			UPDATE "Balance" b
			SET "Current" = "Current" + u."Accrual"
			FROM updater u
			WHERE b."UserId" = u."UserId"
			   AND u."Accrual" > 0
            `,
			orserStt, accrual, orderNum,
		)
	})
	if err != nil {
		return logger.NewTracedError("Error updating order data: ", err)
	}
	logger.Info("Update order: ", orderNum, " Status: ", orserStt, " Accrual: ", accrual)
	return nil
}
