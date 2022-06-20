package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"jkurtz678/moda-viewer/fstore"
	"jkurtz678/moda-viewer/viewer"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

var ctx = context.Background()

func main() {
	script := flag.String("s", "", "name of script to run, options are namePlaque, assignArtist, parseCSV")
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
	case "backupTokenMetas":
		backupTokenMetas()
	case "parseCSV":
		parseCSV(*name)
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

	metas, err := fc.GetTokenMetaByQuery(ctx, fstore.FirestoreQuery{Path: "artist", Op: "==", Value: name})
	if err != nil {
		log.Fatal(err)
	}

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
	fc, err := fstore.NewFirestoreClient(context.Background(), "./serviceAccountKey.json")
	if err != nil {
		log.Fatal(err)
	}
	v := viewer.NewViewer(fc, nil)

	v.PlaqueFile = "./plaque.json"
	v.MediaDir = "./media"
	v.MetadataDir = "./metadata"

	return v, fc
}

func parseCSV(filename string) {
	if filename == "" {
		log.Fatalf("error - no file name specified")
	}

	f, err := os.Open(filename)
	if err != nil {
		log.Fatal("Unable to read input file "+filename, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filename, err)
	}

	metas := make([]fstore.TokenMeta, 0, len(records)-1)

	for _, row := range records[1:] {
		artist := strings.Trim(row[0], " ")
		name := strings.Trim(row[1], " ")
		description := strings.Trim(row[2], " ")
		publicLink := strings.Trim(row[3], " ")
		mediaID := strings.Trim(row[4], " ")
		mediaType := ".mp4"
		log.Printf("artist %s", artist)
		log.Printf("name %s", name)
		log.Printf("public link %s", publicLink)
		log.Printf("media id %s", mediaID)

		meta := fstore.TokenMeta{
			Name:        name,
			Artist:      artist,
			Description: description,
			PublicLink:  publicLink,
			MediaID:     mediaID,
			MediaType:   mediaType,
		}
		metas = append(metas, meta)
	}

	log.Printf("metas %+v", metas)

	serviceAccountKey := "../serviceAccountKey.json"
	fstoreClient, err := fstore.NewFirestoreClient(context.Background(), serviceAccountKey)
	if err != nil {
		log.Fatalln(err)
	}

	for _, t := range metas {
		log.Printf("inserting token meta %s", t.Name)
		_, err := fstoreClient.CreateTokenMeta(ctx, &t)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("Successfully created %v token metas", len(metas))
}

func backupTokenMetas() {
	ctx := context.Background()
	client, err := fstore.NewFirestoreClient(ctx, "../serviceAccountKey.json")
	if err != nil {
		log.Fatal(err)
	}

	iter := client.Collection("token-meta").Documents(ctx)

	file, err := os.OpenFile(fmt.Sprintf("backup-%s.json", time.Now().Format("20060102150405")), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	defer file.Close()
	datawriter := bufio.NewWriter(file)
	defer datawriter.Flush()
	for {
		snap, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		tokenMeta := new(fstore.TokenMeta)
		err = snap.DataTo(tokenMeta)
		if err != nil {
			log.Fatal(err)
		}

		marshalMeta, err := json.Marshal(tokenMeta)
		if err != nil {
			log.Fatal(err)
		}

		_, err = datawriter.WriteString(string(marshalMeta) + "\n")
		if err != nil {
			log.Fatal(err)
		}
	}

}
