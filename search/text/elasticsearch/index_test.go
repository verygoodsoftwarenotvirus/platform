package elasticsearch

type example struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type invalidJSON struct {
	Channel chan int `json:"channel"`
}
