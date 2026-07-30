// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package web

import (
	"net/http"

	"gitea.dev/modules/setting"
	"gitea.dev/modules/validation"
	"gitea.dev/modules/web"
	"gitea.dev/routers/common"
	"gitea.dev/routers/web/repo"
	"gitea.dev/services/context"
)

// GitOnlyRoutes returns a minimal router containing only Git HTTP protocol
// routes (clone/push/pull) and Git LFS routes. All Web UI routes are excluded.
// This is used by the LFS-customized build to serve as a pure Git backend.
func GitOnlyRoutes() *web.Router {
	routes := web.NewRouter()

	// Initialize binding rules (required for API JSON binding to work)
	validation.AddBindingRules()

	var mid []any
	mid = append(mid, common.MustInitSessioner(), context.Contexter())

	webAuth := newWebAuthMiddleware()
	mid = append(mid, webAuth.MiddlewareHandler)

	gitRoutes := web.NewRouter()
	gitRoutes.AfterRouting(mid...)

	// Git LFS support
	lfsServerEnabled := func(ctx *context.Context) {
		if !setting.LFS.StartServer {
			ctx.HTTPError(http.StatusNotFound)
			return
		}
	}
	common.AddOwnerRepoGitLFSRoutes(gitRoutes, lfsServerEnabled, webAuth.AllowBasic, repo.CorsHandler(), optSignInFromAnyOrigin)

	// Git Smart HTTP protocol (clone/push/pull)
	addOwnerRepoGitHTTPRouters(gitRoutes, repo.HTTPGitEnabledHandler, webAuth.AllowBasic, webAuth.AllowOAuth2, repo.CorsHandler(), optSignInFromAnyOrigin, context.UserAssignmentWeb())

	routes.Mount("", gitRoutes)
	return routes
}
