package app

import (
	"context"
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"strconv"
)

var errImageNotFound = errors.New("image not found")
var errItemNotFound = errors.New("item not found")

type Item struct {
	ID       int    `db:"id" json:"-"`
	Name     string `db:"name" json:"name"`
	Category string `db:"category" json:"category"`
	Image    string `db:"image_name" json:"image"`
}

// Please run `go generate ./...` to generate the mock implementation
// ItemRepository is an interface to manage items.
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
	GetAll(ctx context.Context) ([]*Item, error)
	GetByID(ctx context.Context, id string) (*Item, error)
}

// itemRepository is an implementation of ItemRepository
type itemRepository struct {
	dbPath string
}

// NewItemRepository creates a new itemRepository.
func NewItemRepository() ItemRepository {
	db, err := sql.Open("sqlite3", "db/mercari.sqlite3")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	sql, err := os.ReadFile("db/items.sql")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(string(sql))
	if err != nil {
		panic(err)
	}
	return &itemRepository{dbPath: "db/mercari.sqlite3"}
}

type ItemsWrapper struct {
	Items []Item `json:"items"`
}

// Insert inserts an item into the repository.
func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	db, err := sql.Open("sqlite3", i.dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(item.Name, item.Category, item.Image)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (i *itemRepository) GetAll(ctx context.Context) ([]*Item, error) {
	db, err := sql.Open("sqlite3", i.dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, "SELECT id, name, category, image_name FROM items")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		item := &Item{}
		if err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Image); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (i *itemRepository) GetByID(ctx context.Context, id string) (*Item, error) {
	db, err := sql.Open("sqlite3", i.dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	itemID, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	item := &Item{}
	err = db.QueryRowContext(ctx, "SELECT id, name, category, image_name FROM items WHERE id = ?", itemID).Scan(
		&item.ID, &item.Name, &item.Category, &item.Image)
	if err == sql.ErrNoRows {
		return nil, errItemNotFound
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(fileName string, image []byte) error {
	err := os.WriteFile(fileName, image, 0644)
	if err != nil {
		return err
	}

	return nil
}
