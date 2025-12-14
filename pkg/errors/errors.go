package errors

import "fmt"

var (
	ErrInvalidPayload                 = fmt.Errorf("payload of event has a wrong type")
	ErrNumberOfRobots                 = fmt.Errorf("number of robots should be at least 2")
	ErrNegativeBufferSize             = fmt.Errorf("buffer size should be positive")
	ErrNegativePercentageOfLost       = fmt.Errorf("percentage of lost should be positive")
	ErrNegativePercentageOfDuplicated = fmt.Errorf("percentage of duplicated should be positive")
	ErrNegativeDuplicatedNumber       = fmt.Errorf("duplicated number should be positive")
	ErrNegativeMaxAttempts            = fmt.Errorf("max attempts should be positive")
	ErrWorkerPanic                    = fmt.Errorf("worker panic")
)
