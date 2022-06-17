package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
)

type PgStorage struct {
	conn *pgx.Conn
}

var (
	ErrNotFound = errors.New("no data found")
)

func (db *PgStorage) Close(ctx context.Context) error {
	return db.conn.Close(ctx)
}

func NewPgStorage(databaseURL string) (*PgStorage, error) {
	conn, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping to database: %v", err)
	}

	return &PgStorage{
		conn: conn,
	}, nil
}

func (db *PgStorage) CreateLink(ctx context.Context, link Link) error {
	_, err := db.conn.Exec(ctx, "insert into links (employee_id, external_service_id, external_service_type_id) values ($1, $2, $3)",
		link.EmployeeId, link.ExternalServiceId, link.ExternalServiceTypeId)
	return err
}

func (db *PgStorage) GetEmployeeId(ctx context.Context, externalServiceId string, externalServiceType ExternalServiceType) (string, error) {
	var employeeId string

	row := db.conn.QueryRow(ctx, "select employee_id from links where external_service_id = $1 and external_service_type_id = $2",
		externalServiceId, externalServiceType)

	err := row.Scan(&employeeId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", ErrNotFound
		}

		return "", err
	}

	return employeeId, nil
}

func (db *PgStorage) CreateLog(ctx context.Context, externalServiceId ExternalServiceType, status LoginStatus) error {
	_, err := db.conn.Exec(ctx, "insert into logs (external_service_type_id, status) values ($1, $2)",
		externalServiceId, status)
	return err
}