package model

import "encoding/json"

type SeatalkGetBotSubscriberListResponse struct {
	Code        int         `json:"code"`
	NextCursor  string      `json:"next_cursor"`
	Subscribers Subscribers `json:"subscribers"`
}

type Subscribers struct {
	EmployeeCodes []string `json:"employee_code"`
}

type SendToSubsRequestBody struct {
	EmployeeCode string          `json:"employee_code"`
	Message      json.RawMessage `json:"message"`
}

type AccessTokenReqBody struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}
