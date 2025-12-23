package ui

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Part001-R/YaPr-GP-2/client/internal/service/udt"
	"github.com/jroimartin/gocui"
)

var once sync.Once

type encrKey struct {
	secretKey string // секретный ключ
}

// Введённые пользователем данные.
type typeData struct {
	login     string // введённое значение в поле имя пользователя
	password1 string // введённое значение в воле пароль
	password2 string // введённое значение в воле пароль (подтверждение)
	ip        string // введённое значение в поле IP
	port      string // введённое значение в поле Port
}

// Признаки выполнения логики
type flags struct {
	checkConnectStatus bool // Результат процедуры проверки связи с сервером.
	checkConnectPassed bool // Признак, что проверка связи была запущена.
	addUserSUCCESS     bool // Признак успешного добавления пользователя.
	addUserPassed      bool // Признак, что была запущена процедура регистрации пользователя.
}

// Для навигации по экранам.
type screens struct {
	activeView   string // название активного экрана
	currentFocus string // на какой элемент установлен фокус
}

// Общий тип для CLI UI.
type handlerUI struct {
	conf   *udt.Configuration
	typed  typeData
	flag   flags
	view   screens
	secret encrKey
}

var inst *handlerUI

// Конструктор.
func new(conf *udt.Configuration) *handlerUI {
	once.Do(func() {
		inst = &handlerUI{
			conf:  conf,
			typed: typeData{},
			flag:  flags{},
			view: screens{
				activeView:   "",
				currentFocus: "Login",
			},
		}
	})
	return inst
}

// Главное окно.
func layout(g *gocui.Gui) error {

	// Главное меню
	mainView, err := g.SetView(viewMain, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	mainView.Wrap = true
	mainView.Clear()

	fmt.Fprintln(mainView, strings.Repeat("\n", 3))
	fmt.Fprintf(mainView, "%s МЕНЕДЖЕР ПАРОЛЕЙ\n", strings.Repeat(" ", 54))

	fmt.Fprintln(mainView, strings.Repeat("\n", 15))
	fmt.Fprintf(mainView, "%s Регистрация     (Ctrl+A)\n", strings.Repeat(" ", 50))
	fmt.Fprintf(mainView, "%s Аутентификация  (Ctrl+B)\n", strings.Repeat(" ", 50))
	fmt.Fprintf(mainView, "%s Настройки       (Ctrl+D)\n", strings.Repeat(" ", 50))
	fmt.Fprintln(mainView, "")
	fmt.Fprintf(mainView, "%s Выход           (Ctrl+C)\n", strings.Repeat(" ", 50))

	return nil
}

// Главное окно.
func (c *handlerUI) showMain(g *gocui.Gui, v *gocui.View) error {

	c.view.activeView = "" // Сброс признака активного окна

	// Удаление видов
	if err := deleteViews(g); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция deleteViews вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews вернула ошибку: <%w>", err)
	}

	// Сброс флагов
	layoutInitialized = false
	c.view.currentFocus = ""

	// Пересоздание главного меню
	err := layout(g)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция layout вернула ошибку: <%v>", err))
		return fmt.Errorf("функция layout вернула ошибку: <%w>", err)
	}
	c.view.activeView = viewMain // Установка признака активного окна
	return nil
}

// Регистрация.
func (c *handlerUI) showRegistration(g *gocui.Gui, _ *gocui.View) error {

	// Сбросы
	c.view.activeView = ""
	c.typed.login = ""
	c.typed.password1 = ""
	c.typed.password2 = ""
	c.flag.addUserPassed = false
	c.flag.addUserSUCCESS = false

	// Удаляем все зависимые виды (включая поля ввода)
	if err := deleteViews(g); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}

	// Сброс состояния
	layoutInitialized = false
	c.view.currentFocus = "Login" // Установка фокуса

	// Создание контейнера регистрации
	loginView, err := g.SetView(viewRegistration, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
	}
	loginView.Title = "Регистрация"
	loginView.Wrap = true
	loginView.Clear()

	//
	// --- Поля ввода ---
	//

	if v, err := g.SetView("Login", 50, 2, inputWidth+1, inputHeight+1+1); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Title = "Логин"
		v.Editable = true
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorBlack
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack

		v.Write([]byte(".........."))
		c.setFocusStyle(v, "Login")
	}

	if v, err := g.SetView("Password-1", 50, 7, inputWidth+1, inputHeight+1+6); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Title = "Пароль"
		v.Editable = true
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorBlack
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack

		v.Write([]byte(".........."))
		c.setFocusStyle(v, "Password-1")
	}

	if v, err := g.SetView("Password-2", 50, 10, inputWidth+1, inputHeight+1+9); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Title = "Подтверждение"
		v.Editable = true
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorBlack
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack

		v.Write([]byte(".........."))
		c.setFocusStyle(v, "Password-2")
	}

	//
	// --- Индикаторы ---
	//

	// Признак корректных данных пользователя
	indicatorY := inputHeight*3 + 7
	indicatorX := 58
	v, err := g.SetView("indicator-match", indicatorX, indicatorY, indicatorX+20, indicatorY+2)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
	}
	v.Frame = false
	v.BgColor = gocui.ColorDefault
	v.FgColor = gocui.ColorGreen

	// Признак успешной регистрации.
	indicatorY = inputHeight*3 + 10
	indicatorX = 43
	v, err = g.SetView("indicator-registration", indicatorX, indicatorY, indicatorX+80, indicatorY+2)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
	}
	v.Frame = false
	v.BgColor = gocui.ColorDefault
	v.FgColor = gocui.ColorGreen

	//
	// --- Пояснение по навигации ---
	//

	//
	if v, err := g.SetView("TAB", 1, 26, inputWidth-55, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Tab - перевод фокуса"))
	}
	if v, err := g.SetView("Enter", 26, 26, inputWidth-29, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Enter - фиксация ввода"))
	}
	if v, err := g.SetView("MainMenu", 52, 26, inputWidth-3, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+H - главное меню"))
	}
	if v, err := g.SetView("Exit", 78, 26, inputWidth+23, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+C - выход"))
	}
	if v, err := g.SetView("DoRegistration", 104, 26, inputWidth+48, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
			return fmt.Errorf("Не удалось установить фокус: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+W - регистрация"))
	}

	// Установка фокуса.
	if _, err := g.SetCurrentView(c.view.currentFocus); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	layoutInitialized = true

	c.view.activeView = viewRegistration // Установка признака активного окна

	return nil
}

// Аутентификация.
func (c *handlerUI) showAuthentication(g *gocui.Gui, _ *gocui.View) error {

	// Сбросы
	c.view.activeView = ""
	c.typed.login = ""
	c.typed.password1 = ""
	c.typed.password2 = ""

	// Удаляем все зависимые виды (включая поля ввода)
	if err := deleteViews(g); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}

	// Сброс состояния
	layoutInitialized = false
	c.view.currentFocus = "Login" // Установка фокуса

	// Создание контейнера регистрации
	loginView, err := g.SetView(viewAutentification, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
	}
	loginView.Title = "Аутентификация"
	loginView.Wrap = true
	loginView.Clear()

	//
	// --- Поля ввода ---
	//
	if v, err := g.SetView("Login", 50, 2, inputWidth+1, inputHeight+1+1); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Title = "Логин"
		v.Editable = true
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorBlack
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack

		v.Write([]byte(".........."))
		c.setFocusStyle(v, "Login")
	}

	if v, err := g.SetView("Password-1", 50, 7, inputWidth+1, inputHeight+1+6); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Title = "Пароль"
		v.Editable = true
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorBlack
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack

		v.Write([]byte(".........."))
		c.setFocusStyle(v, "Password-1")
	}

	//
	// --- Индикаторы ---
	//

	// Индикатор валидности
	indicatorY := inputHeight*3 + 7
	indicatorX := 58
	v, err := g.SetView("indicator-match", indicatorX, indicatorY, indicatorX+20, indicatorY+2)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
	}
	v.Frame = false
	v.BgColor = gocui.ColorDefault
	v.FgColor = gocui.ColorGreen

	//
	// --- Пояснение по навигации ---
	//

	if v, err := g.SetView("TAB", 1, 26, inputWidth-55, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Tab - перевод фокуса"))
	}
	if v, err := g.SetView("Enter", 26, 26, inputWidth-29, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Enter - фиксация ввода"))
	}
	if v, err := g.SetView("MainMenu", 52, 26, inputWidth-3, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+H - главное меню"))
	}
	if v, err := g.SetView("Exit", 78, 26, inputWidth+23, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+C - выход"))
	}

	if v, err := g.SetView("DoAuthentication", 104, 26, inputWidth+48, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
			return fmt.Errorf("Не удалось установить фокус: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+L - Подключение"))
	}

	// Установка фокуса.
	if _, err := g.SetCurrentView(c.view.currentFocus); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	layoutInitialized = true

	c.view.activeView = viewAutentification // Установка признака активного окна

	return nil
}

// Настройки.
func (c *handlerUI) showSettings(g *gocui.Gui, _ *gocui.View) error {

	c.view.activeView = "" // Сброс признака активного окна

	// Удаляем все зависимые виды (включая поля ввода)
	if err := deleteViews(g); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}

	// Сброс состояния
	layoutInitialized = false
	c.view.currentFocus = "IP" // Установка фокуса

	// Создание контейнера регистрации
	view, err := g.SetView(viewSettings, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	view.Title = "Настройки"
	view.Wrap = true
	view.Clear()

	// Поля ввода
	if v, err := g.SetView("IP", 50, 2, inputWidth+1, inputHeight+1+1); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
			return fmt.Errorf("Не удалось установить фокус: <%w>", err)
		}
		v.Title = "IP"
		v.Editable = true
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorBlack
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack

		if c.typed.ip == "" {
			v.Write([]byte(".........."))
			c.setFocusStyle(v, "IP")
		}
		if c.typed.ip != "" {
			v.Write([]byte(c.typed.ip))
			c.setFocusStyle(v, "IP")
		}
	}

	if v, err := g.SetView("Port", 50, 7, inputWidth+1, inputHeight+1+6); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
			return fmt.Errorf("Не удалось установить фокус: <%w>", err)
		}
		v.Title = "Порт"
		v.Editable = true
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorBlack
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack

		if c.typed.port == "" {
			v.Write([]byte(".........."))
			c.setFocusStyle(v, "Port")
		}
		if c.typed.port != "" {
			v.Write([]byte(c.typed.port))
			c.setFocusStyle(v, "Port")
		}
	}

	// Пояснение к действию.
	fmt.Fprintf(view, "%s", strings.Repeat("\n", 12))
	fmt.Fprintf(view, "%sУкажите данные сервера и выполните тест.\n", strings.Repeat(" ", 45))

	// Пояснение по навигации.
	if v, err := g.SetView("TAB", 1, 26, inputWidth-55, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
			return fmt.Errorf("Не удалось установить фокус: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Tab - перевод фокуса"))
	}
	if v, err := g.SetView("Enter", 26, 26, inputWidth-29, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
			return fmt.Errorf("Не удалось установить фокус: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Enter - фиксация ввода"))
	}
	if v, err := g.SetView("MainMenu", 52, 26, inputWidth-3, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
			return fmt.Errorf("Не удалось установить фокус: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+H - главное меню"))
	}
	if v, err := g.SetView("Exit", 78, 26, inputWidth+23, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
			return fmt.Errorf("Не удалось установить фокус: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+C - выход"))
	}

	if v, err := g.SetView("TestConnect", 104, 26, inputWidth+48, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
			return fmt.Errorf("Не удалось установить фокус: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+N - тест связи"))
	}

	// Установка фокуса.
	if _, err := g.SetCurrentView(c.view.currentFocus); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%v>", err)
	}
	layoutInitialized = true

	c.view.activeView = viewSettings // Установка признака активного окна

	return nil
}

// Перевод фокуса.
func (c *handlerUI) nextFocus(g *gocui.Gui, v *gocui.View) error {

	// Перевод фокуса
	switch c.view.activeView {
	case viewRegistration: // Если вызывается из окна регистрации.
		switch c.view.currentFocus {
		case "Login":
			c.view.currentFocus = "Password-1"
		case "Password-1":
			c.view.currentFocus = "Password-2"
		case "Password-2":
			c.view.currentFocus = "Login"
		}

	case viewAutentification: //Если вызывается из окна аутентификации.
		switch c.view.currentFocus {
		case "Login":
			c.view.currentFocus = "Password-1"
		case "Password-1":
			c.view.currentFocus = "Login"
		}

	case viewSettings: //Если вызывается из окна настроек.
		switch c.view.currentFocus {
		case "IP":
			c.view.currentFocus = "Port"
		case "Port":
			c.view.currentFocus = "IP"
		}
	default:
		return nil
	}

	_, err := g.SetCurrentView(c.view.currentFocus)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Ошибка в функции SetCurrentView: <%v>", err))
		return fmt.Errorf("Ошибка в функции SetCurrentView: <%w>", err)
	}

	// Обновляем стили всех полей
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

	for _, name := range fields {
		view, err := g.View(name)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
			return fmt.Errorf("Не удалось установить фокус: <%w>", err)
		}
		if view != nil {
			c.setFocusStyle(view, name)
		}
	}

	// Установка курсора в конец текущей строки
	currentView, err := g.View(c.view.currentFocus)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Ошибка в функции View: <%v>", err))
		return fmt.Errorf("Ошибка в функции View: <%w>", err)
	}
	if currentView != nil {
		buffer := strings.TrimSuffix(currentView.Buffer(), "\n")
		cursorX := len(buffer)
		currentView.SetCursor(cursorX, 0)
	}

	return nil
}

// Выход.
func (c *handlerUI) quit(g *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}

// Обработка нажатия Enter.
func (c *handlerUI) handleEnter(g *gocui.Gui, v *gocui.View) error {

	switch c.view.activeView {
	case viewRegistration: // Если окно registration
		switch v.Name() {
		case "Login":
			c.typed.login = strings.TrimSpace(v.Buffer())
		case "Password-1":
			c.typed.password1 = strings.TrimSpace(v.Buffer())
		case "Password-2":
			c.typed.password2 = strings.TrimSpace(v.Buffer())
		}
	case viewAutentification: // Если окно authentication
		switch v.Name() {
		case "Login":
			c.typed.login = strings.TrimSpace(v.Buffer())
		case "Password-1":
			c.typed.password1 = strings.TrimSpace(v.Buffer())
		}
	case viewSettings: // Если окно settings

		c.flag.checkConnectPassed = false // Сброс признака процесса проверки связи.
		c.flag.checkConnectStatus = false // Сброс статуса результата проверки связи.

		testConnectView, err := g.View("TestConnect")
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Ошибка в функции View: <%v>", err))
			return fmt.Errorf("Ошибка в функции View: <%w>", err)
		}
		testConnectView.FgColor = gocui.ColorWhite // Сбор статусного цвета надписи.

		switch v.Name() {
		case "IP":
			c.typed.ip = strings.TrimSpace(v.Buffer())
		case "Port":
			c.typed.port = strings.TrimSpace(v.Buffer())
		}

	case viewRequestSecretKey: // Если окно с запросом дополнительного ключа шифрования.
		switch v.Name() {
		case "scrtKey":
			c.secret.secretKey = strings.TrimSpace(v.Buffer())
		}
		c.showWindow(g, v) // Отображение данных пользователя.

	default:
		return nil
	}

	return nil
}

// Проверка совпадения паролей при регистрации.
func (c *handlerUI) indicators(g *gocui.Gui) {

	// Окно регистрации.
	if c.view.activeView == viewRegistration {

		// Обработка индикатора проверки введённых данных
		indicator, err := g.View("indicator-match")
		if err != nil || indicator == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View: <%v>", err))
			return
		}
		if c.typed.password1 != "" && c.typed.password2 != "" && c.typed.password1 == c.typed.password2 {
			indicator.Clear()
			indicator.Write([]byte("Данные приняты!"))
			indicator.FgColor = gocui.ColorGreen
			indicator.BgColor = gocui.ColorDefault
		} else {
			indicator.Clear()
			indicator.Write([]byte("Укажите данные"))
			indicator.FgColor = gocui.ColorRed
			indicator.BgColor = gocui.ColorDefault
		}

		// Обработка индикатора процесса регистрации.
		if c.flag.addUserPassed {
			indicator, err = g.View("indicator-registration")
			if err != nil || indicator == nil {
				c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View: <%v>", err))
				return
			}
			if c.flag.addUserSUCCESS {
				indicator.Clear()
				indicator.Write([]byte("Пользователь зарегистрирован. Выполните вход."))
				indicator.FgColor = gocui.ColorGreen
				indicator.BgColor = gocui.ColorDefault
			} else {
				indicator.Clear()
				indicator.Write([]byte("Ошибка при регистрации нового пользователя."))
				indicator.FgColor = gocui.ColorRed
				indicator.BgColor = gocui.ColorDefault
			}
		}

		return
	}

	// Окно настроек.
	if c.view.activeView == viewSettings {

		// Есть установлен признак отработки проверки связи.
		if c.flag.checkConnectPassed {

			// Изменение цвета, в зависимости от результата.
			testConnectView, err := g.View("TestConnect")
			if err == nil {
				if c.flag.checkConnectStatus {
					testConnectView.FgColor = gocui.ColorGreen
				} else {
					testConnectView.FgColor = gocui.ColorRed
				}
			}
		}
	}

	// Окно аутентификации.
}

// Проверка связи с сервером.
func (c *handlerUI) testConnect(g *gocui.Gui, _ *gocui.View) error {

	// Запуск проверки связи с сервером.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Выполнение проверки связи.
	ok, err := pingContext(ctx, c)
	if err != nil {
		c.flag.checkConnectStatus = false
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция pingContext, вернула ошибку: <%v>", err))
		return nil // Возврат nil, чтобы приложение продолжило работу.
	}

	// Результат.
	c.flag.checkConnectStatus = ok
	c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: проверка связи с %s:%s пройдена", c.typed.ip, c.typed.port))
	return nil
}

// Обновляет стиль вида в зависимости от того, имеет ли он фокус
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

// Запуск процесса регистрации нового пользователя.
func (c *handlerUI) doRegistrationUser(gui *gocui.Gui, v *gocui.View) error {

	c.flag.addUserSUCCESS = false // сброс признака успешности регистрации пользователя.
	c.flag.addUserPassed = false  // сброс признака, что процедура регистрации быд запущена.

	userName := c.typed.login
	userPwd1 := c.typed.password1
	userPwd2 := c.typed.password2

	// Проверка корректности введённых пользователем данных
	if err := checkDataRegistration(userName, userPwd1, userPwd2); err != nil {
		c.conf.PtrLoggerFile.Write("Error: данные регистрации не прошли проверку")
		return nil
	}

	// Контекст для запроса.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Добавление пользователя в БД.
	if err := c.conf.DB.AddUserContext(ctx, userName, userPwd1); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция AddUserContext, вернуля ошибку: <%v>", err))
		return nil
	}

	c.flag.addUserPassed = true  // установка признака, что процедура регистрации была запущена.
	c.flag.addUserSUCCESS = true // установка признака, что пользователь зарегистрировался в системе.
	c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: выполнена регистрация пользователя с именем: <%s>", userName))

	return nil
}

// Запуск процесса аутентификации пользователя.
func (c *handlerUI) doAuthenticationUser(gui *gocui.Gui, v *gocui.View) error {

	userName := c.typed.login
	userPwd1 := c.typed.password1

	// Контекст для запроса.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Выполнение запроса.
	ok, err := c.conf.DB.AuthenticateUserContext(ctx, userName, userPwd1)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция AuthenticateUserContext, вернуля ошибку: <%v>", err))
		return nil
	}

	// Обработка результата
	if !ok {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: пользователь <%s>, не прошел аутентификацию.", userName))
		return nil
	}
	c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: пользователь <%s>, прошел аутентификацию.", userName))

	// Открытие окна, с запросом ввода дополнительного кода шифрования.

	c.showRequestEncryptKey(gui, v)

	return nil
}

// Окно с запросом ввода дополнительного ключа шифрования.
func (c *handlerUI) showRequestEncryptKey(g *gocui.Gui, _ *gocui.View) error {
	c.view.activeView = "" // Сброс признака активного окна

	// Удаляем все зависимые виды
	if err := deleteViews(g); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}

	// Сброс состояния
	layoutInitialized = false

	// Создание контейнера запроса ввода дополнительного секретного ключа.
	view, err := g.SetView(viewRequestSecretKey, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	view.Title = "Секретный ключ"
	view.Wrap = true
	view.Clear()

	// Поля ввода
	if v, err := g.SetView("scrtKey", 50, 7, inputWidth+1, inputHeight+1+6); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
			return fmt.Errorf("Не удалось установить фокус: <%w>", err)
		}
		v.Title = "Ключ"
		v.Editable = true
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorBlack
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack
	}

	// Пояснение к действию
	fmt.Fprintf(view, "%s", strings.Repeat("\n", 15))
	fmt.Fprintf(view, "%sОпционально. Укажите дополнительный ключ шифрования.\n", strings.Repeat(" ", 38))
	fmt.Fprintf(view, "%sИли оставьте поле пустым.", strings.Repeat(" ", 38))
	fmt.Fprintf(view, "%s", strings.Repeat("\n", 2))
	fmt.Fprintf(view, "%sНажмите на Enter.\n", strings.Repeat(" ", 38))

	// Установка фокуса на поле ввода "scrtKey"
	if _, err := g.SetCurrentView("scrtKey"); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус на 'scrtKey': <%v>", err))
		return fmt.Errorf("Не удалось установить фокус на 'scrtKey': <%v>", err)
	}

	layoutInitialized = true
	c.view.activeView = viewRequestSecretKey // Установка признака активного окна

	return nil
}

// Окно взаимодействия с данными.
func (c *handlerUI) showWindow(g *gocui.Gui, _ *gocui.View) error { //================= в проработке интерфейса
	c.view.activeView = "" // Сброс признака активного окна

	// Удаляем все зависимые виды
	if err := deleteViews(g); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}

	// Сброс состояния
	layoutInitialized = false

	// Создание контейнера запроса ввода дополнительного секретного ключа.
	view, err := g.SetView(viewUserData, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	view.Title = "Данные"
	view.Wrap = true
	view.Clear()

	//
	// ---
	//

	layoutInitialized = true
	c.view.activeView = viewUserData // Установка признака активного окна

	return nil
}
