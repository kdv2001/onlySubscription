package communication

// Message модель сообщения
type Message struct {
	// ChatID ID чата
	ChatID int64
	// Title заголовок
	Title string
	// Description описание
	Description string
}
