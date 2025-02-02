package main

import (
    "database/sql"
    "html/template"
    "io"
    "log"
    "net/http"
    "path/filepath"
    _ "modernc.org/sqlite"
    "github.com/labstack/echo/v4"
)

var db *sql.DB


type TemplateRenderer struct {
    templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
    return t.templates.ExecuteTemplate(w, name, data)
}

func initDB() {
    path := "" //ваша база данных
    absPath, err := filepath.Abs(path)
    if err != nil {
        log.Fatalf("Ошибка при получении пути к базе: %v", err)
    }
    log.Println("Используется база данных:", absPath)

    db, err = sql.Open("sqlite", absPath)
    if err != nil {
        log.Fatalf("Не удалось подключиться к базе данных: %v", err)
    }

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
        log.Fatalf("Не удалось создать таблицу: %v", err)
    }

    log.Println("База данных инициализирована.")
}

func main() {

    initDB()

    if db != nil {
        defer db.Close()
    }

    e := echo.New()
    e.Static("/public", "public")

    renderer := &TemplateRenderer{
        templates: template.Must(template.ParseGlob("")), //Расположение вашего html файла
    }
    e.Renderer = renderer


    e.GET("/home", homeHandler)
    e.POST("/books/add", addBookHandler)
    e.GET("/books", getBooksHandler)
    e.DELETE("/books/delete/:id", deleteBookHandler)


    e.Logger.Fatal(e.Start(":8080"))
}


func homeHandler(c echo.Context) error {
    return c.Render(http.StatusOK, "main.html", nil)
}


func addBookHandler(c echo.Context) error {
    var book struct {
        Title  string `json:"title"`
        Author string `json:"author"`
        Genre  string `json:"genre"`
    }

    if err := c.Bind(&book); err != nil {
        log.Println("Ошибка при связывании данных:", err)
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат запроса"})
    }

    query := `INSERT INTO books (title, author, genre) VALUES (?, ?, ?)`
    _, err := db.Exec(query, book.Title, book.Author, book.Genre)
    if err != nil {
        log.Println("Ошибка при добавлении книги:", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при добавлении книги"})
    }

    return c.JSON(http.StatusCreated, map[string]string{"message": "Книга добавлена успешно"})
}


func getBooksHandler(c echo.Context) error {
    log.Println("Запрос на получение книг")

    rows, err := db.Query(`SELECT id, title, author, genre, is_borrowed FROM books`)
    if err != nil {
        log.Println("Ошибка при получении книг:", err)
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
            log.Println("Ошибка при сканировании строки:", err)
            return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении данных"})
        }
        books = append(books, map[string]interface{}{
            "id":          id,
            "title":       title,
            "author":      author,
            "genre":       genre,
            "is_borrowed": isBorrowed,
        })
    }

    return c.JSON(http.StatusOK, books)
}


func deleteBookHandler(c echo.Context) error {
    bookId := c.Param("id")


    query := `DELETE FROM books WHERE id = ?`
    _, err := db.Exec(query, bookId)
    if err != nil {
        log.Println("Ошибка при удалении книги:", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении книги"})
    }

    return c.JSON(http.StatusOK, map[string]string{"message": "Книга успешно удалена"})
}