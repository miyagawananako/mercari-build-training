package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"database/sql"
	"github.com/google/go-cmp/cmp"
	gomock "go.uber.org/mock/gomock"
	"mime/multipart"
	"os"
	"strings"
)

func TestParseAddItemRequest(t *testing.T) {
	t.Parallel()

	type wants struct {
		req *AddItemRequest
		err bool
	}

	cases := map[string]struct {
		args map[string]string
		wants
	}{
		"ok: valid request": {
			args: map[string]string{
				"name":     "jacket",
				"category": "fashion",
				"image":    "jacket.jpg",
			},
			wants: wants{
				req: &AddItemRequest{
					Name:     "jacket",
					Category: "fashion",
					Image:    []byte("jacket.jpg"),
				},
				err: false,
			},
		},
		"ng: empty request": {
			args: map[string]string{},
			wants: wants{
				req: nil,
				err: true,
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var b bytes.Buffer
			w := multipart.NewWriter(&b)

			for k, v := range tt.args {
				if k == "image" {
					fw, err := w.CreateFormFile("image", v)
					if err != nil {
						t.Fatal(err)
					}
					fw.Write([]byte(v))
				} else {
					if err := w.WriteField(k, v); err != nil {
						t.Fatal(err)
					}
				}
			}
			w.Close()

			req, err := http.NewRequest("POST", "/items", &b)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", w.FormDataContentType())

			// execute test target
			got, err := parseAddItemRequest(req)

			// confirm the result
			if err != nil {
				if !tt.err {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if diff := cmp.Diff(tt.wants.req, got); diff != "" {
				t.Errorf("unexpected request (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHelloHandler(t *testing.T) {
	t.Parallel()

	type wants struct {
		code int
		body map[string]string
	}
	want := wants{
		code: http.StatusOK,
		body: map[string]string{"message": "Hello, world!"},
	}

	req := httptest.NewRequest("GET", "/hello", nil)
	res := httptest.NewRecorder()

	h := &Handlers{}
	h.Hello(res, req)

	if res.Code != want.code {
		t.Errorf("expected status code %d, got %d", want.code, res.Code)
	}

	var got map[string]string
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if diff := cmp.Diff(want.body, got); diff != "" {
		t.Errorf("unexpected response body (-want +got):\n%s", diff)
	}
}

func TestAddItem(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	type wants struct {
		code int
	}
	cases := map[string]struct {
		args     map[string]string
		injector func(m *MockItemRepository)
		wants
	}{
		"ok: correctly inserted": {
			args: map[string]string{
				"name":     "used iPhone 16e",
				"category": "phone",
				"image":    "test.jpg",
			},
			injector: func(m *MockItemRepository) {
				m.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(nil)
			},
			wants: wants{
				code: http.StatusOK,
			},
		},
		"ng: failed to insert": {
			args: map[string]string{
				"name":     "used iPhone 16e",
				"category": "phone",
				"image":    "test.jpg",
			},
			injector: func(m *MockItemRepository) {
				m.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(errors.New("failed to insert"))
			},
			wants: wants{
				code: http.StatusInternalServerError,
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockIR := NewMockItemRepository(ctrl)
			tt.injector(mockIR)
			h := &Handlers{
				imgDirPath: tmpDir,
				itemRepo:   mockIR,
			}

			var b bytes.Buffer
			w := multipart.NewWriter(&b)

			for k, v := range tt.args {
				if k == "image" {
					fw, err := w.CreateFormFile("image", v)
					if err != nil {
						t.Fatal(err)
					}
					fw.Write([]byte(v))
				} else {
					if err := w.WriteField(k, v); err != nil {
						t.Fatal(err)
					}
				}
			}
			w.Close()

			req := httptest.NewRequest("POST", "/items", &b)
			req.Header.Set("Content-Type", w.FormDataContentType())

			rr := httptest.NewRecorder()
			h.AddItem(rr, req)

			if tt.wants.code != rr.Code {
				t.Errorf("expected status code %d, got %d", tt.wants.code, rr.Code)
			}
			if tt.wants.code >= 400 {
				return
			}

			var resp AddItemResponse
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			expectedMessage := fmt.Sprintf("item received: %s", tt.args["name"])
			if resp.Message != expectedMessage {
				t.Errorf("unexpected message, want %q, got %q", expectedMessage, resp.Message)
			}
		})
	}
}

func TestAddItemE2e(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test")
	}

	db, closers, dbPath, err := setupDB(t)
	if err != nil {
		t.Fatalf("failed to set up database: %v", err)
	}
	t.Cleanup(func() {
		for _, c := range closers {
			c()
		}
	})

	type wants struct {
		code int
	}
	cases := map[string]struct {
		args map[string]string
		wants
	}{
		"ok: correctly inserted": {
			args: map[string]string{
				"name":     "used iPhone 16e",
				"category": "phone",
				"image":    "test.jpg",
			},
			wants: wants{
				code: http.StatusOK,
			},
		},
		"ng: failed to insert": {
			args: map[string]string{
				"name":     "",
				"category": "phone",
				"image":    "test.jpg",
			},
			wants: wants{
				code: http.StatusBadRequest,
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			h := &Handlers{
				imgDirPath: t.TempDir(),
				itemRepo: &itemRepository{
					db:     db,
					dbPath: dbPath,
				},
			}

			var b bytes.Buffer
			w := multipart.NewWriter(&b)

			for k, v := range tt.args {
				if k == "image" {
					fw, err := w.CreateFormFile("image", v)
					if err != nil {
						t.Fatal(err)
					}
					fw.Write([]byte("test image data"))
				} else {
					if err := w.WriteField(k, v); err != nil {
						t.Fatal(err)
					}
				}
			}
			w.Close()

			req := httptest.NewRequest("POST", "/items", &b)
			req.Header.Set("Content-Type", w.FormDataContentType())

			rr := httptest.NewRecorder()
			h.AddItem(rr, req)

			// check response
			if tt.wants.code != rr.Code {
				t.Errorf("expected status code %d, got %d", tt.wants.code, rr.Code)
			}
			if tt.wants.code >= 400 {
				return
			}
			if !strings.Contains(rr.Body.String(), tt.args["name"]) {
				t.Errorf("response body does not contain %s, got: %s", tt.args["name"], rr.Body.String())
			}

			if tt.wants.code == http.StatusOK {
				var categoryID int
				err := db.QueryRow("SELECT id FROM categories WHERE name = ?", tt.args["category"]).Scan(&categoryID)
				if err != nil {
					t.Errorf("failed to get category: %v", err)
				}

				var item struct {
					Name      string
					Category  string
					ImageName string
				}
				err = db.QueryRow(`
					SELECT i.name, c.name, i.image_name
					FROM items i
					JOIN categories c ON i.category_id = c.id
					WHERE i.name = ?`, tt.args["name"]).Scan(&item.Name, &item.Category, &item.ImageName)
				if err != nil {
					t.Errorf("failed to get item: %v", err)
				}

				if item.Name != tt.args["name"] {
					t.Errorf("unexpected name: want %q, got %q", tt.args["name"], item.Name)
				}
				if item.Category != tt.args["category"] {
					t.Errorf("unexpected category: want %q, got %q", tt.args["category"], item.Category)
				}
			}
		})
	}
}

func setupDB(t *testing.T) (db *sql.DB, closers []func(), dbPath string, err error) {
	t.Helper()

	defer func() {
		if err != nil {
			for _, c := range closers {
				c()
			}
		}
	}()

	// create a temporary file for e2e testing
	f, err := os.CreateTemp(".", "*.sqlite3")
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to create temp file: %w", err)
	}
	closers = append(closers, func() {
		f.Close()
		os.Remove(f.Name())
	})

	// set up tables
	db, err = sql.Open("sqlite3", f.Name())
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to open database: %w", err)
	}
	closers = append(closers, func() {
		db.Close()
	})

	cmd := `CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		category_id INTEGER NOT NULL,
		image_name TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	);`

	_, err = db.Exec(cmd)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to create tables: %w", err)
	}

	return db, closers, f.Name(), nil
}
