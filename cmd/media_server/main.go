package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/Benson003/media_server/internal/handlers"
	"github.com/Benson003/media_server/internal/media"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	_ "github.com/mattn/go-sqlite3"
)
const createTableSQL = `
CREATE TABLE IF NOT EXISTS media (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    path TEXT NOT NULL,
    ext TEXT
);`

func main(){
	config ,err := media.LoadConfig("config.json")
	if err != nil{
		log.Printf("Failed to read config \n Reason: %v\n",err)
		return
	}
	media_files,err := media.ScanMediaDirs(*config)
	if err != nil{
		log.Printf("Failed to load files \n Reason: %v\n",err)
		return
	}


	dbConn,err := initDatabase("media.db")
	if err != nil{
		log.Printf("Failed to load database \n Reason: %v\n",err)
		return
	}
	defer dbConn.Close()
	mediaStore := media.NewMediaStore(dbConn)


	ctx := context.Background()
for _, file := range media_files {
    ext := sql.NullString{String: "", Valid: false}
    // Extract extension if you want
    if dot := strings.LastIndex(file.Name, "."); dot != -1 {
        ext.String = file.Name[dot+1:]
        ext.Valid = true
    }

	err := mediaStore.InsertMediaIfNotExists(ctx, file)
	if err != nil {
		log.Printf("Failed to insert media %s: %v", file.Name, err)
	}
}




	r := chi.NewRouter()
	h := handlers.NewHandler(mediaStore)


	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Or specific origins like "http://192.168.0.123"
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           2500, // Max value: 600
	}))

	r.Get("/media",h.HandleGetAllMedia)
	r.Get("/media/{id}/stream", h.HandleStreamMedia)
	r.Get("/media/{id}/info",h.HandleGetMediaInfo)
	r.Get("/media/{id}/thumbnail",h.HandleThumbnail)

	log.Println("Starting server on :8080")
    if err := http.ListenAndServe(":8080", r); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}


func initDatabase(name string)(*sql.DB,error){
	dbConn,err := sql.Open("sqlite3",name)
	if err != nil{
		return nil,err
	}

	if err := dbConn.Ping(); err != nil{
		return nil,err
	}

	// Create table if it doesn't exist
    if _, err := dbConn.Exec(createTableSQL); err != nil {
        return nil, err
    }

	return dbConn,nil
}
