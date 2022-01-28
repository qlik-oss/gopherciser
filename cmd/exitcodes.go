package cmd

// Unfortunately go doesn't support negative exit codes,
// so same logic as sdkexerciser can't be used (positive for error count and negative for other errors)
// as exit codes seems to be limited to 8 bits (even though defined as an int),
// we will use setting the highest bit as a representation of negative values,
// i.e. errors not corresponding to simulation errors. Exit codes with highest bit set to 0
// corresponds to error count from simulation, where 0x7F will represent >127 errors.
const (
	// ExitCodeExecutionError error during execution
	ExitCodeExecutionError int = 0x80 + iota
	// ExitCodeJSONParseError error parsing JSON config
	ExitCodeJSONParseError
	// ExitCodeJSONValidateError validating JSON config failed
	ExitCodeJSONValidateError
	// ExitCodeLogFormatError error resolving log format
	ExitCodeLogFormatError
	// ExitCodeObjectDefError error reading object definitions
	ExitCodeObjectDefError
	// ExitCodeProfilingError error starting profiling
	ExitCodeProfilingError
	// ExitCodeMetricError error starting prometheus
	ExitCodeMetricError
	// ExitCodeOsError error when interacting with host OS
	ExitCodeOsError
	// ExitCodeSummaryTypeError incorrect summary type
	ExitCodeSummaryTypeError
	// ExitCodeConnectionError error during test connection
	ExitCodeConnectionError
	// ExitCodeAppStructure error during get app structure
	ExitCodeAppStructure
	// ExitCodeMissingParameter
	ExitCodeMissingParameter
	// ExitCodeForceQuit
	ExitCodeForceQuit
	// ExitCodeMaxErrorsReached
	ExitCodeMaxErrorsReached
)
