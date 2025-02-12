package avito_models

// sendMsgRequest
type Msg struct {
	Text string `json:"text"`
}

type ToAvitoMsg struct {
	Message Msg    `json:"message"`
	Type    string `json:"type"`
}

// sendMsgResponse
type ResponseContent struct {
	Text *string `json:"text"`
}

type Direction string

const (
	Out Direction = "out"
)

type ResponseMsgType string

const (
	Text ResponseMsgType = "text"
)

type SendMsgResponse struct {
	Content   ResponseContent `json:"content"`
	Created   int        `json:"created"`
	Direction Direction  `json:"direction"`
	Id       string     `json:"id"`
	MsgType   ResponseMsgType `json:"type"`
}

// readChatResponse
type OK bool

const (
	Ok OK = true
)

type ReadChatResponse struct {
	Ok OK `json:"ok"`
}

// getChatInfoResponse
type Value struct {
	Id int `json:"id"`
	PriceString string `json:"price_string"`
	StatusId int8 `json:"status_id"`
	Title string `json:"title"`
	Url string `json:"url"`
	UserId int `json:"user_id"`

} 

type Context struct {
	Type  string `json:"type"`
	Value Value `json:"value"`
}

type GetChatInfoResponse struct{
	Context Context `json:"context"`
}