package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/GeorgeShibanin/InternWB/internal/storage"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"log"
)

const (
	dsnTemplate       = "postgres://%s:%s@%s:%v/%s"
	GetOrderByIDQuery = `SELECT attrs FROM items WHERE id = $1`
	InsertOrderQuery  = `INSERT INTO items (id, attrs) values ($1, $2)`
	GetAllData        = `SELECT * FROM items`
)

type StoragePostgres struct {
	conn     *pgx.Conn
	inMemory map[string]storage.Orders
}

func initConnection(conn *pgx.Conn) *StoragePostgres {
	s := &StoragePostgres{}
	s.conn = conn
	err := errors.New("data to cache")
	s.inMemory, err = s.GetAllFromDB()
	if err != nil {
		log.Fatalf("cant get data from Db to inMemory")
	}
	return s
}

func Init(ctx context.Context, host, user, db, password string, port uint16) (*StoragePostgres, error) {
	//подключение к базе через переменные окружения
	conn, err := pgx.Connect(ctx, fmt.Sprintf(dsnTemplate, user, password, host, port, db))
	if err != nil {
		return nil, errors.Wrap(err, "can't connect to postgres")
	}
	return initConnection(conn), nil
}

func (s *StoragePostgres) PutData(model storage.Orders) (storage.Id, error) {
	ctx := context.Background()
	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", errors.Wrap(err, "can't create tx")
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()
	order := &storage.Orders{}

	modelJson, err := json.Marshal(model)
	if err != nil {
		log.Println("can't marshall", err)
		return "", errors.Wrap(err, "can't marshall")
	}

	tag, err := tx.Exec(ctx, InsertOrderQuery, model.OrderUID, modelJson)
	log.Println("put data to postgres")
	if err != nil {
		log.Println("can't insert order", err)
		return "", errors.Wrap(err, "can't insert order")
	}

	if tag.RowsAffected() != 1 {
		log.Println("unexpected rows affected value:", model)
		return "", errors.Wrap(err, fmt.Sprintf("unexpected rows affected value: %v", tag.RowsAffected()))
	}
	log.Println("put order to cache", model)
	s.inMemory[order.OrderUID] = model
	return storage.Id(order.OrderUID), nil
}

func (s *StoragePostgres) GetData(ctx context.Context, id storage.Id) (storage.Orders, error) {
	order := &storage.Orders{}
	dataCache, ok := s.inMemory[string(id)]
	if ok {
		log.Println("get data from Cache")
		return dataCache, nil
	}
	err := s.conn.QueryRow(ctx, GetOrderByIDQuery, string(id)).
		Scan(&order)
	if err != nil {
		return storage.Orders{}, fmt.Errorf("something went wrong - %w", err)
	}
	return *order, nil
}

func (s *StoragePostgres) GetAllFromDB() (map[string]storage.Orders, error) {
	ctx := context.Background()
	result := make(map[string]storage.Orders)

	rows, err := s.conn.Query(ctx, GetAllData)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		order := &storage.Orders{}
		if err = rows.Scan(&id, &order); err != nil {
			log.Fatalf("CANT SCAN ROWS")
		}
		result[id] = *order
	}
	log.Println("put all data to cache")
	return result, err
}
