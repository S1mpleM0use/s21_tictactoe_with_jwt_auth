package model

type SignUpRequest struct {
	Login    string
	Password string
}

type JwtRequest struct {
	Login    string
	Password string
}

type JwtResponse struct {
	Type         string
	AccessToken  string
	RefreshToken string
}

type RefreshJwtRequest struct {
	RefreshToken string
}
