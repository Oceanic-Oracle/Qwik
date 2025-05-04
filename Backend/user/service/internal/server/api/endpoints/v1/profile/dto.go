package profile

type ()

type (
	GetProfileRes struct {
		Surname    string `json:"surname"`
		Name       string `json:"name"`
		Patronymic string `json:"patronymic"`
		CreatedAt  string `json:"created_at"`
	}
)
