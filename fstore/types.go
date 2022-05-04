package fstore

import "fmt"

type TokenMeta struct {
	Name        string `json:"name" firestore:"name"`
	Artist      string `json:"artist" firestore:"artist"`
	Description string `json:"description" firestore:"description"`
	PublicLink  string `json:"public_link" firestore:"public_link"`
	MediaID     string `json:"media_id" firestore:"media_id"`
	MediaType   string `json:"media_type" firestore:"media_type"` // file extension of media file, e.g. '.mp4'
}

type FirestoreTokenMeta struct {
	DocumentID string    `json:"document_id"`
	TokenMeta  TokenMeta `json:"token_meta"`
}

// MediaFileName returns name of media file, combination of media id and media type
func (tm *FirestoreTokenMeta) MediaFileName() string {
	return fmt.Sprintf("%s%s", tm.TokenMeta.MediaID, tm.TokenMeta.MediaType)
}

type Plaque struct {
	Name            string   `json:"name" firestore:"name"`
	TokenMetaIDList []string `json:"token_meta_id_list" firestore:"token_meta_id_list"` // list of token meta document ids which the plaque will display
}

type FirestorePlaque struct {
	DocumentID string `json:"document_id"`
	Plaque     Plaque `json:"plaque"`
}
