package httpx

import (
	"errors"
	"time"
)

// 用于根据ID查询对象
type (
	IDRequest struct {
		ID uint64
	}
)

// 规范化HTTP Header Key: the first letter and any letter following a hyphen is upper case; the rest are lowercase
const (
	ApplicationJSON = "application/json;charset=UTF-8"
	ContentEncoding = "Content-Encoding"
	ContentSecurity = "X-Content-Security"
	ContentType     = "Content-Type"

	KeyField     = "key"
	SecretField  = "secret"
	TypeField    = "type"
	CryptionType = 1

	// contentTypeJSON is used to reject JSON responses when a file was expected.
	contentTypeJSON = "application/json"
)

const (
	// defaultTimeout is the default timeout for outbound download requests.
	defaultTimeout = 30 * time.Second
)

var (
	ErrInconsistentIDs = errors.New("inconsistent IDs")
	ErrAlreadyExists   = errors.New("already exists")
	ErrNotFound        = errors.New("not found")

	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)
