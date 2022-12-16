package switcherr

import "go/token"

type CaseType int

const (
	// CaseInvalid indicates an invalid case.
	CaseInvalid CaseType = iota
	// CaseErrorEqNil indicates that case has err == nil or nil == err.
	CaseErrorEqNil
	// CaseErrorIs indicates that case has error.Is and/or error.As functions.
	CaseErrorIs
	// CaseErrNeqNil indicates that case has err != nil or nil != err.
	CaseErrNeqNil
	// CaseNotErrorHandler indicates that case not checking errors.
	CaseNotErrorHandler
)

type ErrorType int

const (
	// ErrorTypeInvalid indicates an invalid error type.
	ErrorTypeInvalid ErrorType = iota

	// ErrorTypeIsAfterNeqNil indicates errors.Is/As coming after err != nil.
	ErrorTypeIsAfterNeqNil

	// ErrorTypeNeqNilAfterNonError indicates err != nil comes after non-error case checks.
	ErrorTypeNeqNilAfterNonError

	// ErrorTypeNoNeqNil indicates that errors.Is/As is present but err != nil is not.
	ErrorTypeNoNeqNil

	// ErrorTypeNoError indicates that the switch is not an error checking switch.
	ErrorTypeNoError
)

func (t ErrorType) String() string {
	switch t {
	case ErrorTypeInvalid:
		return "invalid type"
	case ErrorTypeIsAfterNeqNil:
		return "error type check comes after err != nil"
	case ErrorTypeNeqNilAfterNonError:
		return "case err != nil comes after non-error case"
	case ErrorTypeNoNeqNil:
		return "case err != nil is missing"
	case ErrorTypeNoError:
		return "no error"

	}
	return ""
}

type ErrorItem struct {
	Pos     token.Pos
	ErrType ErrorType
}
