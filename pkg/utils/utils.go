package utils

import (
	"math/rand"
	"strings"
	"time"

	"github.com/oklog/ulid"
)

// GetULID gives us a new ulid
//
func GetULID() string {
	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	return ulid.MustNew(ulid.Timestamp(t), entropy).String()
}

// CloudFormationName fixes up names to be compatible with cloudformation
//
func CloudFormationName(name string) string {
	return strings.ToLower(strings.Replace(name, "-", "", -1))
}
