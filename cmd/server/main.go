package main

import (
	"context"
	"github.com/EgorKo25/DevOps-Track-Yandex/internal/file"
	"log"
	"net/http"

	"github.com/EgorKo25/DevOps-Track-Yandex/internal/configuration"
	"github.com/EgorKo25/DevOps-Track-Yandex/internal/database"
	"github.com/EgorKo25/DevOps-Track-Yandex/internal/hashing"
	"github.com/EgorKo25/DevOps-Track-Yandex/internal/middleware"
	"github.com/EgorKo25/DevOps-Track-Yandex/internal/server/handlers"
	"github.com/EgorKo25/DevOps-Track-Yandex/internal/server/routers"
	"github.com/EgorKo25/DevOps-Track-Yandex/internal/storage"
)

func main() {

	ctx := context.Background()

	cfg := config.NewServerConfig()

	str := storage.NewStorage()

	hsr := hashing.MewHash(cfg.Key)

	db := database.NewDB(cfg, ctx, str)

	if db != nil {
		db.CreateTable()
	}

	compressor := middleware.NewCompressor()

	handler := handlers.NewHandler(str, compressor, hsr, db, ctx)

	router := routers.NewRouter(handler)

	save := file.NewSave(cfg, str)

	read, _ := file.NewRead(cfg, str)

	if cfg.Restore {
		read.ReadAll()
	}

	go save.Run()
	log.Println(http.ListenAndServe(cfg.Address, router))

}
