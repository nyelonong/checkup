package checkup

import "errors"

var (
	ErrReadFile              = errors.New("error when reading file.")
	ErrHTTPCall              = errors.New("error when call http.")
	ErrInvalidStatusCode     = errors.New("error status code is invalid.")
	ErrFailedConnectPostgres = errors.New("error failed to connect to postgres.")
	ErrFailedConnectRedis    = errors.New("error failed to connect to redis.")
	ErrCheckupFailed         = errors.New("Some dependencies is down.")
	ErrFailedConnectGrpc     = errors.New("error failed to connect to grpc.")
)
