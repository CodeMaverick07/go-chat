package store

import (
	"database/sql"
	"fmt"
	"go-chat/internals/config"
	"io/fs"
	"log"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
)

func Open()(*sql.DB,error){
	cfg:= config.Load()
	dsn := fmt.Sprintf(
    "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
    cfg.DB_HOST,
    cfg.DB_PORT,
    cfg.DB_USER,
    cfg.DB_PASSWORD,
    cfg.DB_NAME,
)

db, err := sql.Open("pgx", dsn)
if err != nil {
    log.Fatal(err)
}
 if (err!=nil) {
	fmt.Println("error in database connection:",err)
	return nil,err
 }
 fmt.Println("successfully connected to database")
 return db,nil
}

func Migrate(db *sql.DB,dir string) error {
err := goose.SetDialect("postgres")
if (err != nil ){
	return fmt.Errorf("issue with migrations 1 %w", err)
}
err = goose.Up(db,dir)

if (err != nil) {
	return fmt.Errorf("issue with migrations 2 %w", err)
}
return nil
}


func MigrateFS(db *sql.DB, migrationFS fs.FS, dir string) error {
	goose.SetBaseFS(migrationFS)
	defer func() {
		goose.SetBaseFS(nil)
	}()
	return Migrate(db, dir)

}



