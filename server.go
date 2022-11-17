package rest

import (
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

const (
	defaultAddr = "127.0.0.1:8080"
)

var (
	defaultCORS = cors.New(
		cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"*"},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		},
	)
)

type API struct {
	logger *zap.Logger
	addr   string
	router *mux.Router
	cors   *cors.Cors
}

func WithCORS(origins []string, methods []string, headers []string) func(*API) {
	return func(a *API) {
		a.cors = cors.New(cors.Options{
			AllowOriginFunc: func(origin string) bool {
				fmt.Println(origin)
				for _, o := range origins {
					if origin == o {
						return true
					}
				}
				return false
			},
			AllowedMethods:   methods,
			AllowedHeaders:   headers,
			AllowCredentials: false,
		})

		// CORS polices from the browser perform preflight-requests
		// to validate cross-origin header etc within an OPTIONS requests.
		// This request is not there to actually perform the requested action
		// but moreover for informational purposes as such we can return http.StatusOK
		// as a response asap without wasting time.
		a.router.Use(
			func(next http.Handler) http.Handler {
				// OPTIONS requests are request made by the browser to check origin policies and
				// check header meta-data. As such no processing need to happen be we can return OK
				// ASAP
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

					if ok := subtle.ConstantTimeCompare([]byte(r.Method), []byte("OPTIONS")); ok == 1 {
						NoContent().Write(w)
						return
					}

					next.ServeHTTP(w, r)
				})
			},
		)
	}
}

func WithTracing() func(*API) {
	return func(a *API) {
		a.router.Use(
			// inject tracing id which also is returned in the response as a X-Request-ID
			func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ctx := NewRequestID(r.Context())
					w.Header().Set("x-request-id", RequestID(ctx))

					next.ServeHTTP(w, r.WithContext(ctx))
				})
			},
		)
	}
}

func WithAddr(addr string) func(*API) {
	return func(a *API) {
		a.addr = addr
	}
}

func New(logger *zap.Logger, opts ...func(*API)) *API {

	a := API{
		logger: logger,
		addr:   defaultAddr,
		router: mux.NewRouter(),
		cors:   defaultCORS,
	}

	for _, opt := range opts {
		opt(&a)
	}

	return &a
}

func (a API) Route(prefix string) *mux.Route {
	return a.router.PathPrefix(prefix)
}

func (a API) Listen() error {

	ln, err := net.Listen("tcp", a.addr)
	if err != nil {
		return fmt.Errorf("[api.Listen] %w", err)
	}

	a.logger.Info("starting API service", zap.String("address", a.addr))
	return http.Serve(ln, a.cors.Handler(a.router))
}
