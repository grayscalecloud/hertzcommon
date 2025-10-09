package model

type UserImageMqDto struct {
	UserID   int64
	TenantId string
}

type UserImageResultMqDto struct {
	UserID    int64  `json:"user_id"`
	AvatarUrl string `json:"avatar_url"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	TenantId  string
}

type ActivityMqDto struct {
	ActivityId string
	TenantId   string
}
