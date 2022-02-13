package http

var (
	ValidateFileParameter = validateFileParameter
	ParseLimit            = parseLimit
	CheckFile             = checkFile

	LogHandler = logHandler
)

type ErrorResponse errorResponse
type LogResponse logResponse
