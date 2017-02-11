// Copyright 2017 Dmitry Frank <mail@dmitryfrank.com>
// Licensed under the BSD, see LICENCE file for details.

package server

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"goji.io"
	"goji.io/pat"

	hh "dmitryfrank.com/geekmarks/server/httphelper"
	"dmitryfrank.com/geekmarks/server/middleware"
	"dmitryfrank.com/geekmarks/server/storage"
	"github.com/golang/glog"
	"github.com/juju/errors"
)

func (gm *GMServer) authnRequiredMiddleware(inner http.Handler) http.Handler {
	mw := func(w http.ResponseWriter, r *http.Request) {
		v := r.Context().Value("authUserData")
		if v == nil {
			// No authentication data: respond with an error
			w.Header().Set("WWW-Authenticate", "Bearer realm=\"login please\"")
			hh.RespondWithError(w, r, hh.MakeUnauthorizedError())
			return
		}

		// Authentication data is found; proceed.
		inner.ServeHTTP(w, r)
	}
	return middleware.MkMiddleware(mw)
}

func parseBearerAuth(r *http.Request) (token string, ok bool) {
	header := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	return header[len(prefix):], true
}

// Middleware which populates the context with the authentication data, if
// it is provided and is correct.
//
// If it's provided but isn't correct, responds with an error. TODO: do we
// really need that behaviour? Maybe it's better to just proceed without authn
// data? Dunno.
//
// NOTE: be sure to use it after httphelper.MakeDesiredContentTypeMiddleware(),
// since the error response should be in the right format
func (gm *GMServer) authnMiddleware(inner http.Handler) http.Handler {
	mw := func(w http.ResponseWriter, r *http.Request) {
		// TODO: use https://github.com/abbot/go-http-auth for digest auth
		token, ok := parseBearerAuth(r)

		if !ok {
			// When connecting via websocket protocol, JavaScript API does not have a
			// way to provide HTTP authorization header, so we have to use a trick
			// here: get token from the query string.
			token = r.FormValue("token")
			if token != "" {
				glog.V(2).Infof("Getting token from the query string")
				ok = true
			}
		}

		if ok {
			var ud *storage.UserData
			err := gm.si.Tx(func(tx *sql.Tx) error {
				ud2, err := gm.si.GetUserByAccessToken(tx, token)

				if err != nil {
					return errors.Trace(err)
				}

				ud = ud2
				return nil
			})
			if err != nil {
				w.Header().Set("WWW-Authenticate", "Bearer realm=\"login please\"")
				hh.RespondWithError(w, r, err)
				return
			}

			// Authn data is correct: create a new request with updated context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "authUserData", ud)
			r = r.WithContext(ctx)
		}

		// Process request, whether authn data was not provided at all, or was
		// provided correctly.
		inner.ServeHTTP(w, r)
	}
	return middleware.MkMiddleware(mw)
}

func getAuthnUserDataByReq(r *http.Request) *storage.UserData {
	v := r.Context().Value("authUserData")
	if v == nil {
		// Not authenticated
		return nil
	}

	return v.(*storage.UserData)
}

func (gm *GMServer) oauthClientIDGet(gmr *GMRequest) (resp interface{}, err error) {
	provider := pat.Param(gmr.HttpReq, "provider")
	oauthCreds, ok := gm.oauthProviders[provider]
	if !ok {
		return nil, errors.Errorf("unknown auth provider: %q", provider)
	}

	resp = clientIDGetResp{
		ClientID: oauthCreds.ClientID,
	}

	return resp, nil
}

func (gm *GMServer) authenticatePost(gmr *GMRequest) (resp interface{}, err error) {
	provider := pat.Param(gmr.HttpReq, "provider")
	oauthCreds, ok := gm.oauthProviders[provider]
	if !ok {
		return nil, errors.Errorf("unknown auth provider: %q", provider)
	}

	if oauthCreds == nil {
		return nil, errors.Errorf("auth provider %q is disabled (corresponding flag to the creds file was not provided)", provider)
	}

	err = gm.si.Tx(func(tx *sql.Tx) error {
		var err error
		switch provider {
		case providerGoogle:
			resp, err = gm.authenticatePostOAuthGoogle(tx, gmr, oauthCreds, googleEndpoint)
			if err != nil {
				return errors.Trace(err)
			}
		default:
			return hh.MakeInternalServerError(
				errors.Errorf("auth provider %q exists, but is not handled", provider),
			)
		}
		return nil
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	return resp, nil
}

func (gm *GMServer) setupAuthAPIEndpoints(mux *goji.Mux, gsu getSubjUser) {
	setUserEndpoint(pat.Get("/client_id"), gm.oauthClientIDGet, nil, mux, gsu)
	setUserEndpoint(pat.Post("/authenticate"), gm.authenticatePost, nil, mux, gsu)
}
