package fstore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
)

const plaqueCollection = "plaque"

// CreatePlaque creates a plaque and returns the firestore version of it
func (fc *FirestoreClient) CreatePlaque(ctx context.Context, plaque *Plaque) (*FirestorePlaque, error) {
	ref, _, err := fc.Collection(plaqueCollection).Add(ctx, plaque)
	if err != nil {
		return nil, err
	}

	snapshot, err := ref.Get(ctx)
	if err != nil {
		return nil, err
	}

	err = snapshot.DataTo(plaque)
	if err != nil {
		return nil, err
	}

	fp := &FirestorePlaque{
		Plaque:     *plaque,
		DocumentID: snapshot.Ref.ID,
	}

	return fp, nil
}

// GetPlaque returns a plaque by document id
func (fc *FirestoreClient) GetPlaque(ctx context.Context, documentID string) (*FirestorePlaque, error) {
	ref := fc.Collection(plaqueCollection).Doc(documentID)
	if ref == nil {
		return nil, fmt.Errorf("Plaque not found for document id")
	}
	snapshot, err := ref.Get(ctx)
	if err != nil {
		return nil, err
	}

	plaque := new(Plaque)
	err = snapshot.DataTo(plaque)
	if err != nil {
		return nil, err
	}

	return &FirestorePlaque{Plaque: *plaque, DocumentID: ref.ID}, nil
}

// UpdatePlaque performs a list of updates to the given document
func (fc *FirestoreClient) UpdatePlaque(ctx context.Context, documentID string, update []firestore.Update) error {
	_, err := fc.Collection(plaqueCollection).Doc(documentID).Update(ctx, update)
	return err
}

// ListenPlaque will listen for changes to the given plaque and call the callback function upon changes
// callback will be called immediately once when function is called
func (fc *FirestoreClient) ListenPlaque(ctx context.Context, documentID string, cb func(plaque *FirestorePlaque) error) error {
	ref := fc.Collection(plaqueCollection).Doc(documentID)
	if ref == nil {
		return fmt.Errorf("Plaque not found for document id")
	}
	it := ref.Snapshots(ctx)
	for {
		snap, err := it.Next()
		if err != nil {
			return err
		}

		plaque := new(Plaque)
		err = snap.DataTo(plaque)
		if err != nil {
			return err
		}

		err = cb(&FirestorePlaque{Plaque: *plaque, DocumentID: snap.Ref.ID})
		if err != nil {
			return err
		}
	}
}
