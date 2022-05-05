package fstore

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

const tokenMetaCollection = "token-meta"

// CreateTokenMeta creates a token meta and returns the firestore version of it
func (fc *FirestoreClient) CreateTokenMeta(ctx context.Context, tokenMeta *TokenMeta) (*FirestoreTokenMeta, error) {
	ref, _, err := fc.Collection(tokenMetaCollection).Add(ctx, tokenMeta)
	if err != nil {
		return nil, err
	}

	snapshot, err := ref.Get(ctx)
	if err != nil {
		return nil, err
	}

	err = snapshot.DataTo(tokenMeta)
	if err != nil {
		return nil, err
	}

	fp := &FirestoreTokenMeta{
		TokenMeta:  *tokenMeta,
		DocumentID: snapshot.Ref.ID,
	}

	return fp, nil
}

// GetTokenMeta returns a TokenMeta by document id
func (fc *FirestoreClient) GetTokenMeta(ctx context.Context, documentID string) (*FirestoreTokenMeta, error) {
	ref := fc.Collection(tokenMetaCollection).Doc(documentID)
	if ref == nil {
		return nil, fmt.Errorf("Plaque not found for document id")
	}
	snapshot, err := ref.Get(ctx)
	if err != nil {
		return nil, err
	}

	tokenMeta := new(TokenMeta)
	err = snapshot.DataTo(tokenMeta)
	if err != nil {
		return nil, err
	}

	return &FirestoreTokenMeta{TokenMeta: *tokenMeta, DocumentID: ref.ID}, nil
}

// GetTokenMetaList returns a list of token metas for a document id list
func (fc *FirestoreClient) GetTokenMetaList(ctx context.Context, documentIDList []string) ([]*FirestoreTokenMeta, error) {

	tokenMetaList := make([]*FirestoreTokenMeta, 0, len(documentIDList))
	for _, id := range documentIDList {
		firestoreTokenMeta, err := fc.GetTokenMeta(ctx, id)
		if err != nil {
			return nil, err
		}
		tokenMetaList = append(tokenMetaList, firestoreTokenMeta)
	}

	return tokenMetaList, nil
}

// GetTokenMetaByQuery returns a list of token meta by a given firestore query
func (fc *FirestoreClient) GetTokenMetaByQuery(ctx context.Context, query FirestoreQuery) ([]*FirestoreTokenMeta, error) {
	log.Printf("QUERY %v", query)
	iter := fc.Collection(tokenMetaCollection).Where(query.Path, query.Op, query.Value).Documents(ctx)

	tokenMetaList := make([]*FirestoreTokenMeta, 0, 5)
	for {
		snap, err := iter.Next()
		log.Printf("ITER %+v", err)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		tokenMeta := new(TokenMeta)
		err = snap.DataTo(tokenMeta)
		if err != nil {
			return nil, err
		}
		tokenMetaList = append(tokenMetaList, &FirestoreTokenMeta{TokenMeta: *tokenMeta, DocumentID: snap.Ref.ID})
	}
	return tokenMetaList, nil
}

// UpdateTokenMeta performs a list of updates to the given document
func (fc *FirestoreClient) UpdateTokenMeta(ctx context.Context, documentID string, update []firestore.Update) error {
	_, err := fc.Collection(tokenMetaCollection).Doc(documentID).Update(ctx, update)
	return err
}

// DeleteAllTokenMetas removes all token meta documents, should only be used in tests
func (fc *FirestoreClient) DeleteAllTokenMetas(ctx context.Context) error {
	iter := fc.Collection(tokenMetaCollection).Documents(ctx)
	for {
		snap, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		_, err = snap.Ref.Delete(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
