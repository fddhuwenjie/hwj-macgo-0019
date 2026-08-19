package repository

import "errors"

var (
	ErrNotFound           = errors.New("repository: aggregate not found")
	ErrVersionConflict    = errors.New("repository: version conflict")
	ErrInvalidAggregate   = errors.New("repository: invalid aggregate")
	ErrStoreClosed        = errors.New("repository: store closed")
	ErrTransactionAborted = errors.New("repository: transaction aborted")
	ErrDuplicateKey       = errors.New("repository: duplicate key")
	ErrInvalidDestination = errors.New("repository: invalid destination for query")
)
