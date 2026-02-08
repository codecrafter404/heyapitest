package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type Options struct {
	Host      string `doc:"The host to listen on" default:"0.0.0.0" short:"-"`
	Port      int    `doc:"Listening Port" short:"p" default:"8080"`
	PublicUrl string `doc:"The server's public URL" default:"http://localhost:8080/"`
}

type GreetingOutput struct {
	Body struct {
		Message string `json:"msg" example:"Hello John Doe!" doc:"The response of a greeting"`
	}
}

func NewAuthMiddleware(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		var scopes []string
		for _, s := range ctx.Operation().Security {
			if res, ok := s["token"]; ok {
				scopes = res
				break
			}
		}
		if slices.Contains(scopes, "NO_AUTH") && len(scopes) == 1 {
			next(ctx)
			return
		}
		if len(scopes) == 0 {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Missing or invalid session cookie")
			return
		}

		session, err := huma.ReadCookie(ctx, "session")
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Missing or invalid session cookie")
			return
		}

		log.Info().Any("session", session).Msg("Got session")
		//TODO: implement
		if session.Value == "super-secure-cookie" {
			next(ctx)
			return

		}
		huma.WriteErr(api, ctx, http.StatusUnauthorized, "Missing or invalid session cookie")
	}
}

type LoginResponse struct {
	Body struct {
		Success bool
	}
	SetCookie http.Cookie `header:"Set-Cookie"`
}

//go:embed embedded/*
var content embed.FS

func main() {
	router := chi.NewMux()
	static, err := fs.Sub(content, "embedded")
	if err != nil {
		panic(fmt.Sprintf("Failed to extract embedded: %v", err))
	}

	file_indexb, err := fs.ReadFile(static, "index.html")
	if err != nil {
		panic(fmt.Sprintf("Failed to extract index.html: %v", err))
	}

	file_nfb, err := fs.ReadFile(static, "404.html")
	if err != nil {
		panic(fmt.Sprintf("Failed to extract 404.html: %v", err))
	}

	var file_index []byte
	var file_nf []byte

	router.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		log.Info().Str("path", r.URL.Path).Msg("get /*")

		switch r.URL.Path {
		case "/":
			w.Write(file_index)
		case "/index.html":
			w.Write(file_index)
		case "/404.html":
			w.Write(file_nf)
		default:
			res := http.FileServerFS(static)
			res.ServeHTTP(w, r)
		}
	})

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		log.Info().Str("path", r.URL.Path).Msg("404")
		content, err := fs.ReadFile(static, "404.html")
		if err != nil {
			panic(fmt.Sprintf("Failed to serve 404 page: %v", err))
		}
		w.Write(content)
	})

	apiRouter := chi.NewRouter()
	router.Mount("/api", apiRouter)
	config := huma.DefaultConfig("Greetings API", "1.0.0")
	config.Servers = []*huma.Server{
		{
			URL: "/api",
		},
	}

	config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"token": {
			Type: "apiKey",
			In:   "cookie",
			Name: "session",
		},
	}

	api := humachi.New(apiRouter, config)
	api.UseMiddleware(NewAuthMiddleware(api))

	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      http.MethodPost,
		Path:        "/login",
		Summary:     "Log in",
		Description: "By loggin in a cookie will be assinged required for auth",
		Security: []map[string][]string{
			{"token": {"NO_AUTH"}},
		},
	}, func(ctx context.Context, input *struct {
		Body struct {
			Token string `required:"true"`
		}
	}) (*LoginResponse, error) {
		res := &LoginResponse{}
		res.Body.Success = true
		if input.Body.Token != "super-secure-token" {
			return nil, huma.Error401Unauthorized("Invalid token")
		}
		cookie := http.Cookie{
			Name:     "session",
			Value:    "super-secure-cookie",
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Now().Add(time.Second * 10),
		}
		res.SetCookie = cookie

		return res, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-greeting",
		Method:      http.MethodGet,
		Path:        "/greeting/{user}",
		Summary:     "Get a greeting",
		Description: "Greet a person by name",
		Security: []map[string][]string{
			{"token": {"greeting"}},
		},
	}, func(ctx context.Context, i *struct {
		User string `path:"user" maxLength:"15" example:""`
	}) (*GreetingOutput, error) {
		res := GreetingOutput{}
		res.Body.Message = fmt.Sprintf("Hellooooo %s!", i.User)
		return &res, nil
	})

	cli := humacli.New(func(h humacli.Hooks, o *Options) {
		file_mod := fmt.Sprintf("window.APP_CONFIG = {publicURL: \"%s\"}", o.PublicUrl)
		file_index = []byte(strings.ReplaceAll(string(file_indexb), "<!-- VARIABLE_INJECT -->", file_mod))
		file_nf = []byte(strings.ReplaceAll(string(file_nfb), "<!-- VARIABLE_INJECT -->", file_mod))
		fmt.Println(strings.ReplaceAll(string(file_indexb), "<!-- VARIABLE_INJECT -->", file_mod))

		server := http.Server{
			Addr:    fmt.Sprintf("%s:%d", o.Host, o.Port),
			Handler: router,
		}
		h.OnStart(func() {
			log.Info().Str("host", o.Host).Int("port", o.Port).Msg("Listening")
			server.ListenAndServe()
		})
		h.OnStop(func() {
			ctx := context.Background()
			server.Shutdown(ctx)
		})
	})

	cli.Root().AddCommand(&cobra.Command{
		Use:   "openapi",
		Short: "Print the OpenAPI spec",
		Run: func(cmd *cobra.Command, args []string) {
			b, err := api.OpenAPI().YAML()
			if err != nil {
				log.Fatal().Err(err)
			}
			fmt.Println(string(b))
		},
	})
	cli.Run()
}
