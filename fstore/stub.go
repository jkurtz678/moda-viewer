package fstore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
)

type FstoreClientStub struct{}

// createPlaque return err to simulate offline client
func (f *FstoreClientStub) CreatePlaque(ctx context.Context, plaque *Plaque) (*FirestorePlaque, error) {
	return nil, fmt.Errorf("error offline")
}

// GetPlaque return err to simulate offline client
func (f *FstoreClientStub) GetPlaque(ctx context.Context, documentID string) (*FirestorePlaque, error) {
	return nil, fmt.Errorf("error offline")
}

// UpdatePlaque return err to simulate offline client
func (f *FstoreClientStub) UpdatePlaque(ctx context.Context, documentID string, update []firestore.Update) error {
	return fmt.Errorf("error offline")
}
