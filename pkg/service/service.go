package service

import (
	"bytes"
	"mime"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/skip2/go-qrcode"
	"golang.org/x/time/rate"
)

type Config struct {
	RequireHttps bool     `name:"https"`
	AllowOrigins []string `name:"allow-origin"`
}

func New(cfg Config) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Use(
		middleware.Logger(),
		middleware.HTTPSRedirectWithConfig(middleware.RedirectConfig{
			Skipper: func(c echo.Context) bool {
				return !cfg.RequireHttps
			},
		}),
		middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
			IdentifierExtractor: func(ctx echo.Context) (string, error) {
				return ctx.Request().RemoteAddr, nil
			},
			Store: middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(10),
				Burst:     3,
				ExpiresIn: 5 * time.Minute,
			}),
		}),
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowMethods: []string{http.MethodGet},
			AllowOrigins: cfg.AllowOrigins,
		}),
		middleware.RemoveTrailingSlash(),
		middleware.Gzip(),
	)

	e.GET("/api/v1/qr", generateQr, func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "max-age=604800")
			return next(c)
		}
	})

	return e
}

type QrParams struct {
	Size uint16 `query:"size"`
	Data string `query:"data"`
}

func generateQr(c echo.Context) error {
	params := QrParams{
		Size: 128,
		Data: "",
	}
	binder := echo.QueryParamsBinder(c).
		Uint16("size", &params.Size).
		MustString("data", &params.Data)
	if err := binder.BindError(); err != nil {
		return err
	}
	if params.Size > 512 {
		return echo.NewHTTPError(http.StatusBadRequest, "size must be less than 512")
	}

	qr, err := qrcode.Encode(params.Data, qrcode.Highest, int(params.Size))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.Stream(200, mime.TypeByExtension(".png"), bytes.NewReader(qr))
}
