package main

import (
	"log"
	"net/http"
	"os"
	"query-service/db"
	"query-service/generated"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow all origins; adjust as needed for production.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			return
		}
		h.ServeHTTP(w, r)
	})
}

// Example using Apollo Sandbox redirection.
// If you still want an embedded solution, you may try GraphQL Playground instead.
func playgroundRedirectHandler(w http.ResponseWriter, r *http.Request) {
	// Redirect to Apollo Sandbox which expects the endpoint to be set in its UI.
	// Note: Apollo Sandbox must be loaded in its own tab rather than embedded.
	http.Redirect(w, r, "https://studio.apollographql.com/sandbox/explorer?endpoint=http://localhost:8081/graphql", http.StatusTemporaryRedirect)
}

func main() {

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://root@localhost:26257/layerg?sslmode=disable"
	}
	db.InitDB(connStr)
	s, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name:   "Query",
			Fields: generated.QueryFields,
		}),
	})
	if err != nil {
		log.Fatal(err)
	}

	h := handler.New(&handler.Config{
		Schema: &s,
		Pretty: true,
	})
	// Wrap the GraphQL handler with CORS support.
	http.Handle("/graphql", withCORS(h))
	// Provide a playground endpoint – here we redirect to Apollo Sandbox.
	http.HandleFunc("/playground", playgroundRedirectHandler)

	log.Println("Server running on :8081")
	http.ListenAndServe(":8081", nil)
}
