package fstore

import "context"

const plaqueCollection = "plaque"

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
