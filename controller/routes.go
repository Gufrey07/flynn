package main

import (
	"net/http"
	"sort"

	"github.com/flynn/flynn/controller/data"
	"github.com/flynn/flynn/controller/schema"
	"github.com/flynn/flynn/pkg/ctxhelper"
	"github.com/flynn/flynn/pkg/httphelper"
	router "github.com/flynn/flynn/router/types"
	"golang.org/x/net/context"
)

func (c *controllerAPI) CreateRoute(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	var route router.Route
	if err := httphelper.DecodeJSON(req, &route); err != nil {
		respondWithError(w, err)
		return
	}
	route.ParentRef = routeParentRef(c.getApp(ctx).ID)

	if err := schema.Validate(&route); err != nil {
		respondWithError(w, err)
		return
	}

	err := c.routeRepo.Add(&route)
	if err != nil {
		respondWithError(w, err)
		return
	}

	httphelper.JSON(w, 200, &route)
}

func (c *controllerAPI) GetRoute(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	route, err := c.getRoute(ctx)
	if err != nil {
		respondWithError(w, err)
		return
	}

	httphelper.JSON(w, 200, route)
}

type sortedRoutes []*router.Route

func (p sortedRoutes) Len() int           { return len(p) }
func (p sortedRoutes) Less(i, j int) bool { return p[i].CreatedAt.After(p[j].CreatedAt) }
func (p sortedRoutes) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (c *controllerAPI) GetRouteList(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	routes, err := c.routeRepo.List("")
	if err != nil {
		respondWithError(w, err)
		return
	}
	sort.Sort(sortedRoutes(routes))
	httphelper.JSON(w, 200, routes)
}

func (c *controllerAPI) GetAppRouteList(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	routes, err := c.routeRepo.List(routeParentRef(c.getApp(ctx).ID))
	if err != nil {
		respondWithError(w, err)
		return
	}
	sort.Sort(sortedRoutes(routes))
	httphelper.JSON(w, 200, routes)
}

func (c *controllerAPI) UpdateRoute(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	params, _ := ctxhelper.ParamsFromContext(ctx)

	var route router.Route
	if err := httphelper.DecodeJSON(req, &route); err != nil {
		respondWithError(w, err)
		return
	}
	route.Type = params.ByName("routes_type")
	route.ID = params.ByName("routes_id")

	if err := c.routeRepo.Update(&route); err != nil {
		if err == data.ErrRouteNotFound {
			err = ErrNotFound
		}
		respondWithError(w, err)
		return
	}
	httphelper.JSON(w, 200, route)
}

func (c *controllerAPI) DeleteRoute(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	route, err := c.getRoute(ctx)
	if err != nil {
		respondWithError(w, err)
		return
	}

	if err := c.routeRepo.Delete(route); err != nil {
		if err == data.ErrRouteNotFound {
			err = ErrNotFound
		}
		respondWithError(w, err)
		return
	}
	w.WriteHeader(200)
}
