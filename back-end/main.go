package main

import (
	"database/sql"
	"html/template"
	"io"
	"log"
	"net/http"

	_ "modernc.org/sqlite" // Используем modernc.org/sqlite
	"github.com/labstack/echo/v4"
)

// Глобальная переменная для базы данных
var db *sql.DB

// Шаблонизатор
type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// Инициализация базы данных
func initDB() {
	var err error
	db, err = sql.Open("sqlite", "./library.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Создание таблицы книг, если её нет
	query := `
	CREATE TABLE IF NOT EXISTS books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		author TEXT NOT NULL,
		genre TEXT NOT NULL,
		is_borrowed BOOLEAN DEFAULT 0,
		borrower TEXT,
		due_date TEXT
	);`
	_, err = db.Exec(query)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	log.Println("Database initialized.")
}

func main() {
	// Инициализация базы данных
	initDB()
	defer db.Close()

	// Создание Echo instance
	e := echo.New()
	e.Static("/public", "public")

	// Рендеринг HTML
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
	e.Renderer = renderer

	// Маршруты
	e.GET("/home", homeHandler)
	e.POST("/books/add", addBookHandler)
	e.GET("/books", getBooksHandler)

	// Запуск сервера
	e.Logger.Fatal(e.Start(":8080"))
}

// Обработчик для добавления книги
func addBookHandler(c echo.Context) error {
	type BookInput struct {
		Title  string `form:"title"`
		Author string `form:"author"`
		Genre  string `form:"genre"`
	}

	var input BookInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid input",
		})
	}

	query := `INSERT INTO books (title, author, genre) VALUES (?, ?, ?)`
	_, err := db.Exec(query, input.Title, input.Author, input.Genre)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to add book",
		})
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "Book added successfully",
	})
}

// Обработчик для получения списка книг
func getBooksHandler(c echo.Context) error {
    rows, err := db.Query(`SELECT id, title, author, genre, is_borrowed FROM books`)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to fetch books",
        })
    }
    defer rows.Close()

    var books []map[string]interface{}
    for rows.Next() {
        var (
            id         int
            title      string
            author     string
            genre      string
            isBorrowed bool
        )
        if err := rows.Scan(&id, &title, &author, &genre, &isBorrowed); err != nil {
            return c.JSON(http.StatusInternalServerError, map[string]string{
                "error": "Failed to parse books",
            })
        }

        status := "Available"
        if isBorrowed {
            status = "Borrowed"
        }

        books = append(books, map[string]interface{}{
            "id":     id,
            "title":  title,
            "author": author,
            "genre":  genre,
            "status": status,
        })
    }

    return c.JSON(http.StatusOK, books)
}


// Обработчик главной страницы
func homeHandler(c echo.Context) error {
	err := c.Render(http.StatusOK, "main.html", nil)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error rendering template: "+err.Error())
	}
	return nil
}
