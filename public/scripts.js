document.getElementById("addBookForm").addEventListener("submit", async function (e) {
    e.preventDefault();

    const formData = new FormData(this);
    const data = {
        title: formData.get("title"),
        author: formData.get("author"),
        genre: formData.get("genre"),
    };

    // Отправка данных о книге на сервер
    const response = await fetch("/books/add", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
    });

    const result = await response.json();

    if (response.ok) {
        // Обновление таблицы книг после успешного добавления
        addBookToTable(data.title, data.author, data.genre, "Available");
        alert(result.message);
    } else {
        alert(result.error);
    }
});

// Загрузка всех книг
async function loadBooks() {
    const response = await fetch("/books");
    const books = await response.json();

    const tbody = document.getElementById("bookTableBody");
    tbody.innerHTML = ""; // Очистка таблицы

    books.forEach(book => {
        addBookToTable(book.title, book.author, book.genre, book.status);
    });
}

// Добавление книги в таблицу
function addBookToTable(title, author, genre, status) {
    const tbody = document.getElementById("bookTableBody");
    const row = `
            <tr>
                <td>${title}</td>
                <td>${author}</td>
                <td>${genre}</td>
                <td>${status}</td>
            </tr>`;
    tbody.innerHTML += row;
}

// Загрузка книг при загрузке страницы
document.addEventListener("DOMContentLoaded", loadBooks);