package fstore

import (
	"context"
	"sync"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/franela/goblin"
	"github.com/stretchr/testify/assert"
)

func TestPlaque(t *testing.T) {
	g := goblin.Goblin(t)
	ctx := context.Background()
	client := NewFirestoreTestClient(ctx)
	defer client.Close()
	g.Describe("fstore.Plaque", func() {
		g.It("should create, retrieve, and update plaques", func() {
			// create
			p := &Plaque{Name: "test"}
			fp, err := client.CreatePlaque(ctx, p)
			g.Assert(err).IsNil()
			g.Assert(fp.DocumentID != "").IsTrue()

			// retrieve
			fp2, err := client.GetPlaque(ctx, fp.DocumentID)
			g.Assert(err).IsNil()
			g.Assert(fp2.DocumentID).Equal(fp.DocumentID)

			// update
			g.Assert(client.UpdatePlaque(ctx, fp2.DocumentID, []firestore.Update{{
				Path: "name", Value: "update-test",
			}})).IsNil()

			// confirm change
			fp3, err := client.GetPlaque(ctx, fp2.DocumentID)
			g.Assert(err).IsNil()
			g.Assert(fp3.Plaque.Name).Equal("update-test")
		})
	})
}

func TestListenPlaque(t *testing.T) {
	a := assert.New(t)
	ctx := context.Background()
	client := NewFirestoreTestClient(ctx)
	defer client.Close()

	p := &Plaque{Name: "test"}
	fp, err := client.CreatePlaque(ctx, p)
	a.NoError(err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := client.ListenPlaque(ctx, fp.DocumentID, func(plaque *FirestorePlaque) error {
			defer wg.Done()
			a.Equal("update-test", plaque.Plaque.Name)
			return nil
		})
		a.NoError(err)
	}()

	a.NoError(client.UpdatePlaque(ctx, fp.DocumentID, []firestore.Update{{
		Path: "name", Value: "update-test",
	}}))

	wg.Wait()

}
