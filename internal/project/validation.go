package project

import (
	"errors"

	"github.com/google/uuid"
)

var errInvalidUUID = errors.New("Invalid UUID")

func ParseUUID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, errInvalidUUID
	}
	return id, nil
}
