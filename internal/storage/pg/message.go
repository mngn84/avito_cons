package pg

type Message struct {
    Id        int64
    ChatId    string
    UserId    int
    Content   string
    Role      string
    CreatedAt int
}

type GptMsg  struct {
    Role    string
    Content string
}

type DbRow struct {
    UserId    int
    ChatId    string
    Content   string
    Role      string
    CreatedAt int
}