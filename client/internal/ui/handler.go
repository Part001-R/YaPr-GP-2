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
	secretKey [32]byte // секретный ключ
}

// Введённые пользователем данные.
type typeData struct {
	login        string // введённое значение в поле имя пользователя
	password1    string // введённое значение в воле пароль
	password2    string // введённое значение в воле пароль (подтверждение)
	ip           string // введённое значение в поле IP
	port         string // введённое значение в поле Port
	dataFor      string // информация - для чего формируются данные
	dataLogin    string // информация - логин
	dataPassword string // информация - пароль
}

// Признаки выполнения логики
type flags struct {
	checkConnectStatus       bool // Результат процедуры проверки связи с сервером.
	checkConnectPassed       bool // Признак, что проверка связи была запущена.
	addUserSUCCESS           bool // Признак успешного добавления пользователя.
	addUserPassed            bool // Признак, что была запущена процедура регистрации пользователя.
	addUserRegBusy           bool // Признак, чот уже есть заргистрированный пользователь
	addLoginPaaswordSUCCESS  bool // Признак успешного добавления пары логин/пароль.
	addLoginPaaswordPassed   bool // Признак, что выполнена процедура добавления пары логин/пароль.
	readLoginPaaswordSUCCESS bool // Признак, успешного получения данных логин/пароль.
	readLoginPaaswordPassed  bool // Признак, что при получении данных логин/пароль, произошла ошибка.
	delLoginPaaswordSUCCESS  bool // Признак, успешного удаления данных логин/пароль.
	delLoginPaaswordPassed   bool // Признак, что при удалении данных логин/пароль, произошла ошибка.
}

// Для навигации по экранам.
type screens struct {
	activeView   string // название активного экрана
	currentFocus string // на какой элемент установлен фокус
}

// Представление записи логин/пароль
type loginPassword struct {
	name      string
	login     string
	password  string
	createdAt string
}

// Данные БД.
type dataDB struct {
	encryptLoginPassword []loginPassword // закодированные данные логин/пароль
	loginPassword        []loginPassword // данные логин/пароль
}

// Индесы.
type indexes struct {
	loginPassword int // текущий индекс для обхода массива логин/пароль
}

// Общий тип для CLI UI.
type handlerUI struct {
	conf   *udt.Configuration // конфигурация сервиса
	typed  typeData           // введённые пользователем данные
	flag   flags              // признаки сервиса
	view   screens            // взаимодействие с окнами
	secret encrKey            // секретность
	data   dataDB             // данные
	index  indexes            // индексы для обхода массивов
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
			secret: encrKey{
				secretKey: [32]byte{},
			},
			data: dataDB{
				encryptLoginPassword: []loginPassword{},
				loginPassword:        []loginPassword{},
			},
			index: indexes{
				loginPassword: 0,
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
		default:
		}

	case viewAutentification: //Если вызывается из окна аутентификации.
		switch c.view.currentFocus {
		case "Login":
			c.view.currentFocus = "Password-1"
		case "Password-1":
			c.view.currentFocus = "Login"
		default:
		}

	case viewSettings: //Если вызывается из окна настроек.
		switch c.view.currentFocus {
		case "IP":
			c.view.currentFocus = "Port"
		case "Port":
			c.view.currentFocus = "IP"
		default:
		}

	case viewSelectType: //Если вызывается из окна выбора типа данных.
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

	case viewLoginPasswordData: // Если окно для взаимодействия с логин/пароль.
		switch c.view.currentFocus {
		case "fieldAddFor":
			c.view.currentFocus = "fieldAddLogin"
		case "fieldAddLogin":
			c.view.currentFocus = "fieldAddPassword"
		case "fieldAddPassword":
			c.view.currentFocus = "fieldAddFor"
		default:
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
	if c.view.activeView == viewSelectType {
		fields = []string{"selectLoginPassword", "selectText", "SelectBinary", "SelectBankCard"}
	}
	if c.view.activeView == viewLoginPasswordData {
		fields = []string{"fieldAddFor", "fieldAddLogin", "fieldAddPassword"}
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
		default:
		}

	case viewRequestSecretKey: // Если окно с запросом дополнительного ключа шифрования.
		switch v.Name() {
		case "scrtKey":
			str := c.typed.login + c.typed.password1 + strings.TrimSpace(v.Buffer())
			c.secret.secretKey = generateSecretKey(str) // создание ключа шифрования из введённых данных.
		default:
		}
		c.showSelectType(g, v) // Отображение данных пользователя.

	case viewSelectType: // Если окно с выбором типа данных.
		// определение, какое окно открыть.
		switch c.view.currentFocus {
		case "selectLoginPassword":
			c.showLoginPassword(g, v)
		case "selectText":
			c.showText(g, v)
		case "SelectBinary":
			c.showBinary(g, v)
		case "SelectBankCard":
			c.showBankCard(g, v)
		default:
		}

	case viewLoginPasswordData: // Если окно взаимодействия с логин/пароль.
		c.flag.addLoginPaaswordPassed = false
		switch v.Name() {
		case "fieldAddFor":
			c.typed.dataFor = strings.TrimSpace(v.Buffer())
		case "fieldAddLogin":
			c.typed.dataLogin = strings.TrimSpace(v.Buffer())
		case "fieldAddPassword":
			c.typed.dataPassword = strings.TrimSpace(v.Buffer())
		default:
		}

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

		// Обработка случая, если при регистрации пользователя, в системе уже присутствует запись.
		if c.flag.addUserRegBusy {
			indicator, err = g.View("indicator-registration")
			if err != nil || indicator == nil {
				c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View: <%v>", err))
				return
			}
			indicator.Clear()
			indicator.Write([]byte("Уже есть зарегистрированный пользователь."))
			indicator.FgColor = gocui.ColorRed
			indicator.BgColor = gocui.ColorDefault
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

	// Окно логин/пароль
	if c.view.activeView == viewLoginPasswordData {

		// Обработка индикатора получения данных.
		indicatorRead, err := g.View("indicatorReadStatus")
		if err != nil || indicatorRead == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при работе с индикатором indicatorReadStatus: <%v>", err))
			return
		}
		if c.flag.readLoginPaaswordPassed { // обработка при чтении
			if c.flag.readLoginPaaswordSUCCESS {
				indicatorRead.Clear()
				indicatorRead.Write([]byte(fmt.Sprintf("Всего записей: %d", len(c.data.loginPassword))))
				indicatorRead.FgColor = gocui.ColorGreen
				indicatorRead.BgColor = gocui.ColorDefault
			} else {
				indicatorRead.Clear()
				indicatorRead.Write([]byte("Ошибка"))
				indicatorRead.FgColor = gocui.ColorRed
				indicatorRead.BgColor = gocui.ColorDefault
			}
		}
		if c.flag.delLoginPaaswordPassed { // обработка при удалении
			if c.flag.delLoginPaaswordSUCCESS {
				indicatorRead.Clear()
				indicatorRead.Write([]byte("Запись удалена"))
				indicatorRead.FgColor = gocui.ColorGreen
				indicatorRead.BgColor = gocui.ColorDefault
			} else {
				indicatorRead.Clear()
				indicatorRead.Write([]byte("Ошибка удаления"))
				indicatorRead.FgColor = gocui.ColorRed
				indicatorRead.BgColor = gocui.ColorDefault
			}
		}

		// Обработка индикатора добавления записи.
		indicator, err := g.View("indicatorAddSuccess")
		if err != nil || indicator == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при работе с индикатором indicatorAddSuccess: <%v>", err))
			return
		}

		if c.flag.addLoginPaaswordPassed {
			if c.flag.addLoginPaaswordSUCCESS {
				indicator.Clear()
				indicator.Write([]byte("Данные приняты!"))
				indicator.FgColor = gocui.ColorGreen
				indicator.BgColor = gocui.ColorDefault
			} else {
				indicator.Clear()
				indicator.Write([]byte("Ошибка добавления."))
				indicator.FgColor = gocui.ColorRed
				indicator.BgColor = gocui.ColorDefault
			}
		} else {
			indicator.Clear()
			indicator.Write([]byte(""))
		}
	}
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
	c.flag.addUserRegBusy = false // сброс признака, что в системе уже есть зарегистрированный пользователь.

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

	// Проверка, что в БД уже есть регистрация пользователя.
	busy, err := c.conf.DB.UserExistContext(ctx)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция UserExistContext, вернуля ошибку: <%v>", err))
		return nil
	}

	if busy {
		c.flag.addUserRegBusy = true // установка признака, что в системе уже есть зарегистрированный пользоатель.
		return nil
	}

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

// Запуск процесса сохранения данных логин/пароль.
func (c *handlerUI) doStoreLoginPasswordDB(gui *gocui.Gui, v *gocui.View) error {

	c.conf.PtrLoggerFile.Write("Debug: запущена функция doStoreLoginPasswordDB")

	c.flag.addLoginPaaswordPassed = true
	c.flag.addLoginPaaswordSUCCESS = false

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Шифрование данных
	encrFor, err := encrypt(c.typed.dataFor, c.secret.secretKey)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого dataFor: <%v>", err))
		return nil
	}
	encrLogin, err := encrypt(c.typed.dataLogin, c.secret.secretKey)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого dataLogin: <%v>", err))
		return nil
	}
	encrPassword, err := encrypt(c.typed.dataPassword, c.secret.secretKey)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого dataPassword: <%v>", err))
		return nil
	}
	tn := time.Now().UTC()
	strT := tn.Format(time.RFC3339)
	encrCreatedAt, err := encrypt(strT, c.secret.secretKey)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого времени создания: <%v>", err))
		return nil
	}

	// Добавление зашифрованных данных в БД.
	if err := c.conf.DB.AddDataLoginPasswordContext(ctx, encrFor, encrLogin, encrPassword, encrCreatedAt); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка добавления пары логин/пароль в БД: <%v>", err))
		return nil
	}

	c.conf.PtrLoggerFile.Write("Debug: пара логин/пароль добавлена в БД")
	c.flag.addLoginPaaswordSUCCESS = true
	return nil
}

// Отображение слудующего элемента.
func (c *handlerUI) doShowNextElement(gui *gocui.Gui, v *gocui.View) error {

	switch c.view.activeView {
	case viewLoginPasswordData: // Взаимодействие с логин/пароль

		el := loginPasswordByIndex(c) // получение записи по индексу
		incrIndexloginPassword(c)     // увеличение значения индекса

		// отображение содержимого поля For.
		fieldName, err := gui.View("fieldShowFor")
		if err != nil || fieldName == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowFor: <%v>", err))
			return nil
		}
		if el.name != "" {
			fieldName.Clear()
			fieldName.Write([]byte(el.name))

		} else {
			fieldName.Clear()
			fieldName.Write([]byte(""))
		}

		// отображение содержимого поля Login.
		fieldLogin, err := gui.View("fieldShowLogin")
		if err != nil || fieldLogin == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowLogin: <%v>", err))
			return nil
		}
		if el.login != "" {
			fieldLogin.Clear()
			fieldLogin.Write([]byte(el.login))

		} else {
			fieldLogin.Clear()
			fieldLogin.Write([]byte(""))
		}

		// отображение содержимого поля Password.
		fieldPassword, err := gui.View("fieldShowPassword")
		if err != nil || fieldPassword == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowPassword: <%v>", err))
			return nil
		}
		if el.password != "" {
			fieldPassword.Clear()
			fieldPassword.Write([]byte(el.password))

		} else {
			fieldPassword.Clear()
			fieldPassword.Write([]byte(""))
		}

	default:
	}

	return nil
}

// Отображение предыдущего элемента.
func (c *handlerUI) doShowPrevElement(gui *gocui.Gui, v *gocui.View) error {

	switch c.view.activeView {
	case viewLoginPasswordData: // Взаимодействие с логин/пароль

		decrIndexloginPassword(c)     // уменьшение значения индекса
		el := loginPasswordByIndex(c) // получение записи по индексу

		// отображение содержимого поля For.
		fieldName, err := gui.View("fieldShowFor")
		if err != nil || fieldName == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowFor: <%v>", err))
			return nil
		}
		if el.name != "" {
			fieldName.Clear()
			fieldName.Write([]byte(el.name))

		} else {
			fieldName.Clear()
			fieldName.Write([]byte(""))
		}

		// отображение содержимого поля Login.
		fieldLogin, err := gui.View("fieldShowLogin")
		if err != nil || fieldLogin == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowLogin: <%v>", err))
			return nil
		}
		if el.login != "" {
			fieldLogin.Clear()
			fieldLogin.Write([]byte(el.login))

		} else {
			fieldLogin.Clear()
			fieldLogin.Write([]byte(""))
		}

		// отображение содержимого поля Password.
		fieldPassword, err := gui.View("fieldShowPassword")
		if err != nil || fieldPassword == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowPassword: <%v>", err))
			return nil
		}
		if el.password != "" {
			fieldPassword.Clear()
			fieldPassword.Write([]byte(el.password))

		} else {
			fieldPassword.Clear()
			fieldPassword.Write([]byte(""))
		}

	default:
	}

	return nil
}

// Удаление записи.
func (c *handlerUI) doDeleteElement(gui *gocui.Gui, v *gocui.View) error {

	switch c.view.activeView {
	case viewLoginPasswordData: // Взаимодействие с логин/пароль

		c.flag.delLoginPaaswordPassed = true
		c.flag.delLoginPaaswordSUCCESS = false

		v, err := gui.View("fieldShowFor")
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка получения вида fieldShowFor, при удалении записи логин/пароль: <%v>", err))
			return nil
		}
		textEl := v.Buffer() // Получаем содержимое поля ввода
		textEl = strings.ReplaceAll(textEl, "\n", "")

		textEl, err = encrypt(textEl, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция encrypt, вернула ошибку: <%v>", err))
			return nil
		}

		// Удаление записи в БД.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		if err := c.conf.DB.DelDataLoginPasswordContext(ctx, textEl); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция DelDataLoginPasswordContext, вернула ошибку: <%v>", err))
			return nil
		}

		c.flag.delLoginPaaswordSUCCESS = true
		c.conf.PtrLoggerFile.Write(("Debug: данные логин/пароль, успешно удалены"))
	default:
	}

	return nil
}

// Окно с запросом ввода дополнительного ключа шифрования.
func (c *handlerUI) showRequestEncryptKey(g *gocui.Gui, _ *gocui.View) error {

	c.view.activeView = "" // Сброс признака активного окна

	// Очистка.
	if err := deleteViews(g); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}

	// Сброс состояния.
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

	//
	// --- Поля ввода ---
	//
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

	//
	// --- Пояснение к действию ---
	//
	fmt.Fprintf(view, "%s", strings.Repeat("\n", 15))
	fmt.Fprintf(view, "%sОпционально. Укажите дополнительный ключ шифрования.\n", strings.Repeat(" ", 38))
	fmt.Fprintf(view, "%sИли оставьте поле пустым.", strings.Repeat(" ", 38))
	fmt.Fprintf(view, "%s", strings.Repeat("\n", 2))
	fmt.Fprintf(view, "%sНажмите на Enter.\n", strings.Repeat(" ", 38))

	//
	// --- Пояснение по навигации ---
	//
	if v, err := g.SetView("Enter", 1, 26, inputWidth-55, inputHeight+1+25); err != nil {
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
	if v, err := g.SetView("MainMenu", 26, 26, inputWidth-29, inputHeight+1+25); err != nil {
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
	if v, err := g.SetView("Exit", 52, 26, inputWidth-3, inputHeight+1+25); err != nil {
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

	// Установка фокуса на поле ввода "scrtKey"
	if _, err := g.SetCurrentView("scrtKey"); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус на 'scrtKey': <%v>", err))
		return fmt.Errorf("Не удалось установить фокус на 'scrtKey': <%v>", err)
	}

	layoutInitialized = true
	c.view.activeView = viewRequestSecretKey // Установка признака активного окна

	return nil
}

// Окно с выбором типа записей.
func (c *handlerUI) showSelectType(g *gocui.Gui, _ *gocui.View) error {

	c.conf.PtrLoggerFile.Write(fmt.Sprintf("Debug: выполнен переход на окно: <%s>", viewSelectType))

	c.view.activeView = "" // Сброс признака активного окна

	// Удаляем все зависимые виды
	if err := deleteViews(g); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}

	// Сброс состояния
	layoutInitialized = false
	c.view.currentFocus = "selectLoginPassword" // Установка фокуса

	// Создание контейнера запроса ввода дополнительного секретного ключа.
	view, err := g.SetView(viewSelectType, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	view.Title = "Тип данных"
	view.Wrap = true
	view.Clear()

	//
	// --- Отображение разделов ---
	//
	if v, err := g.SetView("selectLoginPassword", 50, 2, inputWidth+1, inputHeight+1+1); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorBlack
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack

		v.Write([]byte("логин/пароль"))
		c.setFocusStyle(v, "selectLoginPassword")
	}

	if v, err := g.SetView("selectText", 50, 5, inputWidth+1, inputHeight+1+4); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorBlack
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack

		v.Write([]byte("текстовые данные"))
		c.setFocusStyle(v, "selectText")
	}

	if v, err := g.SetView("SelectBinary", 50, 8, inputWidth+1, inputHeight+1+7); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorBlack
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack

		v.Write([]byte("бинарные данные"))
		c.setFocusStyle(v, "SelectBinary")
	}
	if v, err := g.SetView("SelectBankCard", 50, 11, inputWidth+1, inputHeight+1+10); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorBlack
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack

		v.Write([]byte("банковские карты"))
		c.setFocusStyle(v, "SelectBankCard")
	}

	//
	// --- Пояснение к действию ---
	//
	fmt.Fprintf(view, "%s", strings.Repeat("\n", 15))
	fmt.Fprintf(view, "%sВыберите нужный раздел через Tab и нажмите Enter.\n", strings.Repeat(" ", 40))

	//
	// --- Нижняя часть экрана ---
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
		v.Write([]byte("Enter - переход"))
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

	// Установка фокуса.
	if _, err := g.SetCurrentView(c.view.currentFocus); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	layoutInitialized = true
	c.view.activeView = viewSelectType // Установка признака активного окна
	c.conf.PtrLoggerFile.Write(fmt.Sprintf("Debug: в окне: <%s>, установлен фокус на: <%s>", viewSelectType, c.view.activeView))

	return nil
}

// Окно для взаимодействия с логин/пароль.
func (c *handlerUI) showLoginPassword(g *gocui.Gui, _ *gocui.View) error {

	c.conf.PtrLoggerFile.Write("Debug: выполнен вход в окно typeLoginPassword")

	c.flag.readLoginPaaswordPassed = false // Сброс признака.
	c.flag.addLoginPaaswordPassed = false
	c.flag.delLoginPaaswordPassed = false
	c.index.loginPassword = 0 // Сброс индекса навигации по массиву логин/пароль.

	// Логика.
	//
	c.view.activeView = "" // Сброс признака активного окна

	// Очистка.
	if err := deleteViews(g); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}

	// Сброс состояния
	layoutInitialized = false
	c.view.currentFocus = "fieldAddFor" // Установка фокуса

	// Создание контейнера запроса ввода дополнительного секретного ключа.
	view, err := g.SetView(viewLoginPasswordData, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	view.Title = "Логины - пароли"
	view.Wrap = true
	view.Clear()

	//
	// --- Отображение разделов ---
	//
	fmt.Fprintf(view, "%s", strings.Repeat("\n", 3))
	fmt.Fprintf(view, "%sПросмотр.\n", strings.Repeat(" ", 56))

	if v, err := g.SetView("fieldShowFor", 1, 4, inputWidth-48, inputHeight+1+3); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Title = "Для"
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault

		v.Write([]byte("....."))
		c.setFocusStyle(v, "fieldShowFor")
	}
	if v, err := g.SetView("fieldShowLogin", 33, 4, inputWidth-4, inputHeight+1+3); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Title = "Логин"
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault

		v.Write([]byte("....."))
		c.setFocusStyle(v, "fieldShowLogin")
	}
	if v, err := g.SetView("fieldShowPassword", 77, 4, inputWidth+48, inputHeight+1+3); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Title = "Пароль"
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault

		v.Write([]byte("....."))
		c.setFocusStyle(v, "fieldShowPassword")
	}

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 7))
	fmt.Fprintf(view, "%sДобавление.\n", strings.Repeat(" ", 55))

	if v, err := g.SetView("fieldAddFor", 1, 12, inputWidth-48, inputHeight+1+11); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Title = "Для"
		v.Editable = true
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorBlack
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack

		v.Write([]byte("....."))
		c.setFocusStyle(v, "fieldAddFor")
	}
	if v, err := g.SetView("fieldAddLogin", 33, 12, inputWidth-4, inputHeight+1+11); err != nil {
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

		v.Write([]byte("....."))
		c.setFocusStyle(v, "fieldAddLogin")
	}
	if v, err := g.SetView("fieldAddPassword", 77, 12, inputWidth+48, inputHeight+1+11); err != nil {
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

		v.Write([]byte("....."))
		c.setFocusStyle(v, "fieldAddPassword")
	}

	//
	// --- Индикаторы ---
	//

	// Результат чтения данных.
	indicatorY := inputHeight * 3
	indicatorX := 51
	vRead, err := g.SetView("indicatorReadStatus", indicatorX, indicatorY, indicatorX+20, indicatorY+2)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
	}
	vRead.Frame = false
	vRead.BgColor = gocui.ColorDefault
	vRead.FgColor = gocui.ColorDefault

	// Результат добавления данных.
	indicatorY = inputHeight*3 + 8
	indicatorX = 53
	vAdd, err := g.SetView("indicatorAddSuccess", indicatorX, indicatorY, indicatorX+20, indicatorY+2)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
	}
	vAdd.Frame = false
	vAdd.BgColor = gocui.ColorDefault
	vAdd.FgColor = gocui.ColorDefault

	//
	// --- Краткое пояснение ---
	//

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 8))
	fmt.Fprintf(view, "%sПри изменении данных, выполните Crl+U.\n", strings.Repeat(" ", 42))

	//
	// --- Нижняя часть экрана ---
	//

	// Верхний ряд.
	if v, err := g.SetView("Save", 1, 23, inputWidth-55, inputHeight+1+22); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+F - сохранение"))
	}
	if v, err := g.SetView("NextElement", 26, 23, inputWidth-29, inputHeight+1+22); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctr+E - Далее"))
	}
	if v, err := g.SetView("PrevElement", 52, 23, inputWidth-3, inputHeight+1+22); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+G - Назад"))
	}
	if v, err := g.SetView("DeleteElement", 78, 23, inputWidth+23, inputHeight+1+22); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+J - Удаление"))
	}

	//Нижний ряд.
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
		v.Write([]byte("Enter - фиксация"))
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
	if v, err := g.SetView("Back", 104, 26, inputWidth+48, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+U - назад"))
	}

	// Получение сохранённых значений логин/пароль.
	c.flag.readLoginPaaswordPassed = true // Установка признака, что был запущен процесс получения значений логин/пароль.

	_, err = showLoginPasswordWorkDB(c)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция funcshowLoginPasswordWorkDB, вернула ошибку: <%v>", err))
		c.flag.readLoginPaaswordSUCCESS = false
	} else {
		c.conf.PtrLoggerFile.Write("Debug: данные логин/пароль успешно прочитаны")
		c.flag.readLoginPaaswordSUCCESS = true
	}

	// Установка фокуса.
	if _, err := g.SetCurrentView(c.view.currentFocus); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	layoutInitialized = true
	c.view.activeView = viewLoginPasswordData // Установка признака активного окна

	return nil
}

// Окно для взаимодействия с логин/пароль.
func (c *handlerUI) showText(g *gocui.Gui, _ *gocui.View) error {

	c.view.activeView = "" // Сброс признака активного окна

	// Удаляем все зависимые виды
	if err := deleteViews(g); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}

	// Сброс состояния
	layoutInitialized = false
	c.view.currentFocus = "selectLoginPassword" // Установка фокуса

	// Создание контейнера запроса ввода дополнительного секретного ключа.
	view, err := g.SetView(viewTextdData, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	view.Title = "Текст"
	view.Wrap = true
	view.Clear()

	//
	// --- Отображение разделов ---
	//

	//
	// --- Нижняя часть экрана ---
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
		v.Write([]byte("Enter - переход"))
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
	if v, err := g.SetView("Back", 104, 26, inputWidth+48, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+U - назад"))
	}

	// Установка фокуса.
	if _, err := g.SetCurrentView(c.view.currentFocus); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	layoutInitialized = true
	c.view.activeView = viewTextdData // Установка признака активного окна

	return nil
}

// Окно для взаимодействия с логин/пароль.
func (c *handlerUI) showBinary(g *gocui.Gui, _ *gocui.View) error {

	c.view.activeView = "" // Сброс признака активного окна

	// Удаляем все зависимые виды
	if err := deleteViews(g); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}

	// Сброс состояния
	layoutInitialized = false
	c.view.currentFocus = "selectLoginPassword" // Установка фокуса

	// Создание контейнера запроса ввода дополнительного секретного ключа.
	view, err := g.SetView(viewBinaryData, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	view.Title = "Файлы"
	view.Wrap = true
	view.Clear()

	//
	// --- Нижняя часть экрана ---
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
		v.Write([]byte("Enter - переход"))
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
	if v, err := g.SetView("Back", 104, 26, inputWidth+48, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+U - назад"))
	}

	// Установка фокуса.
	if _, err := g.SetCurrentView(c.view.currentFocus); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	layoutInitialized = true
	c.view.activeView = viewBinaryData // Установка признака активного окна

	return nil
}

// Окно для взаимодействия с логин/пароль.
func (c *handlerUI) showBankCard(g *gocui.Gui, _ *gocui.View) error {

	c.view.activeView = "" // Сброс признака активного окна

	// Удаляем все зависимые виды
	if err := deleteViews(g); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}

	// Сброс состояния
	layoutInitialized = false
	c.view.currentFocus = "selectLoginPassword" // Установка фокуса

	// Создание контейнера запроса ввода дополнительного секретного ключа.
	view, err := g.SetView(viewBankCardData, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	view.Title = "Банковские карты"
	view.Wrap = true
	view.Clear()

	//
	// --- Нижняя часть экрана ---
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
		v.Write([]byte("Enter - переход"))
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
	if v, err := g.SetView("Back", 104, 26, inputWidth+48, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+U - назад"))
	}

	// Установка фокуса.
	if _, err := g.SetCurrentView(c.view.currentFocus); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	layoutInitialized = true
	c.view.activeView = viewBankCardData // Установка признака активного окна

	return nil
}
