package handlers_models

type MsgType string

const (
    TextMsg    MsgType = "text"
    ImageMsg   MsgType = "image"
    SystemMsg  MsgType = "system"
	ItemMsg    MsgType = "item"
	CallMsg    MsgType = "call"
	LinkMsg    MsgType = "link"
	LocationMsg MsgType = "location"
    DeletedMsg MsgType = "deleted"
	AppCallMsg MsgType = "appCall"
	FileMsg    MsgType = "file"
	VideoMsg   MsgType = "video"
	VoiceMsg   MsgType = "voice"
)

type ChatType string

const (
    UserToItem ChatType = "u2i"
    UserToUser ChatType = "u2u"
)

type MsgContent struct {
	Text string `json:"text"`
}

type FromAvitoMsg struct {
	AuthorId int `json:"author_id"`
	ChatId string `json:"chat_id"`
	ChatType string `json:"chat_type"`
	Content MsgContent `json:"content"`
	Created int `json:"created"`
	Id string `json:"id"`
	ItemId *int `json:"item_id"`
	Read *int `json:"read"`
	Type string `json:"type"`
	UserId int `json:"user_id"`
}