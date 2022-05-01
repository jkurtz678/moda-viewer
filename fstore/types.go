package fstore

type TokenMeta struct {
	Name        string `json:"name"`
	Artist      string `json:"artist"`
	Description string `json:"description"`
}

type FirestoreTokenMeta struct {
	DocumentID string    `json:"document_id"`
	TokenMeta  TokenMeta `json:"token_meta"`
}

type Plaque struct {
	Name        string `json:"name" firestore:"name"`
	TokenMetaID string `json:"token_meta_id" firestore:"token_meta_id"`
	PlaylistID  string `json:"playlist" firestore:"playlist_id"`
}

type FirestorePlaque struct {
	DocumentID string `json:"document_id"`
	Plaque     Plaque `json:"plaque"`
}
