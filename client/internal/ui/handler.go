// Обработчика пакета.
package ui

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Part001-R/YaPr-GP-2/client/internal/service/udt"
	"github.com/Part001-R/YaPr-GP-2/client/internal/utils/flags"
	"github.com/jroimartin/gocui"
)

var once sync.Once // Разовая инициализация.

var inst *handlerUI // Экземпляр.

// Конструктор. Возвращается указатель на экземпляр.
//
// Параметры:
//
//	conf - указатель конфигурации сервиса.
func new(conf *udt.Configuration) *handlerUI {
	once.Do(func() {
		inst = &handlerUI{
			conf:   conf,
			typed:  typeData{},
			status: status{},
			view: screens{
				activeView:   "",
				currentFocus: "Login",
			},
			secret: encrKey{
				secretKey: [32]byte{},
			},
			data: data{
				encryptLoginPassword: []loginPassword{},
				loginPassword:        []loginPassword{},
				namesLoginPassword:   []string{},
			},
			index: indexes{
				loginPassword: 0,
			},
			mutex: mutex{
				restoreBackup:       sync.Mutex{},
				processTxRx:         sync.Mutex{},
				statusBackUp:        sync.Mutex{},
				statusRestore:       sync.Mutex{},
				statusPushContainer: sync.Mutex{},
				statusPopContainer:  sync.Mutex{},
				statusFileTx:        sync.Mutex{},
				statusFileRx:        sync.Mutex{},
				statusIsBusyServer:  sync.Mutex{},
			},
			txrx:       txrx{},
			clientName: "",
			tokenAuth:  "",
		}
	})
	return inst
}

// Главное окно в режиме - локальный. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
func layoutLocal(g *gocui.Gui) error {

	mainView, err := g.SetView(viewMain, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	mainView.Wrap = true
	mainView.Clear()

	fmt.Fprintln(mainView, strings.Repeat("\n", 3))
	fmt.Fprintf(mainView, "%s МЕНЕДЖЕР ПАРОЛЕЙ\n", strings.Repeat(" ", 54))

	fmt.Fprintln(mainView, strings.Repeat("\n", 2))
	fmt.Fprintf(mainView, "%s Режим работы - Локальный\n", strings.Repeat(" ", 50))

	fmt.Fprintln(mainView, strings.Repeat("\n", 12))
	fmt.Fprintf(mainView, "%s Регистрация     (Ctrl+A)\n", strings.Repeat(" ", 50))
	fmt.Fprintf(mainView, "%s Аутентификация  (Ctrl+B)\n", strings.Repeat(" ", 50))
	fmt.Fprintf(mainView, "%s Настройки       (Ctrl+D)\n", strings.Repeat(" ", 50))
	fmt.Fprintln(mainView, "")
	fmt.Fprintf(mainView, "%s Выход           (Ctrl+C)\n", strings.Repeat(" ", 50))

	return nil
}

// Главное окно в режиме - удалённый. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
func layoutRemote(g *gocui.Gui) error {

	mainView, err := g.SetView(viewMain, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	mainView.Wrap = true
	mainView.Clear()

	fmt.Fprintln(mainView, strings.Repeat("\n", 3))
	fmt.Fprintf(mainView, "%s МЕНЕДЖЕР ПАРОЛЕЙ\n", strings.Repeat(" ", 54))

	fmt.Fprintln(mainView, strings.Repeat("\n", 2))
	fmt.Fprintf(mainView, "%s Режим работы - Удалённый\n", strings.Repeat(" ", 50))

	fmt.Fprintln(mainView, strings.Repeat("\n", 12))
	fmt.Fprintf(mainView, "%s Регистрация     (Ctrl+A)\n", strings.Repeat(" ", 50))
	fmt.Fprintf(mainView, "%s Аутентификация  (Ctrl+B)\n", strings.Repeat(" ", 50))
	fmt.Fprintf(mainView, "%s Настройки       (Ctrl+D)\n", strings.Repeat(" ", 50))
	fmt.Fprintln(mainView, "")
	fmt.Fprintf(mainView, "%s Выход           (Ctrl+C)\n", strings.Repeat(" ", 50))

	return nil
}

// Главное окно. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	v - указатель на View.
func (c *handlerUI) showMain(g *gocui.Gui, v *gocui.View) error {

	c.conf.LgrFile.Write("Info: Нажата комбинация Ctrl+H")

	// Запрет активности при активности процессов передачи файлов.
	if layerShowMainIsRestraintRun(c) {
		return nil
	}

	// Сброс признаков.
	if err := layerShowMainReset(c); err != nil {
		return fmt.Errorf("функция layerShowMainReset, вернула ошибку: <%w>", err)
	}

	// Удаление видов.
	if err := layerShowMainClear(c, g); err != nil {
		return fmt.Errorf("функция layerShowMainClear, вернула ошибку: <%w>", err)
	}

	// Отображение главного меню.
	err := layoutLocal(g)
	if err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("Error: функция layout вернула ошибку: <%v>", err))
		return fmt.Errorf("функция layout, вернула ошибку: <%w>", err)
	}

	// Установка фокуса.
	if err := layerShowMainSetFocus(c); err != nil {
		return fmt.Errorf("функция layerShowMainSetFocus, вернула ошибку: <%w>", err)
	}

	return nil
}

// Регистрация. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	_  - заглушка на View.
func (c *handlerUI) showRegistration(g *gocui.Gui, _ *gocui.View) error {

	c.conf.LgrFile.Write("Info: Нажата комбинация Ctrl+A")

	// Ограничение вызова окна.
	if layerShowRegistrationIsRestraintRun(c) {
		return nil
	}

	// Сбросы.
	if err := layerShowRegistrationReset(c); err != nil {
		return fmt.Errorf("функция layerShowRegistrationIsReset, вернула ошибку: <%w>", err)
	}

	// Очистка.
	if err := deleteViews(g); err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}

	// Создание контейнера регистрации.
	window, err := layerShowRegistrationDrawWindow(c, g)
	if err != nil {
		return fmt.Errorf("функция layerShowRegistrationDrowWindow, вернула ошибку: <%w>", err)
	}

	// Отрисовка полей ввода.
	if err := layerShowRegistrationDrawInput(c, g, window); err != nil {
		return fmt.Errorf("функция layerShowRegistrationDrowInput, вернула ошибку: <%w>", err)
	}

	// Отрисовка индикаторов.
	if err := layerShowRegistrationDrawIndicator(c, g); err != nil {
		return fmt.Errorf("функция layerShowRegistrationDrowIndicator, вернула ошибку: <%w>", err)
	}

	// Отрисовка пояснений.
	if err := layerShowRegistrationDrawGuide(c, g); err != nil {
		return fmt.Errorf("функция layerShowRegistrationDrowGuide, вернула ошибку: <%w>", err)
	}

	// Установка фокуса.
	if err := layerShowRegistrationSetFocus(c, g, "Login"); err != nil {
		return fmt.Errorf("функция layerShowRegistrationSetFocus, вернула ошибку: <%w>", err)
	}

	c.conf.LgrFile.Write("Info: Окно регистрации, отображено.")

	return nil
}

// Аутентификация. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	_ - заглушка для View.
func (c *handlerUI) showAuthentication(g *gocui.Gui, _ *gocui.View) error {

	c.conf.LgrFile.Write("Info: Нажата комбинация Ctrl+B")

	// Ограничение вызова окна.
	if layerShowAuthenticationIsRestraintRun(c) {
		return nil
	}

	// Подключение к БД.
	// При срабатывании сторожевого таймера, подключение сбрасывается.
	if c.conf.Flag.Mode == flags.ModeLocal {
		if err := layerShowAuthenticationIsNewInstDB(c); err != nil {
			return fmt.Errorf("функция layerShowAuthenticationIsNewInstDB, вернула ошибку: <%w>", err)
		}
	}

	// Сбросы
	if err := layerShowAuthenticationReset(c); err != nil {
		return fmt.Errorf("функция layershowAuthenticationReset, вернула ошибку: <%w>", err)
	}

	// Очистка экрана.
	if err := layerShowAuthenticationClear(c, g); err != nil {
		return fmt.Errorf("функция layershowAuthenticationClear, вернула ошибку: <%w>", err)
	}

	// Создание окна.
	window, err := layerShowAuthenticationDrawWindow(c, g)
	if err != nil {
		return fmt.Errorf("функция layerShowAuthenticationDrawWindow, вернула ошибку: <%w>", err)
	}

	// отрисовка полей ввода.
	if err := layerShowAuthenticationDrawInput(c, g, window); err != nil {
		return fmt.Errorf("функция layerShowAuthenticationDrawInput, вернула ошибку: <%w>", err)
	}

	// Отрисовка индикаторов.
	if err := layerShowAuthenticationDrawIndicator(c, g); err != nil {
		return fmt.Errorf("функция layerShowAuthenticationDrawIndicator, вернула ошибку: <%w>", err)
	}

	// Отрисовка пояснений.
	if err := layerShowAuthenticationDrawGuide(c, g); err != nil {
		return fmt.Errorf("функция layerShowAuthenticationDrawGuide, вернула ошибку: <%w>", err)
	}

	// Установка фокуса.
	if err := layerShowAuthenticationSetFocus(c, g, "Login"); err != nil {
		return fmt.Errorf("функция layerShowAuthenticationSetFocus, вернула ошибку: <%w>", err)
	}

	// Создание экземпляра сервера.
	if err := layerShowAuthenticationNewInst(c); err != nil {
		return fmt.Errorf("функция layerShowAuthenticationNewInst, вернула ошибку: <%w>", err)
	}

	return nil
}

// Настройки. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	_ - заглушка для View.
func (c *handlerUI) showSettings(g *gocui.Gui, _ *gocui.View) error {

	c.conf.LgrFile.Write("Info: Нажата комбинация Ctrl+D")

	// Ограничение вызова окна.
	if layerShowSettingsIsRestraintRun(c) {
		return nil
	}

	// Очистка экрана.
	if err := layerShowSettingsClear(c, g); err != nil {
		return fmt.Errorf("функция layerShowSettingsClear, вернула ошибку: <%w>", err)
	}

	// Сброс состояния
	if err := layerShowSettingsReset(c); err != nil {
		return fmt.Errorf("функция layerShowSettingsReset, вернула ошибку: <%w>", err)
	}

	// Создание окна
	window, err := layerShowSettingsDrawWindow(c, g)
	if err != nil {
		return fmt.Errorf("функция layerShowSettingsDrawWindow, вернула ошибку: <%w>", err)
	}

	// Поля ввода
	if err := layerShowSettingsDrawInput(c, g, window); err != nil {
		return fmt.Errorf("функция layerShowSettingsDrawInput, вернула ошибку: <%w>", err)
	}

	// Пояснение по навигации.
	if err := layerShowSettingsDrawGuide(c, g, window); err != nil {
		return fmt.Errorf("функция layerShowSettingsDrawGuide, вернула ошибку: <%w>", err)
	}

	// Установка фокуса.
	if err := layerShowSettingsSetFocus(c, g, "IP"); err != nil {
		return fmt.Errorf("функция layerShowSettingsSetFocus, вернула ошибку: <%w>", err)
	}

	return nil
}

// Перевод фокуса. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	_ - заглушка для View.
func (c *handlerUI) nextFocus(g *gocui.Gui, v *gocui.View) error {

	c.conf.LgrFile.Write("Info: Нажат Tab")

	if c.status.statusWDT {
		c.ch.resetWDT <- struct{}{} // Сброс таймера.
	}

	// Перевод фокуса
	switch c.view.activeView {
	case viewRegistration: // Окно регистрации.
		switch c.view.currentFocus {
		case "Login":
			c.view.currentFocus = "Password-1"
		case "Password-1":
			c.view.currentFocus = "Password-2"
		case "Password-2":
			c.view.currentFocus = "Login"
		default:
		}

	case viewAutentification: // Окно аутентификации.
		switch c.view.currentFocus {
		case "Login":
			c.view.currentFocus = "Password-1"
		case "Password-1":
			c.view.currentFocus = "Login"
		default:
		}

	case viewSettings: // Окно настроек.
		switch c.view.currentFocus {
		case "IP":
			c.view.currentFocus = "Port"
		case "Port":
			c.view.currentFocus = "IP"
		default:
		}

	case viewSelectType: // Окно выбора типа данных.
		switch c.view.currentFocus {
		case "selectLoginPassword":
			c.view.currentFocus = "selectText"
		case "selectText":
			c.view.currentFocus = "SelectBinary"
		case "SelectBinary":
			c.view.currentFocus = "SelectBankCard"
		case "SelectBankCard":
			c.view.currentFocus = "selectLoginPassword"
		default:
		}

	case viewLoginPasswordData: // Окно  логин/пароль.
		switch c.view.currentFocus {
		case "fieldAddFor":
			c.view.currentFocus = "fieldAddLogin"
		case "fieldAddLogin":
			c.view.currentFocus = "fieldAddPassword"
		case "fieldAddPassword":
			c.view.currentFocus = "fieldAddFor"
		default:
		}

	case viewTextData: // Окно текст.
		switch c.view.currentFocus {
		case "fieldAddFor":
			c.view.currentFocus = "fieldAddText"
		case "fieldAddText":
			c.view.currentFocus = "fieldAddFor"
		default:
		}

	case viewBankCardData: // Окно банковских карт.
		switch c.view.currentFocus {
		case "fieldAddFor":
			c.view.currentFocus = "fieldAddOwner"
		case "fieldAddOwner":
			c.view.currentFocus = "fieldAddNumber"
		case "fieldAddNumber":
			c.view.currentFocus = "fieldAddValid"
		case "fieldAddValid":
			c.view.currentFocus = "fieldAddCode"
		case "fieldAddCode":
			c.view.currentFocus = "fieldAddFor"
		default:
		}

	case viewBinaryData: // Окно файлов.
		switch c.view.currentFocus {
		case "fieldPathSource":
			c.view.currentFocus = "fieldPathTarget"
		case "fieldPathTarget":
			c.view.currentFocus = "fieldPathSource"
		default:
		}

	default:
		return nil
	}

	_, err := g.SetCurrentView(c.view.currentFocus)
	if err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("Ошибка в функции SetCurrentView: <%v>", err))
		return fmt.Errorf("Ошибка в функции SetCurrentView: <%w>", err)
	}

	// Обновление
	var fields []string
	if c.view.activeView == viewRegistration {
		fields = []string{"Login", "Password-1", "Password-2"}
	}
	if c.view.activeView == viewAutentification {
		fields = []string{"Login", "Password-1"}
	}
	if c.view.activeView == viewSettings {
		fields = []string{"IP", "Port"}
	}
	if c.view.activeView == viewSelectType {
		fields = []string{"selectLoginPassword", "selectText", "SelectBinary", "SelectBankCard"}
	}
	if c.view.activeView == viewLoginPasswordData {
		fields = []string{"fieldAddFor", "fieldAddLogin", "fieldAddPassword"}
	}
	if c.view.activeView == viewTextData {
		fields = []string{"fieldAddFor", "fieldAddText"}
	}
	if c.view.activeView == viewBankCardData {
		fields = []string{"fieldAddFor", "fieldAddOwner", "fieldAddNumber", "fieldAddValid", "fieldAddCode"}
	}
	if c.view.activeView == viewBinaryData {
		fields = []string{"fieldPathSource", "fieldPathTarget"}
	}

	for _, name := range fields {
		view, err := g.View(name)
		if err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
			return fmt.Errorf("Не удалось установить фокус: <%w>", err)
		}
		if view != nil {
			c.setFocusStyle(view, name)
		}
	}

	// Установка курсора в конец текущей строки
	currentView, err := g.View(c.view.currentFocus)
	if err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("Ошибка в функции View: <%v>", err))
		return fmt.Errorf("Ошибка в функции View: <%w>", err)
	}
	if currentView != nil {
		buffer := strings.TrimSuffix(currentView.Buffer(), "\n")
		cursorX := len(buffer)
		currentView.SetCursor(cursorX, 0)
	}

	return nil
}

// Выход. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	_ - заглушка для View.
func (c *handlerUI) quit(g *gocui.Gui, _ *gocui.View) error {

	c.conf.LgrFile.Write("Info: Нажата комбинация Ctrl+C")

	// Ожидание завершения активных процессов.
	for {
		if c.getStatusBackUp() != stageActive &&
			c.getStatusRestore() != stageActive &&
			c.getStatusPopContainer() != stageActive &&
			c.getStatusPushContainer() != stageActive &&
			c.getStatusFileRx() != stageActive &&
			c.getStatusFileTx() != stageActive {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	return gocui.ErrQuit
}

// Обработка нажатия Enter. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	v - указатель на View.
func (c *handlerUI) handleEnter(g *gocui.Gui, v *gocui.View) error {

	c.conf.LgrFile.Write("Info: Нажат Enter")

	// Сброс сторожевого таймера.
	if c.status.statusWDT {
		c.ch.resetWDT <- struct{}{} // Сброс таймера.
	}

	// Запрет активности при активности процессов передачи файлов.
	if c.status.backUp == stageActive || c.status.restore == stageActive {
		return nil
	}

	switch c.view.activeView {

	// Окно регистрации.
	case viewRegistration:
		if err := enterViewRegistration(v, c); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция enterViewRegistration, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция enterViewRegistration, вернула ошибку: <%v>", err)
		}

	// Окно аутентификации.
	case viewAutentification:
		if err := enterViewAutentification(v, c); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция enterViewAutentification, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция enterViewAutentification, вернула ошибку: <%v>", err)
		}

	// Окно настроек.
	case viewSettings:
		if err := enterViewSettings(v, g, c); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция enterViewSettings, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция enterViewSettings, вернула ошибку: <%v>", err)
		}

	// Окно с запросом дополнительного ключа шифрования.
	case viewRequestSecretKey:
		if err := enterViewRequestSecretKey(v, g, c); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция enterViewRequestSecretKey, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция enterViewRequestSecretKey, вернула ошибку: <%v>", err)
		}

	// Окно с выбором типа данных.
	case viewSelectType:
		if err := enterViewSelectType(v, g, c); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция enterViewSelectType, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция enterViewSelectType, вернула ошибку: <%v>", err)
		}

	// Окно взаимодействия с логин/пароль.
	case viewLoginPasswordData:
		if err := enterViewLoginPasswordData(v, c); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция enterViewLoginPasswordData, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция enterViewLoginPasswordData, вернула ошибку: <%v>", err)
		}

	// Окно взаимодействия с текстом.
	case viewTextData:
		if err := enterViewTextData(v, c); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция enterViewTextData, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция enterViewTextData, вернула ошибку: <%v>", err)
		}

	// Окно взаимодействия с банковскими картами.
	case viewBankCardData:
		if err := enterViewBankCardData(v, c); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция enterViewBankCardData, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция enterViewBankCardData, вернула ошибку: <%v>", err)
		}

	// Окно взаимодействия с файлами.
	case viewBinaryData:
		if err := enterViewBinaryData(v, c); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция enterViewBinaryData, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция enterViewBinaryData, вернула ошибку: <%v>", err)
		}

	default:
	}

	return nil
}

// Проверка совпадения паролей при регистрации. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
func (c *handlerUI) indicators(g *gocui.Gui) error {

	switch c.view.activeView {

	// Окно регистрации.
	case viewRegistration:
		if err := indicatorViewRegistration(g, c); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция indicatorViewRegistration, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция indicatorViewRegistration, вернула ошибку: <%w>", err)
		}

	// Окно настроек.
	case viewSettings:
		if err := indicatorViewSettings(g, c); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция indicatorViewSettings, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция indicatorViewSettings, вернула ошибку: <%v>", err)
		}

	// Окно логин/пароль
	case viewLoginPasswordData:
		if err := indicatorViewLoginPasswordData(g, c); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция indicatorViewLoginPasswordData, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция indicatorViewLoginPasswordData, вернула ошибку: <%v>", err)
		}

	// Окно текста.
	case viewTextData:
		if err := indicatorViewTextData(g, c); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция indicatorViewTextData, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция indicatorViewTextData, вернула ошибку: <%v>", err)
		}

	// Окно банковских карт.
	case viewBankCardData:
		if err := indicatorViewBankCardData(g, c); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция indicatorViewBankCardData, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция indicatorViewBankCardData, вернула ошибку: <%v>", err)
		}

	// Окно файлов.
	case viewBinaryData:
		if err := indicatorViewBinaryData(g, c); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция indicatorViewBankCardData, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция indicatorViewBankCardData, вернула ошибку: <%v>", err)
		}

	// Окно выбора типов.
	case viewSelectType:
		if err := indicatorViewSelectType(g, c); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция indicatorViewSelectType, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция indicatorViewSelectType, вернула ошибку: <%v>", err)
		}

	default:
	}
	return nil
}

// Проверка связи с сервером. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	_ - заглушка для View.
func (c *handlerUI) testConnect(g *gocui.Gui, _ *gocui.View) error {

	c.conf.LgrFile.Write("Info: Нажата комбинация Ctrl+N")

	// Запуск проверки связи с сервером.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Выполнение проверки связи.
	ok, err := pingContext(ctx, c)
	if err != nil {
		c.status.checkConnectStatus = false
		c.conf.LgrFile.Write(fmt.Sprintf("Error: функция pingContext, вернула ошибку: <%v>", err))
		return nil
	}

	// Результат.
	c.status.checkConnectStatus = ok
	c.conf.LgrFile.Write(fmt.Sprintf("Info: проверка связи с %s:%s пройдена", c.typed.ip, c.typed.port))
	return nil
}

// Обновляет стиль вида в зависимости от того, имеет ли он фокус
//
// Параметры:
//
//	g - указатель на Gui.
//	name - имя элемента.
func (c *handlerUI) setFocusStyle(v *gocui.View, name string) {
	if name == c.view.currentFocus {
		// Активное поле: яркая рамка + контрастное выделение текста
		v.Frame = true
		v.Highlight = true
		v.SelFgColor = gocui.ColorBlack
		v.SelBgColor = gocui.ColorCyan
		v.FgColor = gocui.ColorCyan
		v.BgColor = gocui.ColorBlack // ← ИЗМЕНЕНО: чёрный фон
	} else {
		// Неактивное поле: сдержанный стиль
		v.Frame = true
		v.Highlight = false
		v.SelFgColor = gocui.ColorWhite
		v.SelBgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.BgColor = gocui.ColorBlack // ← ИЗМЕНЕНО: чёрный фон
	}
}

// Запуск процесса регистрации нового пользователя. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	v - указатель на View.
func (c *handlerUI) doRegistrationUser(g *gocui.Gui, v *gocui.View) error {

	// Если режим - Локальный.
	if c.conf.Flag.Mode == flags.ModeLocal {
		err := doRegistrationUserLocal(c)
		if err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: функция doRegistrationUserLocal, вернула ошибку: <%v>", err))
			return nil
		}
	}

	// Если режим - Удалённый.
	if c.conf.Flag.Mode == flags.ModeRemote {
		err := doRegistrationUserRemote(c)
		if err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: функция doRegistrationUserRemote, вернула ошибку: <%v>", err))
			return nil
		}
	}

	return nil
}

// Запуск процесса аутентификации пользователя. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	v - указатель на View.
func (c *handlerUI) doAuthenticationUser(g *gocui.Gui, v *gocui.View) error {

	c.conf.LgrFile.Write("Info: Нажата комбинация Ctrl+L")

	// Если окно аутентификации.
	if c.view.activeView == viewAutentification {

		// Если режим - локальный.
		if c.conf.Flag.Mode == flags.ModeLocal {
			if err := doAuthenticationUserModeLocal(c); err != nil {
				c.conf.LgrFile.Write(fmt.Sprintf("Error: функция doAuthenticationUserModeLocal, вернуля ошибку: <%v>", err))
				return nil
			}
			c.conf.LgrFile.Write(fmt.Sprintf("Info: пользователь <%s>, прошел аутентификацию. Режим - локальный", c.typed.login))
		}

		// Если режим - удалённый.
		if c.conf.Flag.Mode == flags.ModeRemote {
			if err := doAuthenticationUserModeRemote(c); err != nil {
				c.conf.LgrFile.Write(fmt.Sprintf("Error: функция doAuthenticationUserModeRemote, вернуля ошибку: <%v>", err))
				return nil
			}
			c.conf.LgrFile.Write(fmt.Sprintf("Info: пользователь <%s>, прошел аутентификацию. Режим - удалённый", c.typed.login))
		}

		// Открытие окна, с запросом ввода дополнительного кода шифрования.
		if err := c.showRequestEncryptKey(g, v); err != nil {
			return fmt.Errorf("Error: функция showRequestEncryptKey, вернуля ошибку: <%v>", err)
		}
	}

	return nil
}

// Запуск процесса сохранения данных. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	v - указатель на View.
func (c *handlerUI) doStore(g *gocui.Gui, v *gocui.View) error {

	c.conf.LgrFile.Write("Info: Нажата комбинация Ctrl+F")

	// Сброс сторожевого таймера.
	if c.status.statusWDT {
		c.ch.resetWDT <- struct{}{} // Сброс таймера.
	}

	switch c.view.activeView {
	case viewLoginPasswordData: // Окно - логин/пароль
		if err := doStoreViewLoginPasswordData(c); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция doStoreViewLoginPasswordData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewTextData: // Окно - текст.
		if err := doStoreViewTextData(c); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция doStoreViewTextData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewBankCardData: // Окно - банковские карты.
		if err := doStoreViewBankCardData(c); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция doStoreViewBankCardData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewBinaryData: // Окно - файлы.
		if err := doStoreViewBinaryData(c); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция doStoreViewBinaryData, вернула ошибку: <%v>", err))
			return nil
		}

	default:
	}

	return nil
}

// Отображение следующего элемента. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	v - указатель на View.
func (c *handlerUI) doShowNextElement(g *gocui.Gui, v *gocui.View) error {

	c.conf.LgrFile.Write("Info: Нажата комбинация Ctrl+E")

	// Сброс сторожевого таймера.
	if c.status.statusWDT {
		c.ch.resetWDT <- struct{}{} // Сброс таймера.
	}

	switch c.view.activeView {
	case viewLoginPasswordData: // Окно логин/пароль
		if err := doShowNextElementViewLoginPasswordData(c, g); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция doShowNextElementViewLoginPasswordData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewTextData: // Окно текста
		if err := doShowNextElementViewTextData(c, g); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция doShowNextElementViewTextData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewBankCardData: // Окно банковских карт
		if err := doShowNextElementViewBankCardData(c, g); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция doShowNextElementViewBankCardData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewBinaryData: // Окно файлов
		if err := doShowNextElementViewBinaryData(c, g); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция doShowNextElementViewBinaryData, вернула ошибку: <%v>", err))
			return nil
		}

	default:
	}

	return nil
}

// Отображение предыдущего элемента. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	v - указатель на View.
func (c *handlerUI) doShowPrevElement(g *gocui.Gui, v *gocui.View) error {

	c.conf.LgrFile.Write("Info: Нажата комбинация Ctrl+G")

	// Сброс сторожевого таймера.
	if c.status.statusWDT {
		c.ch.resetWDT <- struct{}{} // Сброс таймера.
	}

	switch c.view.activeView {
	case viewLoginPasswordData: // Окно логин/пароль.
		if err := doShowPrevElementViewLoginPasswordData(c, g); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция doShowPrevElementViewLoginPasswordData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewTextData: // Окно текста.
		if err := doShowPrevElementViewTextData(c, g); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция doShowPrevElementViewTextData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewBankCardData: // Окно банковских карт.
		if err := doShowPrevElementViewBankCardData(c, g); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция doShowPrevElementViewBankCardData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewBinaryData: // Окно файлов.
		if err := doShowPrevElementViewBinaryData(c, g); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция doShowPrevElementViewBinaryData, вернула ошибку: <%v>", err))
			return nil
		}

	default:
	}

	return nil
}

// Удаление записи. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	v - указатель на View.
func (c *handlerUI) doDeleteElement(g *gocui.Gui, v *gocui.View) error {

	c.conf.LgrFile.Write("Info: Нажата комбинация Ctrl+J")

	// Сброс сторожевого таймера.
	if c.status.statusWDT {
		c.ch.resetWDT <- struct{}{} // Сброс таймера.
	}

	switch c.view.activeView {
	case viewLoginPasswordData: // Окно логин/пароль
		if err := doDeleteElementViewLoginPasswordData(c, g); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция doDeleteElementViewLoginPasswordData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewTextData: // Окно текста
		if err := doDeleteElementViewTextData(c, g); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция doDeleteElementViewTextData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewBankCardData: // Окно банковских карт
		if err := doDeleteElementViewBankCardData(c, g); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция doDeleteElementViewBankCardData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewBinaryData: // Окно файлов.
		if err := doDeleteElementViewBinaryData(c, g); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция doDeleteElementViewBinaryData, вернула ошибку: <%v>", err))
			return nil
		}

	default:
	}

	return nil
}

// Извлечение. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	v - указатель на View.
func (c *handlerUI) doExtract(g *gocui.Gui, v *gocui.View) error {

	c.conf.LgrFile.Write("Info: Нажата комбинация Ctrl+K")

	// Сброс сторожевого таймера.
	if c.status.statusWDT {
		c.ch.resetWDT <- struct{}{} // Сброс таймера.
	}

	// Если режим - локальный.
	if c.conf.Flag.Mode == flags.ModeLocal {

		switch c.view.activeView {
		case viewBinaryData: // Окно работы с файлами

			if c.getStatusPopContainer() == stageNotActive && c.getStatusPushContainer() == stageNotActive {

				c.updateStatusPopContainer(stageActive)

				c.txrx.passedKB = 0
				c.txrx.percentTxRx = 0
				c.txrx.totalSizeKB = 0

				// Чтение буфера.
				v, err := g.View("fieldShowFor")
				if err != nil {
					c.conf.LgrFile.Write(fmt.Sprintf("Error: ошибка доступа к элементу fieldShowFor: <%v>", err))
					return nil
				}
				nameFile := v.ViewBuffer()
				nameFile = strings.ReplaceAll(nameFile, "\n", "")

				// Получение файла из хранилища.
				chErr := make(chan error)
				chDone := make(chan struct{})
				chData := make(chan []byte)
				chPercent := make(chan float64)
				chBreak := make(chan struct{})

				// Чтение файла.
				go c.conf.Container.GetFileFromContainer(nameFile, c.secret.secretKey, chPercent, chErr, chDone, chData, chBreak)

				// Приём данных и сборка файла.
				go bufferProcessPopContainer(nameFile, c, chPercent, chErr, chDone, chData, chBreak)
			}
		}
		return nil
	}

	// Если режим - удалённый.
	if c.conf.Flag.Mode == flags.ModeRemote {

		switch c.view.activeView {
		case viewBinaryData: // Окно работы с файлами

			if c.getStatusFileRx() == stageNotActive && c.getStatusFileTx() == stageNotActive {

				// Установка признака, что процесс приёма активный.
				c.updateStatusFileRx(stageActive)

				// Чтение имени запрашиваемого файла.
				v, err := g.View("fieldShowFor")
				if err != nil {
					c.conf.LgrFile.Write(fmt.Sprintf("Error: ошибка доступа к элементу fieldShowFor: <%v>", err))
					return nil
				}
				fileName := v.ViewBuffer()
				fileName = strings.ReplaceAll(fileName, "\n", "")

				// Запрос у сервера информации по файлу.
				_, _, rxFileSize, err := c.conf.Server.RequestFileInfo(c.tokenAuth, c.clientName, fileName)
				if err != nil {
					c.conf.LgrFile.Write(fmt.Sprintf("Error: функция RequestFileInfo, вернула ошибку: <%v>", err))
					return nil
				}

				rxFileSize = rxFileSize / 1024 // Получение КБайт

				// Подготовка данных для реализации запроса файла.
				if err := c.conf.Server.InitDataRequestFileByName(fileName, c.typed.dataPathTrg, c.tokenAuth, c.clientName, rxFileSize, 0, c.secret.secretKey); err != nil {
					c.conf.LgrFile.Write(fmt.Sprintf("Error: функция InitDataRequestFileByName, вернула ошибку: <%v>", err))
					return nil
				}

				chProcess := make(chan float32)
				chErr := make(chan error)
				chDone := make(chan struct{})

				c.conf.LgrFile.Write(fmt.Sprintf("Info: запуск процесса получения файла: <%s>", fileName))

				// Приём файла.
				go c.conf.Server.RequestFileByName(chProcess, chErr, chDone)

				// Приём данных процесса.
				go bufferProcessRxFileByName(c, chProcess, chErr, chDone)
			}
		}
		return nil
	}

	return nil
}

// Окно с запросом ввода дополнительного ключа шифрования. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	_ - заглушка на View.
func (c *handlerUI) showRequestEncryptKey(g *gocui.Gui, _ *gocui.View) error {

	c.conf.LgrFile.Write("Info: Запущен обработчик запроса серкетного ключа")

	// Сброс переменных.
	if err := layerShowRequestEncryptKeyReset(c); err != nil {
		return fmt.Errorf("функция layerShowRequestEncryptKeyReset, вернула ошибку: <%w>", err)
	}

	// Очистка экрана.
	if err := layerShowRequestEncryptKeyClear(c, g); err != nil {
		return fmt.Errorf("функция layerShowRequestEncryptKeyClear, вернула ошибку: <%w>", err)
	}

	// Создание окна.
	window, err := layerShowRequestEncryptKeyWindow(c, g)
	if err != nil {
		return fmt.Errorf("функция layerShowRequestEncryptKeyWindow, вернула ошибку: <%w>", err)
	}

	// Отрисовка полей ввода.
	if err := layerShowRequestEncryptKeyDrawInput(c, g); err != nil {
		return fmt.Errorf("функция layerShowRequestEncryptKeyDrawInput, вернула ошибку: <%w>", err)
	}

	// Отрисовка пояснений.
	if err := layerShowRequestEncryptKeyDrawGuide(c, g, window); err != nil {
		return fmt.Errorf("функция layerShowRequestEncryptKeyDrawGuide, вернула ошибку: <%w>", err)
	}

	// Установка фокуса.
	if err := layerShowRequestEncryptKeySetFocus(c, g, "scrtKey"); err != nil {
		return fmt.Errorf("функция layerShowRequestEncryptKeySetFocus, вернула ошибку: <%w>", err)
	}

	return nil
}

// Окно с выбором типа записей. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	_ - заглушка на View.
func (c *handlerUI) showSelectType(g *gocui.Gui, _ *gocui.View) error {

	c.conf.LgrFile.Write("Info: Нажата комбинация Ctrl+U")

	// Ограничение.
	if layerShowSelectTypeIsRestraintRun(c) {
		return nil
	}

	// Закрытие подключения к БД, чтобы была возможность BackUp и Restore.
	if err := layerShowSelectTypeCloseDB(c); err != nil {
		return fmt.Errorf("Функция layerShowSelectTypeCloseDB, вернула ошибку:<%w>", err)
	}

	// Сброс статусных признаков.
	if err := layerShowSelectTypeReset(c); err != nil {
		return fmt.Errorf("Функция layerShowSelectTypeReset, вернула ошибку:<%w>", err)
	}

	// Создание экземпляра контейнера.
	// Создаётся экземпляр в этом месте, т.к. ключ шифрования формируется после запроса дополнительного ключа.
	if err := layerShowSelectTypeNewInstContainer(c); err != nil {
		return fmt.Errorf("Функция layerShowSelectTypeNewInstContainer, вернула ошибку:<%w>", err)
	}

	// Очистка видов.
	if err := layerShowSelectTypeClear(c, g); err != nil {
		return fmt.Errorf("Функция layerShowSelectTypeClear, вернула ошибку:<%w>", err)
	}

	// Для дополнительного секретного ключа.
	window, err := layerShowSelectDrawWindow(c, g)
	if err != nil {
		return fmt.Errorf("Функция layerShowSelectDrawWindow, вернула ошибку:<%w>", err)
	}

	// Отрисовка типов данных.
	if err := layerShowSelectDrawTypes(c, g, window); err != nil {
		return fmt.Errorf("Функция layerShowSelectDrawTypes, вернула ошибку:<%w>", err)
	}

	// Отрисовка индикаторов.
	if err := layerShowSelectDrawIndicator(c, g); err != nil {
		return fmt.Errorf("Функция layerShowSelectDrawIndicator, вернула ошибку:<%w>", err)
	}

	// Отрисовка пояснений.
	if err := layerShowSelectDrawGuide(c, g, window); err != nil {
		return fmt.Errorf("Функция layerShowSelectDrawGuide, вернула ошибку:<%w>", err)
	}

	// Формирование уникального ID клиента.
	if err := layerShowSelectCreateClientID(c); err != nil {
		return fmt.Errorf("Функция , вернула ошибку:<%w>", err)
	}

	// Установка фокуса.
	if err := layerShowSelectSetFocus(c, g, "selectLoginPassword"); err != nil {
		return fmt.Errorf("Функция layerShowSelectSetFocus, вернула ошибку:<%w>", err)
	}

	return nil
}

// Окно для взаимодействия с логин/пароль. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	_ - заглушка на View.
func (c *handlerUI) showLoginPassword(g *gocui.Gui, _ *gocui.View) error {

	c.conf.LgrFile.Write("Info: Запущен обработчик логин/пароль")

	// Ограничение.
	if layerShowLoginPasswordRestraintRun(c) {
		return nil
	}

	// Подключение к БД.
	if err := layerShowLoginPasswordNewInstDB(c); err != nil {
		return fmt.Errorf("Функция layerShowLoginPasswordNewInstDB, вернула ошибку: <%w>", err)
	}

	// Сброс переменных.
	if err := layerShowLoginPasswordReset(c); err != nil {
		return fmt.Errorf("Функция layerShowLoginPasswordReset, вернула ошибку: <%w>", err)
	}

	// Очистка.
	if err := layerShowLoginPasswordClear(c, g); err != nil {
		return fmt.Errorf("Функция layerShowLoginPasswordClear, вернула ошибку : <%w>", err)
	}

	// Создание окна.
	window, err := layerShowLoginPasswordWindow(c, g)
	if err != nil {
		return fmt.Errorf("Функция layerShowLoginPasswordWindow, вернула ошибку : <%w>", err)
	}

	// Отрисовка полей ввода.
	if err := layerShowLoginPasswordDrawInput(c, g, window); err != nil {
		return fmt.Errorf("Функция layerShowLoginPasswordDrawInput, вернула ошибку : <%w>", err)
	}

	// Отрисовка индикаторов.
	if err := layerShowLoginPasswordDrawIndicator(c, g); err != nil {
		return fmt.Errorf("функция layerShowLoginPasswordDrawIndicator, вернула ошибку: <%w>", err)
	}

	// Отрисовка пояснений.
	if err := layerShowLoginPasswordDrawGuide(c, g, window); err != nil {
		return fmt.Errorf("функция layerShowLoginPasswordDrawGuide, вернула ошибку: <%w>", err)
	}

	// Получение сохранённых значений логин/пароль.
	if err := layerShowLoginPasswordActions(c); err != nil {
		return fmt.Errorf("функция layerShowLoginPasswordActions, вернула ошибку: <%w>", err)
	}

	// Установка фокуса.
	if err := layerShowLoginPasswordSetFocus(c, g, "fieldAddFor"); err != nil {
		return fmt.Errorf("Функция layerShowLoginPasswordSetFocus, вернула ошибку:<%w>", err)
	}

	return nil
}

// Окно для взаимодействия с тектом. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	_ - заглушка на View.
func (c *handlerUI) showText(g *gocui.Gui, _ *gocui.View) error {

	c.conf.LgrFile.Write("Info: Запущен обработчик текста")

	// ограничение.
	if layerShowTextRestraintRun(c) {
		return nil
	}

	// Подключение к БД.
	if err := layerShowTextNewInstDB(c); err != nil {
		return fmt.Errorf("Функция layerShowTextNewInstDB, вернула ошибку:<%w>", err)
	}

	// Сброс переменных.
	if err := layerShowTextReset(c); err != nil {
		return fmt.Errorf("Функция layerShowTextReset, вернула ошибку:<%w>", err)
	}

	// Удаляем все зависимые виды
	if err := layerShowTextClear(c, g); err != nil {
		return fmt.Errorf("функция layerShowTextClear, вернула ошибку: <%w>", err)
	}

	// Создание окна.
	window, err := layerShowTextdWindow(c, g)
	if err != nil {
		return fmt.Errorf("функция layerShowTextdWindow, вернула ошибку: <%w>", err)
	}

	// Отрисовка индикаторов.
	if err := layerShowTextDrawIndicator(c, g); err != nil {
		return fmt.Errorf("функция layerShowTextDrawIndicator, вернула ошибку: <%w>", err)
	}

	// Отрисовка полей ввода.
	if err := layerShowTextDrawInput(c, g); err != nil {
		return fmt.Errorf("функция layerShowTextDrawInput, вернула ошибку: <%w>", err)
	}

	// Отрисовка пояснений.
	if err := layerShowTextDrawGuide(c, g, window); err != nil {
		return fmt.Errorf("функция layerShowTextDrawGuide, вернула ошибку: <%w>", err)
	}

	// Получение сохранённых значений текста.
	if err := layerShowTextActions(c); err != nil {
		return fmt.Errorf("функция layerShowTextActions, вернула ошибку: <%w>", err)
	}

	// Установка фокуса.
	if err := layerShowTextSetFocus(c, g, "fieldAddFor"); err != nil {
		return fmt.Errorf("функция layerShowTextSetFocus, вернула ошибку: <%w>", err)
	}

	return nil
}

// Окно для взаимодействия с файлами. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	_ - заглушка на View.
func (c *handlerUI) showBinary(g *gocui.Gui, _ *gocui.View) error {

	c.conf.LgrFile.Write("Info: Запущен обработчик файлов")

	// Ограничение.
	if layerShowBinaryRestraintRun(c) {
		return nil
	}

	// Подключение к БД.
	if err := layerShowBinaryNewInstDB(c); err != nil {
		return fmt.Errorf("Функция layerShowBinaryNewInstDB, вернула ошибку:<%w>", err)
	}

	// Сброс.
	if err := layerShowBinaryReset(c); err != nil {
		return fmt.Errorf("Функция layerShowBinaryReset, вернула ошибку:<%w>", err)
	}

	// Очистка.
	if err := layerShowBinaryClear(c, g); err != nil {
		return fmt.Errorf("функция layerShowBinaryClear, вернула ошибку: <%w>", err)
	}

	// Создание окна.
	window, err := layerShowBinaryWindow(c, g)
	if err != nil {
		return fmt.Errorf("функция layerShowBinaryWindow, вернула ошибку: <%w>", err)
	}

	// Отрисовка пояснений
	if err := layerShowBinaryDrawGuide(c, g, window); err != nil {
		return fmt.Errorf("функция layerShowBinaryDrawGuide, вернула ошибку: <%w>", err)
	}

	// Оотрисовка индикаторов.
	if err := layerShowBinaryIndicator(c, g); err != nil {
		return fmt.Errorf("функция layerShowBinaryIndicator, вернула ошибку: <%w>", err)
	}

	// Отрирсовка полей вывода.
	if err := layerShowBinaryDrawOut(c, g); err != nil {
		return fmt.Errorf("функция layerShowBinaryDrawOut, вернула ошибку: <%w>", err)
	}

	// Отрисовка полей ввода.
	if err := layerShowBinaryDrawInput(c, g); err != nil {
		return fmt.Errorf("функция layerShowBinaryDrawInput, вернула ошибку: <%w>", err)
	}

	// Действия.
	if err := layerShowBinaryActions(c); err != nil {
		return fmt.Errorf("функция layerShowBinaryActions, вернула ошибку: <%w>", err)
	}

	// Установка фокуса.
	if err := layerShowBinarySetFocus(c, g, "fieldPathSource"); err != nil {
		return fmt.Errorf("функция layerShowBinarySetFocus, вернула ошибку: <%w>", err)
	}

	return nil
}

// Окно для взаимодействия с банковской картой. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	_ - заглушка на View.
func (c *handlerUI) showBankCard(g *gocui.Gui, _ *gocui.View) error {

	c.conf.LgrFile.Write("Info: Запущен обработчик банковских карт")

	// Ограничение.
	if layerShowBankCardRestraintRun(c) {
		return nil
	}

	// Подключение к БД.
	if err := layerShowBankCardNewInstDB(c); err != nil {
		return fmt.Errorf("Функция layerShowBankCardNewInstDB, вернула ошибку:<%w>", err)
	}

	// Сброс переменных.
	if err := layerShowBankCardReset(c); err != nil {
		return fmt.Errorf("Функция layerShowBankCardReset, вернула ошибку:<%w>", err)
	}

	// Очистка экрана.
	if err := layerShowBankCardClear(c, g); err != nil {
		return fmt.Errorf("Функция layerShowBankCardClear, вернула ошибку:<%w>", err)
	}

	// Создание окна.
	window, err := layerShowBankCardWindow(c, g)
	if err != nil {
		return fmt.Errorf("Функция layerShowBankCarddWindow, вернула ошибку:<%w>", err)
	}

	// Вывод пояснений.
	if err := layerShowBankCardDrawGuide(c, g, window); err != nil {
		return fmt.Errorf("функция layerShowBankCardDrawGuide, вернула ошибку: <%w>", err)
	}

	// Отрисовка полей ввода.
	if err := layerShowBankCardDrawInput(c, g); err != nil {
		return fmt.Errorf("функция layerShowBankCardDrawInput, вернула ошибку: <%w>", err)
	}

	// Отрисовка полей вывода.
	if err := layerShowBankCardDrawOut(c, g); err != nil {
		return fmt.Errorf("функция layerShowBankCardDrawOut, вернула ошибку: <%w>", err)
	}

	// Отрисовка индикаторов.
	if err := layerShowBankCardIndicator(c, g); err != nil {
		return fmt.Errorf("функция layerShowBankCardIndicator, вернула ошибку: <%w>", err)
	}

	// Действия.
	if err := layerShowBankCardActions(c); err != nil {
		return fmt.Errorf("функция layerShowBankCardActions, вернула ошибку: <%w>", err)
	}

	// Установка фокуса.
	if err := layerShowBankCardSetFocus(c, g, "fieldAddFor"); err != nil {
		return fmt.Errorf("функция layerShowBankCardSetFocus, вернула ошибку: <%w>", err)
	}

	return nil
}

// Передача данных клиента, на сервер. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	v - указатель на View.
func (c *handlerUI) doBackup(g *gocui.Gui, v *gocui.View) (err error) {

	c.conf.LgrFile.Write("Info: Нажата комбинация Ctrl+O")

	// Сброс сторожевого таймера.
	if c.status.statusWDT {
		c.ch.resetWDT <- struct{}{} // Сброс таймера.
	}

	// Запрет отработки.
	if layerDoBackUpRestraintRun(c) {
		return nil
	}

	// Закрытие подключения к БД
	if err := layerDoBackUpCloseDB(c); err != nil {
		return fmt.Errorf("функция layerDoRestoreCloseDB, вернула ошибку: <%w>", err)
	}

	// Логика обработчика.
	if err := layerDoBackUpActions(c); err != nil {
		return fmt.Errorf("функция layerDoBackUpActions, вернула ошибку: <%w>", err)
	}

	return nil
}

// Получение данных клиента, от сервер. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на Gui.
//	v - указатель на View.
func (c *handlerUI) doRestore(g *gocui.Gui, v *gocui.View) error {

	c.conf.LgrFile.Write("Info: Нажата комбинация Ctrl+P")

	// Сброс сторожевого таймера.
	if c.status.statusWDT {
		c.ch.resetWDT <- struct{}{} // Сброс таймера.
	}

	// Запрет отработки.
	if layerDoRestoreRestraintRun(c) {
		return nil
	}

	// Закрытие подключения к БД
	if err := layerDoRestoreCloseDB(c); err != nil {
		return fmt.Errorf("функция layerDoRestoreCloseDB, вернула ошибку: <%w>", err)
	}

	// Логика обработчика.
	if err := layerDoRestoreActions(c); err != nil {
		return fmt.Errorf("Функция layerDoRestoreActions, вернула ошибку:<%w>", err)
	}

	return nil
}

// Обновление статуса процесса передачи.
//
// Параметры:
//
//	st - новое значение статуса.
func (c *handlerUI) updateStatusBackUp(st int) {

	c.mutex.statusBackUp.Lock()
	defer c.mutex.statusBackUp.Unlock()

	c.status.backUp = st
}

// Получение текущего статуса процесса передачи. Возвращается значение статуса.
func (c *handlerUI) getStatusBackUp() int {

	c.mutex.statusBackUp.Lock()
	defer c.mutex.statusBackUp.Unlock()

	return c.status.backUp
}

// Обновление статуса процесса передачи.
//
// Параметры:
//
//	st - новое значение статуса.
func (c *handlerUI) updateStatusRestore(st int) {

	c.mutex.statusRestore.Lock()
	defer c.mutex.statusRestore.Unlock()

	c.status.restore = st
}

// Получение текущего статуса процесса передачи. Возвращается значение статуса.
func (c *handlerUI) getStatusRestore() int {

	c.mutex.statusRestore.Lock()
	defer c.mutex.statusRestore.Unlock()

	return c.status.restore
}

// Получение процента выполения TxRx. Возвращается значение процентов процесса.
func (c *handlerUI) getPercentTxRx() float32 {

	c.mutex.processTxRx.Lock()
	defer c.mutex.processTxRx.Unlock()

	return c.txrx.percentTxRx
}

// Установка процента выполения TxRx.
//
// Параметры:
//
// percent - текущие проценты процесса.
func (c *handlerUI) setPercentTxRx(percent float32) {

	c.mutex.processTxRx.Lock()
	defer c.mutex.processTxRx.Unlock()

	c.txrx.percentTxRx = percent
}

// Обновление статуса процесса передачи.
//
// Параметры:
//
//	st - новое значение статуса.
func (c *handlerUI) updateStatusPushContainer(st int) {

	c.mutex.statusPushContainer.Lock()
	defer c.mutex.statusPushContainer.Unlock()

	c.status.pushContainer = st
}

// Получение текущего статуса процесса передачи. Возвращается текущее значение статуса.
func (c *handlerUI) getStatusPushContainer() int {

	c.mutex.statusPushContainer.Lock()
	defer c.mutex.statusPushContainer.Unlock()

	return c.status.pushContainer
}

// Обновление статуса процесса передачи.
//
// Параметры:
//
//	st - новое значение статуса.
func (c *handlerUI) updateStatusPopContainer(st int) {

	c.mutex.statusPopContainer.Lock()
	defer c.mutex.statusPopContainer.Unlock()

	c.status.popContainer = st
}

// Получение текущего статуса процесса передачи файла на сервер, в режиме - удалённый. Возвращается текущее значение статуса.
func (c *handlerUI) getStatusFileTx() int {

	c.mutex.statusFileTx.Lock()
	defer c.mutex.statusFileTx.Unlock()

	return c.status.fileTx
}

// Обновление статуса процесса передачи файла на сервер, в режиме - удалённый.
//
// Параметры:
//
//	st - новое значение статуса.
func (c *handlerUI) updateStatusFileTx(st int) {

	c.mutex.statusFileTx.Lock()
	defer c.mutex.statusFileTx.Unlock()

	c.status.fileTx = st
}

// Получение текущего статуса процесса приёма файла от сервера, в режиме - удалённый. Возвращается текущее значение статуса.
func (c *handlerUI) getStatusFileRx() int {

	c.mutex.statusFileRx.Lock()
	defer c.mutex.statusFileRx.Unlock()

	return c.status.fileRx
}

// Обновление статуса процесса приёма файла от сервера, в режиме - удалённый.
//
// Параметры:
//
//	st - новое значение статуса.
func (c *handlerUI) updateStatusFileRx(st int) {

	c.mutex.statusFileRx.Lock()
	defer c.mutex.statusFileRx.Unlock()

	c.status.fileRx = st
}

// Получение текущего статуса процесса передачи. Возвращается текущее значение статуса.
func (c *handlerUI) getStatusPopContainer() int {

	c.mutex.statusPopContainer.Lock()
	defer c.mutex.statusPopContainer.Unlock()

	return c.status.popContainer
}

// Обновление статуса занятости сервера, в режиме - удалённый.
//
// Параметры:
//
//	st - новое значение статуса.
func (c *handlerUI) updateStatusIsBusyServer(st bool) {

	c.mutex.statusIsBusyServer.Lock()
	defer c.mutex.statusIsBusyServer.Unlock()

	c.status.isBusyServer = st
}

// Получение текущего занятости сервера, в режиме - удалённый. Возвращается текущее значение статуса.
func (c *handlerUI) getStatusIsBusyServer() bool {

	c.mutex.statusIsBusyServer.Lock()
	defer c.mutex.statusIsBusyServer.Unlock()

	return c.status.isBusyServer
}

// инициализация каналов сторожевого таймера.
func (c *handlerUI) initChannelsWDT() {

	// Канал для сброса сторожевого таймера.
	if c.ch.resetWDT == nil {
		c.ch.resetWDT = make(chan struct{})
		return
	}
	// Канал для завершения работы.
	if c.ch.resetWDTClose == nil {
		c.ch.resetWDTClose = make(chan struct{})
		return
	}
}
