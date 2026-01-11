package store

import (
	"database/sql"
	"fmt"
	"io/fs"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
)

func Open()(*sql.DB,error){
 db,err:=sql.Open("pgx","host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable")
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



