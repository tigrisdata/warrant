package authz

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type Model interface {
	GetID() int64
	GetObjectId() int64
	GetFeatureId() string
	GetName() database.NullString
	SetName(database.NullString)
	GetDescription() database.NullString
	SetDescription(database.NullString)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() database.NullTime
	ToFeatureSpec() *FeatureSpec
}

type Feature struct {
	ID          int64               `mysql:"id" postgres:"id" sqlite:"id"`
	ObjectId    int64               `mysql:"objectId" postgres:"object_id" sqlite:"objectId"`
	FeatureId   string              `mysql:"featureId" postgres:"feature_id" sqlite:"featureId"`
	Name        database.NullString `mysql:"name" postgres:"name" sqlite:"name"`
	Description database.NullString `mysql:"description" postgres:"description" sqlite:"description"`
	CreatedAt   time.Time           `mysql:"createdAt" postgres:"created_at" sqlite:"createdAt"`
	UpdatedAt   time.Time           `mysql:"updatedAt" postgres:"updated_at" sqlite:"updatedAt"`
	DeletedAt   database.NullTime   `mysql:"deletedAt" postgres:"deleted_at" sqlite:"deletedAt"`
}

func (feature Feature) GetID() int64 {
	return feature.ID
}

func (feature Feature) GetObjectId() int64 {
	return feature.ObjectId
}

func (feature Feature) GetFeatureId() string {
	return feature.FeatureId
}

func (feature Feature) GetName() database.NullString {
	return feature.Name
}

func (feature *Feature) SetName(newName database.NullString) {
	feature.Name = newName
}

func (feature Feature) GetDescription() database.NullString {
	return feature.Description
}

func (feature *Feature) SetDescription(newDescription database.NullString) {
	feature.Description = newDescription
}

func (feature Feature) GetCreatedAt() time.Time {
	return feature.CreatedAt
}

func (feature Feature) GetUpdatedAt() time.Time {
	return feature.UpdatedAt
}

func (feature Feature) GetDeletedAt() database.NullTime {
	return feature.DeletedAt
}

func (feature Feature) ToFeatureSpec() *FeatureSpec {
	return &FeatureSpec{
		FeatureId:   feature.FeatureId,
		Name:        feature.Name,
		Description: feature.Description,
		CreatedAt:   feature.CreatedAt,
	}
}
