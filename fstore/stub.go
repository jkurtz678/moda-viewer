package fstore

import (
	"context"
	"fmt"
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
