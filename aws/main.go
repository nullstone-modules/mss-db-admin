package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/nullstone-modules/mss-db-admin/aws/secrets"
	crud_invoke "github.com/nullstone-modules/mss-db-admin/crud-invoke"
	"github.com/nullstone-modules/mss-db-admin/sqlserver"
)

const (
	dbConnUrlSecretIdEnvVar = "DB_CONN_URL_SECRET_ID"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	connUrlSecretId := os.Getenv(dbConnUrlSecretIdEnvVar)
	log.Printf("Retrieving connection url secret (%s)\n", connUrlSecretId)
	connUrl, err := secrets.GetString(ctx, connUrlSecretId)
	if err != nil {
		log.Println(err.Error())
	}

	store := sqlserver.NewStore(connUrl)
	defer store.Close()

	lambda.Start(HandleRequest(store))
}

func HandleRequest(store *sqlserver.Store) func(ctx context.Context, rawEvent json.RawMessage) (any, error) {
	return func(ctx context.Context, rawEvent json.RawMessage) (any, error) {
		if ok, event := crud_invoke.IsEvent(rawEvent); ok {
			log.Println("Invocation (CRUD) Event", event.Tf.Action, event.Type)
			return crud_invoke.Handle(ctx, event, store)
		}

		log.Println("Unknown Event", string(rawEvent))
		return nil, nil
	}
}
