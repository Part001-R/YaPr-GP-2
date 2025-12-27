package ui

import (
	"fmt"
	"time"

	"github.com/Part001-R/YaPr-GP-2/client/internal/service/udt"
	"github.com/jroimartin/gocui"
	"go.uber.org/zap"
)

// Run, главная функция ui cli. Возвращает ошибку.
func Run(conf *udt.Configuration) error {

	// Проверка аргументов
	if conf == nil {
		return NilPtrArgumentConf
	}
	if err := conf.CheckConf(); err != nil {
		return fmt.Errorf("функция CheckConf, вернула ошибку: <%w>", err)
	}

	// GUI
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		conf.PtrLogger.Error("Ошибка создания GUI", zap.Error(err))
		return fmt.Errorf("Ошибка создания GUI: <%w>", err)
	}
	defer g.Close()

	// Создание экземпляра обработчиков, для передачи параметров сервиса.
	instUI := new(conf)

	g.SetManagerFunc(layout)

	// Привязки клавиш
	//
	// Завершение работы.
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, instUI.quit); err != nil {
		conf.PtrLogger.Error("Ошибка привязки Ctrl+C", zap.Error(err))
		return fmt.Errorf("Ошибка привязки Ctrl+C: <%w>", err)
	}
	// Переход на экран регистрации.
	if err := g.SetKeybinding("", gocui.KeyCtrlA, gocui.ModNone, instUI.showRegistration); err != nil {
		conf.PtrLogger.Error("Ошибка привязки Ctrl+A", zap.Error(err))
		return fmt.Errorf("Ошибка привязки Ctrl+A: <%w>", err)
	}
	// Переход на экран аутентификации.
	if err := g.SetKeybinding("", gocui.KeyCtrlB, gocui.ModNone, instUI.showAuthentication); err != nil {
		conf.PtrLogger.Error("Ошибка привязки Ctrl+B", zap.Error(err))
		return fmt.Errorf("Ошибка привязки Ctrl+B: <%w>", err)
	}
	// Переход на экран настроек.
	if err := g.SetKeybinding("", gocui.KeyCtrlD, gocui.ModNone, instUI.showSettings); err != nil {
		conf.PtrLogger.Error("Ошибка привязки Ctrl+D", zap.Error(err))
		return fmt.Errorf("Ошибка привязки Ctrl+D: <%w>", err)
	}
	// Переход на главный экран.
	if err := g.SetKeybinding("", gocui.KeyCtrlH, gocui.ModNone, instUI.showMain); err != nil {
		conf.PtrLogger.Error("Ошибка привязки Ctrl+H", zap.Error(err))
		return fmt.Errorf("Ошибка привязки Ctrl+H: <%w>", err)
	}
	// Проверки связи с сервером.
	if err := g.SetKeybinding("", gocui.KeyCtrlN, gocui.ModNone, instUI.testConnect); err != nil {
		conf.PtrLogger.Error("Ошибка привязки Ctrl+N", zap.Error(err))
		return fmt.Errorf("Ошибка привязки Ctrl+N: <%w>", err)
	}
	// Переключение фокуса по Tab.
	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, instUI.nextFocus); err != nil {
		conf.PtrLogger.Error("Ошибка привязки Tab", zap.Error(err))
		return fmt.Errorf("Ошибка привязки Tab: <%w>", err)
	}
	// Запуск процедуры регистрации.
	if err := g.SetKeybinding("", gocui.KeyCtrlW, gocui.ModNone, instUI.doRegistrationUser); err != nil {
		conf.PtrLogger.Error("Ошибка привязки Ctrl+W", zap.Error(err))
		return fmt.Errorf("Ошибка привязки Ctrl+W: <%w>", err)
	}
	// Запуск процедуры аутентификации.
	if err := g.SetKeybinding("", gocui.KeyCtrlL, gocui.ModNone, instUI.doAuthenticationUser); err != nil {
		conf.PtrLogger.Error("Ошибка привязки Ctrl+L", zap.Error(err))
		return fmt.Errorf("Ошибка привязки Ctrl+L: <%w>", err)
	}
	// Возврат на предыдущее окно.
	if err := g.SetKeybinding("", gocui.KeyCtrlU, gocui.ModNone, instUI.showSelectType); err != nil {
		conf.PtrLogger.Error("Ошибка привязки Ctrl+U", zap.Error(err))
		return fmt.Errorf("Ошибка привязки Ctrl+U: <%w>", err)
	}
	// Сохранение логин/пароль в БД.
	if err := g.SetKeybinding("", gocui.KeyCtrlF, gocui.ModNone, instUI.doStoreDB); err != nil {
		conf.PtrLogger.Error("Ошибка привязки Ctrl+F", zap.Error(err))
		return fmt.Errorf("Ошибка привязки Ctrl+F: <%w>", err)
	}
	// Отображение следующего элемента.
	if err := g.SetKeybinding("", gocui.KeyCtrlE, gocui.ModNone, instUI.doShowNextElement); err != nil {
		conf.PtrLogger.Error("Ошибка привязки Ctrl+E", zap.Error(err))
		return fmt.Errorf("Ошибка привязки Ctrl+E: <%w>", err)
	}
	// Отображение предыдцщего элемента.
	if err := g.SetKeybinding("", gocui.KeyCtrlG, gocui.ModNone, instUI.doShowPrevElement); err != nil {
		conf.PtrLogger.Error("Ошибка привязки Ctrl+G", zap.Error(err))
		return fmt.Errorf("Ошибка привязки Ctrl+G: <%w>", err)
	}
	// Удаление.
	if err := g.SetKeybinding("", gocui.KeyCtrlJ, gocui.ModNone, instUI.doDeleteElement); err != nil {
		conf.PtrLogger.Error("Ошибка привязки Ctrl+J", zap.Error(err))
		return fmt.Errorf("Ошибка привязки Ctrl+J: <%w>", err)
	}
	// Enter для полей.
	listElement := []string{
		"Login",
		"Password-1",
		"Password-2",
		"IP",
		"Port",
		"scrtKey",
		"selectLoginPassword",
		"selectText",
		"SelectBinary",
		"SelectBankCard",
		"fieldAddFor",
		"fieldAddLogin",
		"fieldAddPassword",
		"indicatorAddSuccess",
		"fieldAddText",
		"fieldAddOwner",
		"fieldAddNumber",
		"fieldAddValid",
		"fieldAddCode",
	}

	for _, name := range listElement {
		if err := g.SetKeybinding(name, gocui.KeyEnter, gocui.ModNone, instUI.handleEnter); err != nil {
			conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция SetKeybinding, вернула ошибку: <%v> при Enter на элементе: <%s>", err, name))
			return fmt.Errorf("Ошибка привязки KeyEnter: <%w>", err)
		}
	}

	// обработка индикаторов на формах.
	go func() {
		for {
			g.Update(func(g *gocui.Gui) error {
				if !layoutInitialized {
					return nil
				}
				instUI.indicators(g)
				return nil
			})
			time.Sleep(100 * time.Millisecond)
		}
	}()

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		conf.PtrLogger.Error("Ошибка MainLoop", zap.Error(err))
		return fmt.Errorf("Ошибка MainLoop: <%w>", err)
	}

	// Завершение работы.
	conf.PtrLogger.Debug("CLI UI завершил работу")
	return nil
}
