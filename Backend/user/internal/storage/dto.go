package storage

type (
	// Key is "verifycode:<session code>"
	VeriFyCode struct {
		Email       string `json:"email"`
		VerifyCode  string `json:"verifycode"`
	}
)
