package authz

import (
	"context"

	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

const ResourceTypeFeature = "feature"

type FeatureService struct {
	service.BaseService
	Repository FeatureRepository
	EventSvc   event.EventService
	ObjectSvc  object.ObjectService
}

func NewService(env service.Env, repository FeatureRepository, eventSvc event.EventService, objectSvc object.ObjectService) FeatureService {
	return FeatureService{
		BaseService: service.NewBaseService(env),
		Repository:  repository,
		EventSvc:    eventSvc,
		ObjectSvc:   objectSvc,
	}
}

func (svc FeatureService) Create(ctx context.Context, featureSpec FeatureSpec) (*FeatureSpec, error) {
	var newFeature Model
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		createdObject, err := svc.ObjectSvc.Create(txCtx, *featureSpec.ToObjectSpec())
		if err != nil {
			return err
		}

		_, err = svc.Repository.GetByFeatureId(txCtx, featureSpec.FeatureId)
		if err == nil {
			return service.NewDuplicateRecordError("Feature", featureSpec.FeatureId, "A feature with the given featureId already exists")
		}

		newFeatureId, err := svc.Repository.Create(txCtx, featureSpec.ToFeature(createdObject.ID))
		if err != nil {
			return err
		}

		newFeature, err = svc.Repository.GetById(txCtx, newFeatureId)
		if err != nil {
			return err
		}

		err = svc.EventSvc.TrackResourceCreated(ctx, ResourceTypeFeature, newFeature.GetFeatureId(), newFeature.ToFeatureSpec())
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return newFeature.ToFeatureSpec(), nil
}

func (svc FeatureService) GetByFeatureId(ctx context.Context, featureId string) (*FeatureSpec, error) {
	feature, err := svc.Repository.GetByFeatureId(ctx, featureId)
	if err != nil {
		return nil, err
	}

	return feature.ToFeatureSpec(), nil
}

func (svc FeatureService) List(ctx context.Context, listParams middleware.ListParams) ([]FeatureSpec, error) {
	featureSpecs := make([]FeatureSpec, 0)
	features, err := svc.Repository.List(ctx, listParams)
	if err != nil {
		return featureSpecs, nil
	}

	for _, feature := range features {
		featureSpecs = append(featureSpecs, *feature.ToFeatureSpec())
	}

	return featureSpecs, nil
}

func (svc FeatureService) UpdateByFeatureId(ctx context.Context, featureId string, featureSpec UpdateFeatureSpec) (*FeatureSpec, error) {
	currentFeature, err := svc.Repository.GetByFeatureId(ctx, featureId)
	if err != nil {
		return nil, err
	}

	currentFeature.SetName(featureSpec.Name)
	currentFeature.SetDescription(featureSpec.Description)
	err = svc.Repository.UpdateByFeatureId(ctx, featureId, currentFeature)
	if err != nil {
		return nil, err
	}

	updatedFeature, err := svc.Repository.GetByFeatureId(ctx, featureId)
	if err != nil {
		return nil, err
	}

	updatedFeatureSpec := updatedFeature.ToFeatureSpec()
	err = svc.EventSvc.TrackResourceUpdated(ctx, ResourceTypeFeature, updatedFeature.GetFeatureId(), updatedFeatureSpec)
	if err != nil {
		return nil, err
	}

	return updatedFeatureSpec, nil
}

func (svc FeatureService) DeleteByFeatureId(ctx context.Context, featureId string) error {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		err := svc.Repository.DeleteByFeatureId(txCtx, featureId)
		if err != nil {
			return err
		}

		err = svc.ObjectSvc.DeleteByObjectTypeAndId(txCtx, objecttype.ObjectTypeFeature, featureId)
		if err != nil {
			return err
		}

		err = svc.EventSvc.TrackResourceDeleted(ctx, ResourceTypeFeature, featureId, nil)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
