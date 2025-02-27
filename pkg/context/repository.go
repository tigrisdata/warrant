package context

import (
	"context"
	"fmt"

	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/service"
)

type ContextRepository interface {
	CreateAll(ctx context.Context, contexts []Model) ([]Model, error)
	ListByWarrantId(ctx context.Context, warrantIds []int64) ([]Model, error)
	DeleteAllByWarrantId(ctx context.Context, warrantId int64) error
}

func NewRepository(db database.Database) (ContextRepository, error) {
	switch db.Type() {
	case database.TypeMySQL:
		mysql, ok := db.(*database.MySQL)
		if !ok {
			return nil, fmt.Errorf("invalid %s database config", database.TypeMySQL)
		}

		return NewMySQLRepository(mysql), nil
	case database.TypePostgres:
		postgres, ok := db.(*database.Postgres)
		if !ok {
			return nil, fmt.Errorf("invalid %s database config", database.TypePostgres)
		}

		return NewPostgresRepository(postgres), nil
	case database.TypeSQLite:
		sqlite, ok := db.(*database.SQLite)
		if !ok {
			return nil, fmt.Errorf("invalid %s database config", database.TypeSQLite)
		}

		return NewSQLiteRepository(sqlite), nil
	default:
		return nil, service.NewInternalError(fmt.Sprintf("Invalid database type %s specified", db.Type()))
	}
}
