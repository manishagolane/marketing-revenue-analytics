package constants

type Status string
type USERID string
type RequestID string

const UserID USERID = "userID"
const REQUEST_ID RequestID = "requestID"
const LOGGED_IN_USER = "loggedInUser"

const (
	ApiSuccess Status = "success"
	ApiFailure Status = "failure"
)

const API_TOKEN string = "Authorization"

const (
	TIMEZONE = "Asia/Kolkata"
)
