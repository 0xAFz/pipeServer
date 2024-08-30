package services

type App struct {
	Account *AccountService
	Message *MessageService
}

func NewApp(
	Account *AccountService,
	Message *MessageService,
) *App {
	return &App{Account: Account, Message: Message}
}
