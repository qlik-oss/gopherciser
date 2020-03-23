package logger

const (
	// FieldAuthUser - User
	FieldAuthUser = "User"
	// FieldThread - Thread number
	FieldThread = "Thread"
	// FieldSession - session number
	FieldSession = "Session"
	// FieldSessionName - session name
	FieldSessionName = "SessionName"
	// FieldAppName - app name
	FieldAppName = "AppName"
	// FieldAppGUID - app GUID
	FieldAppGUID = "AppGUID"
	// FieldTick - high def time
	FieldTick = "Tick"
	// FieldAction - action name
	FieldAction = "Action"
	// FieldResponseTime - response time
	FieldResponseTime = "ResponseTime"
	// FieldSuccess - action success
	FieldSuccess = "Success"
	// FieldWarnings - warning count
	FieldWarnings = "Warnings"
	// FieldErrors - error count
	FieldErrors = "Errors"
	// FieldStack - stack trace
	FieldStack = "Stack"
	// FieldSent - bytes sent
	FieldSent = "Sent"
	// FieldReceived - bytes received
	FieldReceived = "Received"
	// FieldLabel - action label
	FieldLabel = "Label"
	// FieldActionID - action id number
	FieldActionID = "ActionId"
	// FieldObjectType - object type
	FieldObjectType = "ObjectType"
	// FieldDetails - details string
	FieldDetails = "Details"
	// FieldInfoType - type of info logging
	FieldInfoType = "InfoType"
	// FieldRequestsSent - request counter
	FieldRequestsSent = "RequestsSent"
	// FieldTime - logging time
	FieldTime = "time"
	// FieldTimestamp - to be used for time without timezone for G3 compliance
	FieldTimestamp = "timestamp"
	// FieldLevel - logging level
	FieldLevel = "level"
	// FieldMessage - message
	FieldMessage = "message"
)

var (
	//AllFields for logging (i.e. use as headers)
	AllFields = []string{FieldTime, FieldAction, FieldLabel, FieldActionID, FieldLevel, FieldInfoType, FieldMessage, FieldDetails, FieldSuccess, FieldResponseTime,
		FieldAppName, FieldAppGUID, FieldAuthUser, FieldThread, FieldSession, FieldSessionName, FieldTick,
		FieldObjectType, FieldWarnings, FieldErrors, FieldStack, FieldSent, FieldReceived, FieldRequestsSent, FieldTimestamp}
)
