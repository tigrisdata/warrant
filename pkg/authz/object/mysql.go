package authz

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

type MySQLRepository struct {
	database.SQLRepository
}

func NewMySQLRepository(db *database.MySQL) MySQLRepository {
	return MySQLRepository{
		database.NewSQLRepository(&db.SQL),
	}
}

func (repo MySQLRepository) Create(ctx context.Context, model Model) (int64, error) {
	result, err := repo.DB.ExecContext(
		ctx,
		`
			INSERT INTO object (
				objectType,
				objectId
			) VALUES (?, ?)
			ON DUPLICATE KEY UPDATE
				createdAt = CURRENT_TIMESTAMP(6),
				deletedAt = NULL
		`,
		model.GetObjectType(),
		model.GetObjectId(),
	)
	if err != nil {
		return 0, errors.Wrap(err, "Unable to create object")
	}

	newObjectId, err := result.LastInsertId()
	if err != nil {
		log.Err(err).Msg("Unable to create object")
		return 0, service.NewInternalError("Unable to create object")
	}

	return newObjectId, nil
}

func (repo MySQLRepository) GetById(ctx context.Context, id int64) (Model, error) {
	var object Object
	err := repo.DB.GetContext(
		ctx,
		&object,
		`
			SELECT id, objectType, objectId, createdAt, updatedAt, deletedAt
			FROM object
			WHERE
				id = ? AND
				deletedAt IS NULL
		`,
		id,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("Object", id)
		default:
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to get Object %d from mysql", id))
		}
	}

	return &object, nil
}

func (repo MySQLRepository) GetByObjectTypeAndId(ctx context.Context, objectType string, objectId string) (Model, error) {
	var object Object
	err := repo.DB.GetContext(
		ctx,
		&object,
		`
			SELECT id, objectType, objectId, createdAt, updatedAt, deletedAt
			FROM object
			WHERE
				objectType = ? AND
				objectId = ? AND
				deletedAt IS NULL
		`,
		objectType,
		objectId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError(objectType, objectId)
		default:
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to get object %s:%s from mysql", objectType, objectId))
		}
	}

	return &object, nil
}

func (repo MySQLRepository) List(ctx context.Context, filterOptions *FilterOptions, listParams middleware.ListParams) ([]Model, error) {
	models := make([]Model, 0)
	objects := make([]Object, 0)
	query := `
		SELECT id, objectType, objectId, createdAt, updatedAt, deletedAt
		FROM object
		WHERE
			deletedAt IS NULL
	`
	replacements := []interface{}{}

	if filterOptions != nil && filterOptions.ObjectType != "" {
		query = fmt.Sprintf("%s AND objectType = ?", query)
		replacements = append(replacements, filterOptions.ObjectType)
	}

	if listParams.Query != "" {
		searchTermReplacement := fmt.Sprintf("%%%s%%", listParams.Query)
		query = fmt.Sprintf("%s AND (objectType LIKE ? OR objectId LIKE ?)", query)
		replacements = append(replacements, searchTermReplacement, searchTermReplacement)
	}

	if listParams.AfterId != "" {
		if listParams.AfterValue != nil {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s > ? OR (objectId > ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s < ? OR (objectId < ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			}
		} else {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND objectId > ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND objectId < ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	if listParams.BeforeId != "" {
		if listParams.BeforeValue != nil {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s < ? OR (objectId < ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s > ? OR (objectId > ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			}
		} else {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND objectId < ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND objectId > ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	if listParams.SortBy != "objectId" {
		query = fmt.Sprintf("%s ORDER BY %s %s, objectId %s LIMIT ?", query, listParams.SortBy, listParams.SortOrder, listParams.SortOrder)
		replacements = append(replacements, listParams.Limit)
	} else {
		query = fmt.Sprintf("%s ORDER BY objectId %s LIMIT ?", query, listParams.SortOrder)
		replacements = append(replacements, listParams.Limit)
	}

	err := repo.DB.SelectContext(
		ctx,
		&objects,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models, nil
		default:
			return nil, err
		}
	}

	for i := range objects {
		models = append(models, &objects[i])
	}

	return models, nil
}

func (repo MySQLRepository) DeleteByObjectTypeAndId(ctx context.Context, objectType string, objectId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE object
			SET
				deletedAt = ?
			WHERE
				objectType = ? AND
				objectId = ? AND
				deletedAt IS NULL
		`,
		time.Now().UTC(),
		objectType,
		objectId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return service.NewRecordNotFoundError("Object", fmt.Sprintf("%s:%s", objectType, objectId))
		default:
			return err
		}
	}

	return nil
}
