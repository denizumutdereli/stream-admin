package users

type UserKYCFile struct {
	ID        int64  `json:"id"`
	FileID    int64  `json:"file_id"`
	KYC_ID    int64  `json:"kyc_id"`
	FileName  string `json:"file_name"`
	FileType  string `json:"file_type"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type FilesOfKYCs struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	MimeType  string `json:"mime_type"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	IsPrivate int    `json:"is_private"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}
