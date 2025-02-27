package authz

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

// GetRoutes registers all route handlers for this module
func (svc PermissionService) Routes() []service.Route {
	return []service.Route{
		// create
		{
			Pattern: "/v1/permissions",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, CreateHandler),
		},

		// get
		{
			Pattern: "/v1/permissions",
			Method:  "GET",
			Handler: middleware.ChainMiddleware(
				service.NewRouteHandler(svc, ListHandler),
				middleware.ListMiddleware[PermissionListParamParser],
			),
		},
		{
			Pattern: "/v1/permissions/{permissionId}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, GetHandler),
		},

		// update
		{
			Pattern: "/v1/permissions/{permissionId}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, UpdateHandler),
		},
		{
			Pattern: "/v1/permissions/{permissionId}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc, UpdateHandler),
		},

		// delete
		{
			Pattern: "/v1/permissions/{permissionId}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, DeleteHandler),
		},
	}
}

func CreateHandler(svc PermissionService, w http.ResponseWriter, r *http.Request) error {
	var newPermission PermissionSpec
	err := service.ParseJSONBody(r.Body, &newPermission)
	if err != nil {
		return err
	}

	createdPermission, err := svc.Create(r.Context(), newPermission)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, createdPermission)
	return nil
}

func GetHandler(svc PermissionService, w http.ResponseWriter, r *http.Request) error {
	permissionIdParam := mux.Vars(r)["permissionId"]
	permissionId, err := url.QueryUnescape(permissionIdParam)
	if err != nil {
		return service.NewInvalidParameterError("permissionId", "")
	}

	permission, err := svc.GetByPermissionId(r.Context(), permissionId)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, permission)
	return nil
}

func ListHandler(svc PermissionService, w http.ResponseWriter, r *http.Request) error {
	listParams := middleware.GetListParamsFromContext(r.Context())
	permissions, err := svc.List(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, permissions)
	return nil
}

func UpdateHandler(svc PermissionService, w http.ResponseWriter, r *http.Request) error {
	var updatePermission UpdatePermissionSpec
	err := service.ParseJSONBody(r.Body, &updatePermission)
	if err != nil {
		return err
	}

	permissionIdParam := mux.Vars(r)["permissionId"]
	permissionId, err := url.QueryUnescape(permissionIdParam)
	if err != nil {
		return service.NewInvalidParameterError("permissionId", "")
	}

	updatedPermission, err := svc.UpdateByPermissionId(r.Context(), permissionId, updatePermission)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, updatedPermission)
	return nil
}

func DeleteHandler(svc PermissionService, w http.ResponseWriter, r *http.Request) error {
	permissionId := mux.Vars(r)["permissionId"]
	if permissionId == "" {
		return service.NewMissingRequiredParameterError("permissionId")
	}

	err := svc.DeleteByPermissionId(r.Context(), permissionId)
	if err != nil {
		return err
	}

	return nil
}
