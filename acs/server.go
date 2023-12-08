package acs

import (
	"crypto/subtle"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type AuthType string

const (
	AuthTypeNone  AuthType = "None"
	AuthTypeBasic AuthType = "Basic"
)

type Options struct {
	AuthType     AuthType
	AuthUsername string
	AuthPassword string
	DumpBody     bool
}
type AcsServer struct {
	logger              *zap.Logger
	dataRetentionPeriod time.Duration
	uploadBucket        string
	handler             AcsHanlder
}

func NewAcsServer(handler AcsHanlder, dataRetentionPeriod time.Duration) *AcsServer {
	s := AcsServer{
		logger:              zap.L().Named("acs"),
		handler:             handler,
		uploadBucket:        "acs-upload",
		dataRetentionPeriod: dataRetentionPeriod,
	}
	return &s
}

func (s *AcsServer) SetupPostEchoGroup(group *echo.Group, sessionStore sessions.Store) *echo.Group {
	return s.SetupPostEchoGroupWithOptions(group, sessionStore, Options{
		AuthType: AuthTypeNone,
		DumpBody: false,
	})
}

func (s *AcsServer) SetupPostEchoGroupWithOptions(group *echo.Group, sessionStore sessions.Store, opts Options) *echo.Group {
	if opts.DumpBody {
		group.Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
			limit := 1000
			reqTmp := string(reqBody)
			if size := len(reqTmp); size > limit*2 {
				reqTmp = reqTmp[0:limit] + " ... " + reqTmp[size-limit:]
			}
			respTmp := string(resBody)
			if size := len(respTmp); size > limit*2 {
				respTmp = respTmp[0:limit] + " ... " + respTmp[size-limit:]
			}
			s.logger.Debug("acs", zap.String("request", reqTmp))
			s.logger.Debug("acs", zap.String("response", respTmp))
		}))
	}

	if opts.AuthType == AuthTypeBasic {
		group.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
			if subtle.ConstantTimeCompare([]byte(username), []byte(opts.AuthUsername)) == 1 &&
				subtle.ConstantTimeCompare([]byte(password), []byte(opts.AuthPassword)) == 1 {
				return true, nil
			}
			return false, nil
		}))
	}

	group.Use(session.Middleware(sessionStore))
	group.POST("", s.HandlePost)
	return group
}

func (s *AcsServer) SetupUploadEchoGroup(group *echo.Group) *echo.Group {
	group.POST("/:name", s.HandleUpload)
	group.PUT("/:name", s.HandleUpload)
	return group
}
