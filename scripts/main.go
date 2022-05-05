package main

import (
	"context"
	"flag"
	"jkurtz678/moda-viewer/fstore"
	"jkurtz678/moda-viewer/viewer"
	"log"

	"cloud.google.com/go/firestore"
)

var ctx = context.Background()

func main() {
	script := flag.String("s", "", "name of script to run")
	name := flag.String("n", "", "generic name argument, usage depends on script definition")
	flag.Parse()

	if *script == "" {
		log.Fatalf("error - no script name specified")
	}

	switch *script {
	case "namePlaque":
		nameLocalPlaque(*name)
	case "assignArtist":
		assignArtistToPlaque(*name)
	default:
		log.Printf("No matching script name found for %s", *script)
	}
}

func nameLocalPlaque(name string) {
	if name == "" {
		log.Fatalf("error - no plaque name specified")
	}

	v, fc := getScriptClients()

	plaque, err := v.ReadLocalPlaqueFile()
	if err != nil {
		log.Fatal(err)
	}

	plaque.Plaque.Name = name

	err = fc.UpdatePlaque(ctx, plaque.DocumentID, []firestore.Update{{Path: "name", Value: name}})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Plaque renamed to %s", name)
}

func assignArtistToPlaque(name string) {
	if name == "" {
		log.Fatalf("error - no artist name specified")
	}

	v, fc := getScriptClients()

	log.Printf("name %+s", name)
	metas, err := fc.GetTokenMetaByQuery(ctx, fstore.FirestoreQuery{Path: "artist", Op: "==", Value: name})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("metas %+v", metas)

	plaque, err := v.ReadLocalPlaqueFile()
	if err != nil {
		log.Fatal(err)
	}

	metaIDs := make([]string, 0, len(metas))
	for _, m := range metas {
		metaIDs = append(metaIDs, m.DocumentID)
	}

	err = fc.UpdatePlaque(ctx, plaque.DocumentID, []firestore.Update{{Path: "token_meta_id_list", Value: metaIDs}})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Added %v tokens to plaque for artist %s", len(metas), name)
}

func getScriptClients() (*viewer.Viewer, *fstore.FirestoreClient) {
	fc, err := fstore.NewFirestoreClient(context.Background(), "../serviceAccountKey.json")
	if err != nil {
		log.Fatal(err)
	}
	v := viewer.NewViewer(fc, nil)

	v.PlaqueFile = "../plaque.json"
	v.MediaDir = "../media"
	v.MetadataDir = "../metadata"

	return v, fc
}
