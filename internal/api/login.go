package api

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	UserID      int64  `json:"user_id"`
	AccessToken string `json:"access_token"`
}
