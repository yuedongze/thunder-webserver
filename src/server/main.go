package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/samsarahq/thunder/graphql"
	"github.com/samsarahq/thunder/graphql/graphiql"
	"github.com/samsarahq/thunder/graphql/introspection"
	"github.com/samsarahq/thunder/graphql/schemabuilder"
	"github.com/samsarahq/thunder/reactive"
)

type post struct {
	Title     string
	Body      string
	CreatedAt time.Time
}

// server is our graphql server.
type server struct {
	posts []post
}

// registerQuery registers the root query type.
func (s *server) registerQuery(schema *schemabuilder.Schema) {
	obj := schema.Query()

	obj.FieldFunc("posts", func(ctx context.Context) []post {
		reactive.InvalidateAfter(ctx, 5 * time.Second)
		return s.posts
	})
}

// registerMutation registers the root mutation type.
func (s *server) registerMutation(schema *schemabuilder.Schema) {
	object := schema.Mutation()
	object.FieldFunc("addMessage", func(ctx context.Context, args struct{ Text string }) (string, error) {
		return args.Text, nil
	})
	object.FieldFunc("addPost", func(ctx context.Context, args struct {
		Title string
		Body  string
	}) error {
		s.posts = append(s.posts, post{Title: args.Title, Body: args.Body, CreatedAt: time.Now()})
		return nil
	})
}

// registerPost registers the post type.
func (s *server) registerPost(schema *schemabuilder.Schema) {
	obj := schema.Object("Post", post{})
	obj.FieldFunc("age", func(ctx context.Context, p *post) string {
		reactive.InvalidateAfter(ctx, 5*time.Second)
		return time.Since(p.CreatedAt).String()
	})
}

// schema builds the graphql schema.
func (s *server) schema() *graphql.Schema {
	builder := schemabuilder.NewSchema()
	s.registerQuery(builder)
	s.registerMutation(builder)
	s.registerPost(builder)

	valueJSON, err := introspection.ComputeSchemaJSON(*builder)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(valueJSON))

	return builder.MustBuild()
}

func main() {
	// Instantiate a server, build a server, and serve the schema on port 3030.
	server := &server{
		posts: []post{
			{Title: "first post!", Body: "I was here first!", CreatedAt: time.Now()},
			{Title: "graphql", Body: "did you hear about Thunder?", CreatedAt: time.Now()},
			{Title: "hello world", Body: "this is made by Steven.", CreatedAt: time.Now()},
		},
	}

	schema := server.schema()
	introspection.AddIntrospectionToSchema(schema)

	// Expose schema and graphiql.
	http.Handle("/graphql", graphql.Handler(schema))
	http.Handle("/graphiql/", http.StripPrefix("/graphiql/", graphiql.Handler()))
	http.ListenAndServe(":3030", nil)
}
