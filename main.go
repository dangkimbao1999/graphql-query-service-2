package main

import (
	"log"
	"net/http"
	"os"
	"query-service/core"
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
	http.Redirect(w, r, "https://studio.apollographql.com/sandbox/explorer?endpoint=http://localhost:8082/graphql", http.StatusTemporaryRedirect)
}

func main() {
	log.Println("Starting server initialization...")

	// Initialize database connection
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://root@localhost:26257/layerg?sslmode=disable"
		log.Printf("Using default connection string: %s", connStr)
	}

	log.Println("Initializing database connection...")
	db.InitDB(connStr)
	log.Println("Database connection established successfully")

	// Load schema
	schema, err := core.LoadSchema("schema.graphql")
	if err != nil {
		log.Fatalf("Failed to load schema: %v", err)
	}

	// Create query resolver with schema
	resolver := &core.QueryResolver{Schema: schema}

	// Create GraphQL schema
	s, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name:   "Query",
			Fields: generated.CreateQueryFields(resolver),
		}),
	})
	if err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}
	log.Println("GraphQL schema created successfully")

	// Configure the GraphQL handler
	h := handler.New(&handler.Config{
		Schema: &s,
		Pretty: true,
	})

	// Set up routes with CORS
	http.Handle("/graphql", withCORS(h))
	// Provide a playground endpoint – here we redirect to Apollo Sandbox.
	http.HandleFunc("/playground", playgroundRedirectHandler)

	// Start the server
	port := ":8082"
	log.Printf("Server initialization complete. Starting HTTP server on port %s", port)
	log.Printf("GraphQL endpoint: http://localhost%s/graphql", port)
	log.Printf("GraphQL Playground: http://localhost%s/playground", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
