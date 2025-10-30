package auth

//Requests
type (
	VerifyReq struct {
		Email string `json:"email" validate:"required,email"`
	}

	AuthenticationReq struct {
		Login       string `json:"login"`
		Password    string `json:"password"`
	}

	RegistrationReq struct {
		SessionCode string `json:"sessioncode"`
		Email       string `json:"email" validate:"required,email"`
		Login       string `json:"login"`
		Password    string `json:"password"`
		VerifyCode  string `json:"verifycode"`
	}
)

//Responses
type (
	VerifyRes struct {
		SessionCode string `json:"sessioncode"`
	}

	AuthenticationRes struct {
		Jwt string `json:"jwt"`
	}

	RegistrationRes struct {
		Id string `json:"id"`
	}
)
