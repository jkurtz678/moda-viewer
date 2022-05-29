package fstore

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

type DBClient interface {
	CreatePlaque(ctx context.Context, plaque *Plaque) (*FirestorePlaque, error)
	GetPlaque(ctx context.Context, documentID string) (*FirestorePlaque, error)
	UpdatePlaque(ctx context.Context, documentID string, update []firestore.Update) error
	ListenPlaque(ctx context.Context, documentID string, cb func(plaque *FirestorePlaque) error) error

	CreateTokenMeta(ctx context.Context, tokenMeta *TokenMeta) (*FirestoreTokenMeta, error)
	GetTokenMeta(ctx context.Context, documentID string) (*FirestoreTokenMeta, error)
	GetTokenMetaList(ctx context.Context, documentIDList []string) ([]*FirestoreTokenMeta, error)
	UpdateTokenMeta(ctx context.Context, documentID string, update []firestore.Update) error
}

type FirestoreClient struct {
	*firestore.Client
}

func NewFirestoreClient(ctx context.Context, credentialsFile string) (*FirestoreClient, error) {
	sa := option.WithCredentialsFile(credentialsFile)
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		return nil, err
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, err
	}
	return &FirestoreClient{Client: client}, nil
}

func NewFirestoreTestClient(ctx context.Context) *FirestoreClient {
	err := os.Setenv("PROJECT", "moda-viewer")
	if err != nil {
		log.Fatal(err)
	}
	err = os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}

	client, err := firestore.NewClient(ctx, "moda-viewer")
	if err != nil {
		log.Fatal(err)
	}
	return &FirestoreClient{Client: client}
}
