package fstore

import (
	"fmt"
	"path"
)

type TokenMeta struct {
	Name             string `json:"name" firestore:"name"`
	Artist           string `json:"artist" firestore:"artist"`
	Description      string `json:"description" firestore:"description"`
	PublicLink       string `json:"public_link" firestore:"public_link"`
	MediaID          string `json:"media_id" firestore:"media_id"`
	MediaType        string `json:"media_type" firestore:"media_type"`                 // file extension of media file, e.g. '.mp4'
	ExternalMediaURL string `json:"external_media_url" firestore:"external_media_url"` // url of source media file on external server (e.g. opensea servers)
}

type FirestoreTokenMeta struct {
	DocumentID string    `json:"document_id"`
	TokenMeta  TokenMeta `json:"token_meta"`
}

// MediaFileName returns name of media file, combination of media id and media type
// if no archive media is found (empty media id), return external media url base filename
func (tm *FirestoreTokenMeta) MediaFileName() string {
	if tm.TokenMeta.MediaID != "" {
		return fmt.Sprintf("%s%s", tm.TokenMeta.MediaID, tm.TokenMeta.MediaType)
	}
	return path.Base(tm.TokenMeta.ExternalMediaURL) // takes base filename from url
}

type Plaque struct {
	Name            string   `json:"name" firestore:"name"`
	WalletAddress   string   `json:"wallet_address" firestore:"wallet_address"`
	TokenMetaIDList []string `json:"token_meta_id_list" firestore:"token_meta_id_list"` // list of token meta document ids which the plaque will display
}

type FirestorePlaque struct {
	DocumentID string `json:"document_id"`
	Plaque     Plaque `json:"plaque"`
}

type FirestoreQuery struct {
	Path  string
	Op    string
	Value interface{}
}
