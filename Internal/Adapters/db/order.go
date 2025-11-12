package db 

import (
    "database/sql"
    _ "github.com/lib/pq"
    "fmt"
)

func ConnectDB() (*sql.DB, error) {
    connStr := "postgres://postgres;Password=admin123@localhost:5432/AGRITRACEsslmode=disable"
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, fmt.Errorf("gagal membuka koneksi: %v", err)
    }

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("gagal menguji koneksi: %v", err)
    }

    fmt.Println("Berhasil terhubung ke PostgreSQL!")
    return db, nil
}
