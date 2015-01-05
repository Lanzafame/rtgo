package rtgo

type Message struct {
	Room    string `json:"room"`
	Event   string `json:"event"`
	Payload string `json:"payload"`
}

type DBMessage struct {
	DB    string `json:"db"`
	Table string `json:"table"`
	Key   string `json:"key"`
	Data  string `json:"data"`
}
