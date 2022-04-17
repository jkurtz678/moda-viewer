package fstore

type TokenMeta struct {
	Name   string `json:"name"`
	Artist string `json:"artist"`
}

type FirestoreTokenMeta struct {
	DocumentID string    `json:"document_id"`
	TokenMeta  TokenMeta `json:"token_meta"`
}
