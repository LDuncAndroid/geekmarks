package server

import (
	"database/sql"
	"net/http"
	"strconv"

	goji "goji.io"
	"goji.io/pat"

	"dmitryfrank.com/geekmarks/server/cptr"
	hh "dmitryfrank.com/geekmarks/server/httphelper"
	"dmitryfrank.com/geekmarks/server/middleware"
	"dmitryfrank.com/geekmarks/server/storage"
	"github.com/golang/glog"
	"github.com/juju/errors"
)

type GMServer struct {
	si storage.Storage
}

func New(si storage.Storage) (*GMServer, error) {
	return &GMServer{
		si: si,
	}, nil
}

func (gm *GMServer) CreateHandler() (http.Handler, error) {
	rRoot := goji.NewMux()
	rRoot.Use(middleware.MakeLogger())

	rAPI := goji.SubMux()
	rRoot.Handle(pat.New("/api/*"), rAPI)
	{
		rAPI.Use(hh.MakeDesiredContentTypeMiddleware("application/json"))
		// We use authnMiddleware here and not on the root router above, since we
		// need hh.MakeDesiredContentTypeMiddleware to go before it.
		rAPI.Use(gm.authnMiddleware)

		rAPIUsers := goji.SubMux()
		rAPI.Handle(pat.New("/users/:userid/*"), rAPIUsers)
		{
			gm.setupUserAPIEndpoints(rAPIUsers, gm.getUserFromURLParam)
		}

		rAPIMy := goji.SubMux()
		rAPI.Handle(pat.New("/my/*"), rAPIMy)
		{
			// "my" endpoints don't make sense for non-authenticated users
			rAPIMy.Use(gm.authnRequiredMiddleware)

			gm.setupUserAPIEndpoints(rAPIMy, gm.getUserFromAuthn)
		}

		rAPI.HandleFunc(
			pat.Get("/test_internal_error"), hh.MakeAPIHandler(
				func(r *http.Request) (resp interface{}, err error) {
					errTest := errors.Errorf("some private error")
					errTest = errors.Annotatef(errTest, "private annotation")
					return nil, errors.Annotatef(hh.MakeInternalServerError(errTest), "public annotation")
				},
			),
		)
	}

	return rRoot, nil
}

type getUser func(r *http.Request) (*storage.UserData, error)

// Sets up user-related endpoints at a given mux. We need this function since
// we have two ways to access user data: through the "/api/users/:userid" and
// through the shortcut "/api/my"; so, in order to avoid duplication, this
// function sets up everything given the function that gets user data.
func (gm *GMServer) setupUserAPIEndpoints(mux *goji.Mux, gu getUser) {
	mkUserHandler := func(
		uh func(r *http.Request, gu getUser) (resp interface{}, err error),
		gu getUser,
	) func(r *http.Request) (resp interface{}, err error) {
		return func(r *http.Request) (resp interface{}, err error) {
			return uh(r, gu)
		}
	}

	{
		handler := hh.MakeAPIHandler(mkUserHandler(gm.userTagsGet, gu))
		mux.HandleFunc(pat.Get("/tags"), handler)
		mux.HandleFunc(pat.Get("/tags/*"), handler)
	}

	{
		handler := hh.MakeAPIHandler(mkUserHandler(gm.userTagsPost, gu))
		mux.HandleFunc(pat.Post("/tags"), handler)
		mux.HandleFunc(pat.Post("/tags/*"), handler)
	}
}

// Retrieves user data from the userid given in an URL, like "123" in
// "/api/users/123/foo/bar"
func (gm *GMServer) getUserFromURLParam(r *http.Request) (*storage.UserData, error) {
	useridStr := pat.Param(r, "userid")
	userid, err := strconv.Atoi(useridStr)
	if err != nil {
		return nil, errors.Errorf("invalid user id: %q", useridStr)
	}

	var ud *storage.UserData
	err = gm.si.Tx(func(tx *sql.Tx) error {
		var err error
		ud, err = gm.si.GetUser(tx, &storage.GetUserArgs{
			ID: cptr.Int(userid),
		})
		return errors.Trace(err)
	})
	if err != nil {
		glog.Errorf(
			"Failed to get user with id %d (from URL param): %s", userid, err,
		)
		return nil, errors.Errorf("invalid user id: %q", useridStr)
	}

	return ud, nil
}

// Retrieves user data from the authentication data
func (gm *GMServer) getUserFromAuthn(r *http.Request) (*storage.UserData, error) {
	// authUserData should always be present here thanks to
	// authnRequiredMiddleware
	ud := r.Context().Value("authUserData")
	if ud == nil {
		return nil, hh.MakeInternalServerError(
			errors.Errorf("authUserData is nil but it should not be"),
		)
	}

	return ud.(*storage.UserData), nil
}
