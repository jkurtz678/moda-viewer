package fstore

import (
	"context"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/franela/goblin"
)

func TestTokenMeta(t *testing.T) {
	g := goblin.Goblin(t)
	ctx := context.Background()
	client := NewFirestoreTestClient(ctx)
	defer client.Close()
	g.Describe("fstore.TokenMeta", func() {
		g.It("should create, retrieve, and update token metas", func() {
			// create
			tm := &TokenMeta{Name: "test"}
			ftm, err := client.CreateTokenMeta(ctx, tm)
			g.Assert(err).IsNil()
			g.Assert(ftm.DocumentID != "").IsTrue()

			// retrieve
			ftm2, err := client.GetTokenMeta(ctx, ftm.DocumentID)
			g.Assert(err).IsNil()
			g.Assert(ftm2.DocumentID).Equal(ftm.DocumentID)

			// update
			g.Assert(client.UpdateTokenMeta(ctx, ftm2.DocumentID, []firestore.Update{{
				Path: "name", Value: "update-test",
			}})).IsNil()

			// confirm change
			ftm3, err := client.GetTokenMeta(ctx, ftm2.DocumentID)
			g.Assert(err).IsNil()
			g.Assert(ftm3.TokenMeta.Name).Equal("update-test")
		})

		g.It("should return a list of token metas", func() {
			tm1 := &TokenMeta{Name: "sunday", Artist: "georges seurat"}
			ftm1, err := client.CreateTokenMeta(ctx, tm1)
			g.Assert(err).IsNil()

			tm2 := &TokenMeta{Name: "starry night", Artist: "van gogh"}
			ftm2, err := client.CreateTokenMeta(ctx, tm2)
			g.Assert(err).IsNil()

			tm3 := &TokenMeta{Name: "irises", Artist: "van gogh"}
			ftm3, err := client.CreateTokenMeta(ctx, tm3)
			g.Assert(err).IsNil()

			tmList, err := client.GetTokenMetaList(ctx, []string{ftm1.DocumentID, ftm2.DocumentID, ftm3.DocumentID})
			g.Assert(err).IsNil()
			g.Assert(len(tmList)).Equal(3)

			queryMetas, err := client.GetTokenMetaByQuery(ctx, FirestoreQuery{Path: "artist", Op: "==", Value: "van gogh"})
			g.Assert(err).IsNil()
			g.Assert(len(queryMetas)).Equal(2)
			for _, m := range queryMetas {
				g.Assert(m.TokenMeta.Artist).Equal("van gogh")
			}
		})
	})

	// cleanup token metas
	err := client.DeleteAllTokenMetas(ctx)
	if err != nil {
		t.Fatal(err)
	}
}
