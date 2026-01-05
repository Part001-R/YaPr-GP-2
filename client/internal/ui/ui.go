package ui

import (
	"fmt"
	"time"

	"github.com/Part001-R/YaPr-GP-2/client/internal/service/udt"
	"github.com/jroimartin/gocui"
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
		return fmt.Errorf("Ошибка создания GUI: <%w>", err)
	}
	defer g.Close()

	// Создание экземпляра обработчиков, для передачи параметров сервиса.
	instUI := new(conf)

	g.SetManagerFunc(layout)

	// Привязки.
	//
	// Завершение работы.
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, instUI.quit); err != nil {
		return fmt.Errorf("Error: Ошибка Ctrl+C: <%w>", err)
	}
	// Переход на экран регистрации.
	if err := g.SetKeybinding("", gocui.KeyCtrlA, gocui.ModNone, instUI.showRegistration); err != nil {
		return fmt.Errorf("Error: Ошибка Ctrl+A: <%w>", err)
	}
	// Переход на экран аутентификации.
	if err := g.SetKeybinding("", gocui.KeyCtrlB, gocui.ModNone, instUI.showAuthentication); err != nil {
		return fmt.Errorf("Error: Ошибка Ctrl+B: <%w>", err)
	}
	// Переход на экран настроек.
	if err := g.SetKeybinding("", gocui.KeyCtrlD, gocui.ModNone, instUI.showSettings); err != nil {
		return fmt.Errorf("Error: Ошибка Ctrl+D: <%w>", err)
	}
	// Переход на главный экран.
	if err := g.SetKeybinding("", gocui.KeyCtrlH, gocui.ModNone, instUI.showMain); err != nil {
		return fmt.Errorf("Error: Ошибка Ctrl+H: <%w>", err)
	}
	// Проверки связи с сервером.
	if err := g.SetKeybinding("", gocui.KeyCtrlN, gocui.ModNone, instUI.testConnect); err != nil {
		return fmt.Errorf("Error: Ошибка Ctrl+N: <%w>", err)
	}
	// Переключение фокуса по Tab.
	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, instUI.nextFocus); err != nil {
		return fmt.Errorf("Error: Ошибка Tab: <%w>", err)
	}
	// Запуск процедуры регистрации.
	if err := g.SetKeybinding("", gocui.KeyCtrlW, gocui.ModNone, instUI.doRegistrationUser); err != nil {
		return fmt.Errorf("Error: Ошибка Ctrl+W: <%w>", err)
	}
	// Запуск процедуры аутентификации.
	if err := g.SetKeybinding("", gocui.KeyCtrlL, gocui.ModNone, instUI.doAuthenticationUser); err != nil {
		return fmt.Errorf("Error: Ошибка Ctrl+L: <%w>", err)
	}
	// Возврат на предыдущее окно.
	if err := g.SetKeybinding("", gocui.KeyCtrlU, gocui.ModNone, instUI.showSelectType); err != nil {
		return fmt.Errorf("Error: Ошибка Ctrl+U: <%w>", err)
	}
	// Сохранение.
	if err := g.SetKeybinding("", gocui.KeyCtrlF, gocui.ModNone, instUI.doStore); err != nil {
		return fmt.Errorf("Error: Ошибка Ctrl+F: <%w>", err)
	}
	// Отображение следующего элемента.
	if err := g.SetKeybinding("", gocui.KeyCtrlE, gocui.ModNone, instUI.doShowNextElement); err != nil {
		return fmt.Errorf("Error: Ошибка Ctrl+E: <%w>", err)
	}
	// Отображение предыдцщего элемента.
	if err := g.SetKeybinding("", gocui.KeyCtrlG, gocui.ModNone, instUI.doShowPrevElement); err != nil {
		return fmt.Errorf("Error: Ошибка Ctrl+G: <%w>", err)
	}
	// Удаление.
	if err := g.SetKeybinding("", gocui.KeyCtrlJ, gocui.ModNone, instUI.doDeleteElement); err != nil {
		return fmt.Errorf("Error: Ошибка Ctrl+J: <%w>", err)
	}
	// Извлечение.
	if err := g.SetKeybinding("", gocui.KeyCtrlK, gocui.ModNone, instUI.doExtract); err != nil {
		return fmt.Errorf("Error: Ошибка Ctrl+K: <%w>", err)
	}
	// Backup (---> сервер).
	if err := g.SetKeybinding("", gocui.KeyCtrlO, gocui.ModNone, instUI.doBackup); err != nil {
		return fmt.Errorf("Error: Ошибка Ctrl+O: <%w>", err)
	}
	// Restore (<--- сервер).
	if err := g.SetKeybinding("", gocui.KeyCtrlP, gocui.ModNone, instUI.doRestore); err != nil {
		return fmt.Errorf("Error: Ошибка Ctrl+P: <%w>", err)
	}
	// Enter для полей ввода и выбора.
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
		"fieldPathSource",
		"fieldPathTarget",
	}

	for _, name := range listElement {
		if err := g.SetKeybinding(name, gocui.KeyEnter, gocui.ModNone, instUI.handleEnter); err != nil {
			conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка обработки нажатия Enter: <%v>, на элементе: <%s>", err, name))
			return fmt.Errorf("Error: ошибка обработки нажатия Enter: <%w>, на элементе: <%s>", err, name)
		}
	}

	// Обновление внешнего вида индикаторов.
	go func() {
		for {
			g.Update(func(g *gocui.Gui) error {
				if !layoutInitialized {
					return nil
				}
				if err := instUI.indicators(g); err != nil {
					return fmt.Errorf("Error: ошибка обновления индикаторов: <%w>", err)
				}
				return nil
			})
			time.Sleep(100 * time.Millisecond)
		}
	}()

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка MainLoop: <%v>", err))
		return fmt.Errorf("Error: Ошибка MainLoop: <%v>", err)
	}

	// Завершение работы.
	conf.PtrLoggerFile.Write("Info: CLI UI завершил работу")
	return nil
}
