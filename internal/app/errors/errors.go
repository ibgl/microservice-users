package errors

type ErrorType struct {
	t string
}

var (
	ErrorTypeUnknown        = ErrorType{"unknown"}
	ErrorTypeAuthorization  = ErrorType{"authorization"}
	ErrorTypeIncorrectInput = ErrorType{"incorrect-input"}
	ErrorNotFound           = ErrorType{"not-found"}
)

type AppError struct {
	error     string
	slug      string
	errorType ErrorType
}

func (s AppError) Error() string {
	return s.error
}

func (s AppError) Slug() string {
	return s.slug
}

func (s AppError) ErrorType() ErrorType {
	return s.errorType
}

func NewAppError(error string, slug string) AppError {
	return AppError{
		error:     error,
		slug:      slug,
		errorType: ErrorTypeUnknown,
	}
}

func NewAuthorizationError(error string, slug string) AppError {
	return AppError{
		error:     error,
		slug:      slug,
		errorType: ErrorTypeAuthorization,
	}
}

func NewIncorrectInputError(error string, slug string) AppError {
	return AppError{
		error:     error,
		slug:      slug,
		errorType: ErrorTypeIncorrectInput,
	}
}

func NewNotFoundError(error string, slug string) AppError {
	return AppError{
		error:     error,
		slug:      slug,
		errorType: ErrorNotFound,
	}
}

func IsApp(err error) bool {
	_, ok := err.(AppError)
	return ok
}

func IsNotFound(err error) bool {
	app, ok := err.(AppError)
	if ok {
		return app.ErrorType() == ErrorNotFound
	}
	return false
}
