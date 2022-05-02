package fstore

type TokenMeta struct {
	Name        string `json:"name" firestore:"name"`
	Artist      string `json:"artist" firestore:"artist"`
	Description string `json:"description" firestore:"description"`
	PublicLink  string `json:"public_link" firestore:"public_link"`
}

type FirestoreTokenMeta struct {
	DocumentID string    `json:"document_id"`
	TokenMeta  TokenMeta `json:"token_meta"`
}

type Plaque struct {
	Name                    string   `json:"name" firestore:"name"`
	TokenMetaDocumentIDList []string `json:"token_document_meta_id_list" firestore:"token_meta_document_id_list"`
}

type FirestorePlaque struct {
	DocumentID string `json:"document_id"`
	Plaque     Plaque `json:"plaque"`
}
