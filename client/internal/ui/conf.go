// Константы пакета.
package ui

// Размер рабочей области интерфейса.
const (
	// ширина экрана
	screenWidth = 130
	// высота экрана
	screenHeight = 30
)

// Имена окон
const (
	// Главный экран.
	viewMain = "main"
	// Экран регистрации.
	viewRegistration = "registration"
	// Экран аутентификации.
	viewAutentification = "autentification"
	// Окно настроек.
	viewSettings = "settings"
	// Окно с запросом дополнительного ключа шифрования.
	viewRequestSecretKey = "reqSecretKey"
	// Окно выбора типа записей.
	viewSelectType = "selectType"
	// Осно для логин/пароль
	viewLoginPasswordData = "typeLoginPassword"
	// Окно для текста.
	viewTextData = "typeText"
	// Окно для бинарных файлов.
	viewBinaryData = "typeBinary"
	// Окно для банковских карт.
	viewBankCardData = "typeBankCard"
)

// Стадии этапов restore и backUp.
const (
	// Стадия - нет активности.
	stageNotActive = 0
	// Стадия - процесс активен.
	stageActive = 1
	// Стадия - процесс завершён.
	stageOk = 2
	// Стадия - ошибка выполнения.
	stageFault = 3
)
