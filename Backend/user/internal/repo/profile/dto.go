package profile

import "database/sql"

type ()

type (
	GetProfileRes struct {
		Surname    sql.NullString
		Name       sql.NullString
		Patronymic sql.NullString
		CreatedAt  string
	}
)
