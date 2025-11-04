package server

import (
	"net/http"
	"warehouse/internal/graph"

	"github.com/99designs/gqlgen/graphql/handler"
)

func NewGraphQLHandler(resolver *graph.Resolver) http.Handler {
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv.ServeHTTP(w, r)
	})
}