package security

import (
	"github.com/google/uuid"
)

func UUID() string {
	id := uuid.New()
	return id.String()
}
