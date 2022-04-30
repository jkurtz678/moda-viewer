package fstore

import (
	"context"
	"testing"

	"github.com/franela/goblin"
)

func TestPlaque(t *testing.T) {
	g := goblin.Goblin(t)
	ctx := context.Background()
	client := NewFirestoreTestClient(ctx)
	defer client.Close()
	g.Describe("fstore.Plaque", func() {
		g.It("should create and retrieve plaques", func() {
			// create
			p := &Plaque{Name: "test"}
			fp, err := client.CreatePlaque(ctx, p)
			g.Assert(err).IsNil()
			g.Assert(fp.DocumentID != "").IsTrue()

			// retrieve
			fp2, err := client.GetPlaque(ctx, fp.DocumentID)
			g.Assert(err).IsNil()
			g.Assert(fp2.DocumentID).Equal(fp.DocumentID)
		})
	})
}
