package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/lannisite110/web3infoanddex/backend/internal/config"
	"github.com/lannisite110/web3infoanddex/backend/internal/db"
	"github.com/lannisite110/web3infoanddex/backend/internal/migrate"
)

func main() {
	dsn := strings.Trim(strings.TrimSpace(os.Getenv("MYSQL_DSN")), `"'`)
	if dsn == "" {
		log.Fatal("MYSQL_DSN is required")
	}
	normalized, err := config.NormalizeMySQLDSN(dsn)
	if err != nil {
		log.Fatalf("mysql dsn: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	mysql, err := db.ConnectMySQL(ctx, normalized)
	if err != nil {
		log.Fatalf("mysql connect: %v", err)
	}
	defer mysql.Close()

	if err := migrate.Apply(ctx, mysql); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("migrations applied successfully")
}
