package fstore

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/firestore"
)

type FstoreClientStub struct {
	ListenerWaitGroup sync.WaitGroup
}

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

// CreateTokenMeta return err to simulate offline client
func (f *FstoreClientStub) CreateTokenMeta(ctx context.Context, tokenMeta *TokenMeta) (*FirestoreTokenMeta, error) {
	return nil, fmt.Errorf("error offline")
}

// GetTokenMeta return err to simulate offline client
func (f *FstoreClientStub) GetTokenMeta(ctx context.Context, documentID string) (*FirestoreTokenMeta, error) {
	return nil, fmt.Errorf("error offline")
}

// GetTokenMetaList return err to simulate offline client
func (f *FstoreClientStub) GetTokenMetaList(ctx context.Context, documentIDList []string) ([]*FirestoreTokenMeta, error) {
	return nil, fmt.Errorf("error offline")
}

// UpdateTokenMeta return err to simulate offline client
func (f *FstoreClientStub) UpdateTokenMeta(ctx context.Context, documentID string, update []firestore.Update) error {
	return fmt.Errorf("error offline")
}

// ListenPlaque return err to simulate offline client
func (fc *FstoreClientStub) ListenPlaque(ctx context.Context, documentID string, cb func(plaque *FirestorePlaque) error) error {
	//fc.ListenerWaitGroup.Done() // indicate listening has started
	return fmt.Errorf("error offline")
}
