package ui

import (
	"fmt"
	"log"
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
		return err
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
	// Enter для всех полей ввода.
	for _, name := range []string{"Login", "Password-1", "Password-2", "IP", "Port", "scrtKey"} {
		if err := g.SetKeybinding(name, gocui.KeyEnter, gocui.ModNone, instUI.handleEnter); err != nil {
			log.Panicln(err)
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
