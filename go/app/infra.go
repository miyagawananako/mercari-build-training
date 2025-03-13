package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	GetAll(ctx context.Context) (*ItemsWrapper, error)
	GetByID(ctx context.Context, id string) (*Item, error)
}

// itemRepository is an implementation of ItemRepository
type itemRepository struct {
	db      *sql.DB
	dbPath  string
	sqlPath string
}

func (i *itemRepository) createTables(ctx context.Context) error {
	sql, err := os.ReadFile(i.sqlPath)
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %w", err)
	}

	_, err = i.db.ExecContext(ctx, string(sql))
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	return nil
}

// NewItemRepository creates a new itemRepository.
func NewItemRepository() ItemRepository {
	db, err := sql.Open("sqlite3", "db/mercari.sqlite3")
	if err != nil {
		panic(err)
	}

	repo := &itemRepository{
		db:      db,
		dbPath:  "db/mercari.sqlite3",
		sqlPath: "db/items.sql",
	}
	err = repo.createTables(context.Background())
	if err != nil {
		db.Close()
		panic(err)
	}
	return repo
}

type ItemsWrapper struct {
	Items []*Item `json:"items"`
}

// Insert inserts an item into the repository.
func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var categoryID int
	err = tx.QueryRowContext(ctx, "SELECT id FROM categories WHERE name = ?", item.Category).Scan(&categoryID)
	if err == sql.ErrNoRows {
		res, err := tx.ExecContext(ctx, "INSERT INTO categories (name) VALUES (?)", item.Category)
		if err != nil {
			return err
		}
		categoryID64, err := res.LastInsertId()
		if err != nil {
			return err
		}
		categoryID = int(categoryID64)
	} else if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(item.Name, categoryID, item.Image)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (i *itemRepository) GetAll(ctx context.Context) (*ItemsWrapper, error) {
	rows, err := i.db.QueryContext(ctx, "SELECT items.id, items.name, categories.name, items.image_name FROM items INNER JOIN categories ON items.category_id = categories.id")
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

	return &ItemsWrapper{Items: items}, nil
}

func (i *itemRepository) GetByID(ctx context.Context, id string) (*Item, error) {
	itemID, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	item := &Item{}
	err = i.db.QueryRowContext(ctx, "SELECT items.id, items.name, categories.name, items.image_name FROM items INNER JOIN categories ON items.category_id = categories.id WHERE items.id = ?", itemID).Scan(
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
