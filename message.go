package rtgo

// Message defines the structure of incoming JSON messages
// that do not perform DB functions.
type Message struct {
	Room    string `json:"room"`
	Event   string `json:"event"`
	Payload string `json:"payload"`
}

// DBMessage defines the structure of incoming JSON messages
// that do perform DB function.
type DBMessage struct {
	DB    string `json:"db"`
	Table string `json:"table"`
	Key   string `json:"key"`
	Data  string `json:"data"`
}
