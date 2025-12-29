package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Part001-R/YaPr-GP-2/client/internal/container"
	"github.com/Part001-R/YaPr-GP-2/client/internal/service/udt"
	"github.com/jroimartin/gocui"
)

const (
	containerName = "container.data"
)

var once sync.Once

type encrKey struct {
	secretKey [32]byte // секретный ключ
}

// Введённые пользователем данные.
type typeData struct {
	login         string // имя пользователя.
	password1     string // пароль.
	password2     string // пароль (подтверждение).
	ip            string // IP.
	port          string // Port.
	dataFor       string // для чего формируются данные.
	dataLogin     string // логин.
	dataPassword  string // пароль.
	dataText      string // текст.
	dataOwner     string // владелец.
	dataNumb      string // номер.
	dataValidDate string // дата валидности.
	dataCode      string // код.
	dataPathSrc   string // путь, нахождения файла.
	dataPathTrg   string // путь, куда нужно поместить файл.
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
	readLoginPaaswordPassed  bool // Признак, что процедура чтения логин/пароль, пройдена.
	delLoginPaaswordSUCCESS  bool // Признак, успешного удаления данных логин/пароль.
	delLoginPaaswordPassed   bool // Признак, что процедура удаления логин/пароль, пройдена.
	addTextSUCCESS           bool // Признак успешного добавления текста.
	addTextPassed            bool // Признак, что выполнена процедура добавления текста.
	readTextSUCCESS          bool // Признак, успешного получения данных текста.
	readTextPassed           bool // Признак, что процедура получения текста, пройдена.
	delTextSUCCESS           bool // Признак, успешного удаления данных текста.
	delTextPassed            bool // Признак, что процедура удаления текста, пройдена.
	addBankCardSUCCESS       bool // Признак успешного добавления карты.
	addBankCardPassed        bool // Признак, что выполнена процедура добавления карты.
	readBankCardSUCCESS      bool // Признак, успешного получения данных карт.
	readBankCardPassed       bool // Признак, что процедура получения данных карт, пройдена.
	delBankCardSUCCESS       bool // Признак, успешного удаления данных карты.
	delBankCardPassed        bool // Признак, что процедура удаления карты, пройдена.
	addFileSUCCESS           bool // Признак успешного добавления файла.
	addFilePassed            bool // Признак, что выполнена процедура добавления файла.
	readFileSUCCESS          bool // Признак, успешного получения данных файла.
	readFilePassed           bool // Признак, что процедура получения данных файла, пройдена.
	delFileSUCCESS           bool // Признак, успешного удаления файла.
	delFilePassed            bool // Признак, что процедура удаления файла, пройдена.
	extractFileSUCCESS       bool // Признак, успешного извлечения файла.
	extractFilePassed        bool // Признак, что процедура извлечения файла, пройдена.

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

// Представление записи - текст.
type textData struct {
	name      string
	text      string
	createdAt string
}

// Представление записи - банковская карта.
type bankCard struct {
	name      string
	owner     string
	numb      string
	valid     string
	code      string
	createdAt string
}

// Данные БД.
type data struct {
	encryptLoginPassword []loginPassword // закодированные данные - логин/пароль.
	loginPassword        []loginPassword // данные - логин/пароль.
	encryptTextData      []textData      // закодированные данные - текст.
	textData             []textData      // данные - текст.
	encryptBankCard      []bankCard      // закодированные данные - банковские карты.
	bankCard             []bankCard      // данные - банковские карты.
	files                []string        // файлы
}

// Индесы.
type indexes struct {
	loginPassword int // текущий индекс для обхода массива - логин/пароль.
	text          int // текущий индекс для обхода массива - текст.
	bankCard      int // текущий индекс для обхода массива - банковские карты.
	file          int // текущий индекс для обхода массива - файлы.
}

// Общий тип для CLI UI.
type handlerUI struct {
	conf   *udt.Configuration // конфигурация сервиса
	typed  typeData           // введённые пользователем данные
	flag   flags              // признаки сервиса
	view   screens            // взаимодействие с окнами
	secret encrKey            // секретность
	data   data               // данные
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
			data: data{
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

	case viewTextData: // Если окно для взаимодействия с текстом.
		switch c.view.currentFocus {
		case "fieldAddFor":
			c.view.currentFocus = "fieldAddText"
		case "fieldAddText":
			c.view.currentFocus = "fieldAddFor"
		default:
		}

	case viewBankCardData: // Если окно для взаимодействия с банковскими картами.
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

	case viewBinaryData: // Если окно для взаимодействия с файлами.
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
		c.flag.addLoginPaaswordSUCCESS = false

		switch v.Name() {
		case "fieldAddFor":
			c.typed.dataFor = strings.TrimSpace(v.Buffer())
		case "fieldAddLogin":
			c.typed.dataLogin = strings.TrimSpace(v.Buffer())
		case "fieldAddPassword":
			c.typed.dataPassword = strings.TrimSpace(v.Buffer())
		default:
		}

	case viewTextData: // Если окно взаимодействия с текстом.
		c.flag.addTextPassed = false  // Сброс признака.
		c.flag.addTextSUCCESS = false // Сброс статуса.

		switch v.Name() {
		case "fieldAddFor":
			c.typed.dataFor = strings.TrimSpace(v.Buffer())
		case "fieldAddText":
			c.typed.dataText = strings.TrimSpace(v.Buffer())
		default:
		}

	case viewBankCardData: // Если окно взаимодействия с банковскими картами.
		c.flag.addBankCardPassed = false  // Сброс признака.
		c.flag.addBankCardSUCCESS = false // Сброс статуса.

		switch v.Name() {
		case "fieldAddFor":
			c.typed.dataFor = strings.TrimSpace(v.Buffer())
		case "fieldAddOwner":
			c.typed.dataOwner = strings.TrimSpace(v.Buffer())
		case "fieldAddNumber":
			c.typed.dataNumb = strings.TrimSpace(v.Buffer())
		case "fieldAddValid":
			c.typed.dataValidDate = strings.TrimSpace(v.Buffer())
		case "fieldAddCode":
			c.typed.dataCode = strings.TrimSpace(v.Buffer())
		default:
		}

	case viewBinaryData: // Если окно взаимодействия с файлами.
		c.flag.addFilePassed = false  // Сброс признака.
		c.flag.addFileSUCCESS = false // Сброс статуса.

		switch v.Name() {
		case "fieldPathSource":
			c.typed.dataPathSrc = strings.ReplaceAll(c.typed.dataPathSrc, "\n", "")
			c.typed.dataPathSrc = strings.TrimSpace(v.Buffer())
		case "fieldPathTarget":
			c.typed.dataPathTrg = strings.ReplaceAll(c.typed.dataPathTrg, "\n", "")
			c.typed.dataPathTrg = strings.TrimSpace(v.Buffer())
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

	// Окно текста.
	if c.view.activeView == viewTextData {

		// Обработка индикатора получения данных.
		indicatorRead, err := g.View("indicatorReadStatus")
		if err != nil || indicatorRead == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при работе с индикатором indicatorReadStatus: <%v>", err))
			return
		}
		if c.flag.readTextPassed { // обработка при чтении
			if c.flag.readTextSUCCESS {
				indicatorRead.Clear()
				indicatorRead.Write([]byte(fmt.Sprintf("Всего записей: %d", len(c.data.textData))))
				indicatorRead.FgColor = gocui.ColorGreen
				indicatorRead.BgColor = gocui.ColorDefault
			} else {
				indicatorRead.Clear()
				indicatorRead.Write([]byte("Ошибка"))
				indicatorRead.FgColor = gocui.ColorRed
				indicatorRead.BgColor = gocui.ColorDefault
			}
		}
		if c.flag.delTextPassed { // обработка при удалении
			if c.flag.delTextSUCCESS {
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

		if c.flag.addTextPassed {
			if c.flag.addTextSUCCESS {
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

	// Окно банковских карт.
	if c.view.activeView == viewBankCardData {

		// Обработка индикатора получения данных.
		indicatorRead, err := g.View("indicatorReadStatus")
		if err != nil || indicatorRead == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при работе с индикатором indicatorReadStatus: <%v>", err))
			return
		}
		if c.flag.readBankCardPassed { // обработка при чтении
			if c.flag.readBankCardSUCCESS {
				indicatorRead.Clear()
				indicatorRead.Write([]byte(fmt.Sprintf("Всего записей: %d", len(c.data.bankCard))))
				indicatorRead.FgColor = gocui.ColorGreen
				indicatorRead.BgColor = gocui.ColorDefault
			} else {
				indicatorRead.Clear()
				indicatorRead.Write([]byte("Ошибка"))
				indicatorRead.FgColor = gocui.ColorRed
				indicatorRead.BgColor = gocui.ColorDefault
			}
		}
		if c.flag.delBankCardPassed { // обработка при удалении
			if c.flag.delBankCardSUCCESS {
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

		if c.flag.addBankCardPassed {
			if c.flag.addBankCardSUCCESS {
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

	// Окно файлов.
	if c.view.activeView == viewBinaryData {

		// Есть установлен признак отработки добавления файла в контейнер.
		if c.flag.addFilePassed {
			element, err := g.View("Save")
			if err != nil || element == nil {
				c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка доступа к элементу Save: <%v>", err))
			}
			if err == nil {
				if c.flag.addFileSUCCESS {
					element.FgColor = gocui.ColorGreen
				} else {
					element.FgColor = gocui.ColorRed
				}
			}
		}

		// Есть установлен признак извлечения файла из контейнера.
		if c.flag.extractFilePassed {
			element, err := g.View("Extraction")
			if err != nil || element == nil {
				c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка доступа к элементу Extraction: <%v>", err))
			}
			if err == nil {
				if c.flag.extractFileSUCCESS {
					element.FgColor = gocui.ColorGreen
				} else {
					element.FgColor = gocui.ColorRed
				}
			}
		}

		// Есть установлен признак удаления файла из контейнера.
		if c.flag.delFilePassed {
			element, err := g.View("DeleteElement")
			if err != nil || element == nil {
				c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка доступа к элементу DeleteElement: <%v>", err))
			}
			if err == nil {
				if c.flag.delFileSUCCESS {
					element.FgColor = gocui.ColorGreen
				} else {
					element.FgColor = gocui.ColorRed
				}
			}

			c.flag.delFilePassed = false
			c.flag.delFileSUCCESS = false
		}

		// Если чтение файлов контейнера выполнено.
		if c.flag.readFilePassed {

			indicator, err := g.View("indicatorReadStatus")
			if err != nil || indicator == nil {
				c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при работе с индикатором indicatorReadStatus: <%v>", err))
				return
			}
			if c.flag.readFileSUCCESS {

				c.conf.PtrLoggerFile.Write(fmt.Sprintf("--- Debug: Есть признак чтения файлов. В хранилище <%d> файлов", len(c.data.files))) //================

				indicator.Clear()
				indicator.Write([]byte(fmt.Sprintf("Всего файлов: %d", len(c.data.files))))
				indicator.FgColor = gocui.ColorGreen
				indicator.BgColor = gocui.ColorDefault
			} else {
				indicator.Clear()
				indicator.Write([]byte("Ошибка чтения."))
				indicator.FgColor = gocui.ColorRed
				indicator.BgColor = gocui.ColorDefault
			}

			c.flag.readFilePassed = false  // Для разовой отработки при открытии экрана.
			c.flag.readFileSUCCESS = false // Для разовой отработки при открытии экрана.
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

// Запуск процесса сохранения данных.
func (c *handlerUI) doStore(gui *gocui.Gui, v *gocui.View) error {

	c.conf.PtrLoggerFile.Write("Debug: запущена функция doStoreLoginPasswordDB")

	switch c.view.activeView {
	case viewLoginPasswordData: // Если окно - логин/пароль

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

	case viewTextData: // если окно - текст.

		c.conf.PtrLoggerFile.Write(fmt.Sprintf("--- Debug: Добавляются данные For:<%s> Text:<%s>", c.typed.dataFor, c.typed.dataText)) //====================

		c.flag.addTextPassed = true
		c.flag.addTextSUCCESS = false

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		// Шифрование данных
		encrFor, err := encrypt(c.typed.dataFor, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого dataFor: <%v>", err))
			return nil
		}
		encrText, err := encrypt(c.typed.dataText, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого dataText: <%v>", err))
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
		if err := c.conf.DB.AddDataTextContext(ctx, encrFor, encrText, encrCreatedAt); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка добавления текста в БД: <%v>", err))
			return nil
		}

		c.conf.PtrLoggerFile.Write("Debug: текст добавлен в БД")
		c.flag.addTextSUCCESS = true
		return nil

	case viewBankCardData: // если окно - банковские карты.

		c.flag.addBankCardPassed = true
		c.flag.addBankCardSUCCESS = false

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		// Шифрование данных
		encrFor, err := encrypt(c.typed.dataFor, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого dataFor: <%v>", err))
			return nil
		}
		encrOwner, err := encrypt(c.typed.dataOwner, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого dataOwner: <%v>", err))
			return nil
		}
		encrNumb, err := encrypt(c.typed.dataNumb, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого dataNumb: <%v>", err))
			return nil
		}
		encrValid, err := encrypt(c.typed.dataValidDate, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого dataValidDate: <%v>", err))
			return nil
		}
		encrCode, err := encrypt(c.typed.dataCode, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого dataCode: <%v>", err))
			return nil
		}
		tn := time.Now().UTC()
		strT := tn.Format(time.RFC3339)
		encrCreatedAt, err := encrypt(strT, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого времени создания: <%v>", err))
			return nil
		}

		// Проверка номера банковской карты на валидность.
		if !checkCardNumber(c.typed.dataNumb) {
			c.conf.PtrLoggerFile.Write("Error: номер карты, не прошел проверку")
			return nil
		}

		// Добавление зашифрованных данных в БД.
		if err := c.conf.DB.AddDataBankCardContext(ctx, encrFor, encrOwner, encrNumb, encrValid, encrCode, encrCreatedAt); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка добавления карты в БД: <%v>", err))
			return nil
		}

		c.conf.PtrLoggerFile.Write("Debug: карта добавлена в БД")
		c.flag.addBankCardSUCCESS = true
		return nil

	case viewBinaryData: // Окно для работы с файлами.

		c.flag.addFilePassed = true
		c.flag.addFileSUCCESS = false

		if err := c.conf.Container.AddFileToContainer(c.typed.dataPathSrc, c.secret.secretKey); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка при добавлении файла <%s>, в контейнер", c.typed.dataPathSrc))
			return nil
		}
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: файл <%s>, добавлен в контейнер", c.typed.dataPathSrc))
		c.flag.addFileSUCCESS = true

	default:
	}

	return nil

}

// Отображение слудующего элемента.
func (c *handlerUI) doShowNextElement(gui *gocui.Gui, v *gocui.View) error {

	switch c.view.activeView {
	case viewLoginPasswordData: // Взаимодействие с логин/пароль

		if len(c.data.loginPassword) == 0 {
			return nil
		}

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

	case viewTextData: // Взаимодействие с текстом

		if len(c.data.textData) == 0 {
			return nil
		}

		el := textByIndex(c) // получение записи по индексу
		incrIndexText(c)     // увеличение значения индекса

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
		fieldText, err := gui.View("fieldShowText")
		if err != nil || fieldText == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowText: <%v>", err))
			return nil
		}
		if el.text != "" {
			fieldText.Clear()
			fieldText.Write([]byte(el.text))

		} else {
			fieldText.Clear()
			fieldText.Write([]byte(""))
		}
	case viewBankCardData: // Взаимодействие с банковскими картами

		if len(c.data.bankCard) == 0 {
			return nil
		}

		el := bankCardByIndex(c) // получение записи по индексу
		incrIndexBankCard(c)     // увеличение значения индекса

		// отображение содержимого поля Для.
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

		// отображение содержимого поля Владелец.
		fieldOwner, err := gui.View("fieldShowOwner")
		if err != nil || fieldOwner == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowOwner: <%v>", err))
			return nil
		}
		if el.owner != "" {
			fieldOwner.Clear()
			fieldOwner.Write([]byte(el.owner))

		} else {
			fieldOwner.Clear()
			fieldOwner.Write([]byte(""))
		}

		// отображение содержимого поля Номер.
		fieldNumb, err := gui.View("fieldShowNumber")
		if err != nil || fieldNumb == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowNumber: <%v>", err))
			return nil
		}
		if el.numb != "" {
			fieldNumb.Clear()
			fieldNumb.Write([]byte(el.numb))

		} else {
			fieldNumb.Clear()
			fieldNumb.Write([]byte(""))
		}

		// отображение содержимого поля Валидность.
		fieldValid, err := gui.View("fieldShowValid")
		if err != nil || fieldValid == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowValid: <%v>", err))
			return nil
		}
		if el.valid != "" {
			fieldValid.Clear()
			fieldValid.Write([]byte(el.valid))

		} else {
			fieldValid.Clear()
			fieldValid.Write([]byte(""))
		}

		// отображение содержимого поля Код.
		fieldCode, err := gui.View("fieldShowCode")
		if err != nil || fieldCode == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowCode: <%v>", err))
			return nil
		}
		if el.code != "" {
			fieldCode.Clear()
			fieldCode.Write([]byte(el.code))

		} else {
			fieldCode.Clear()
			fieldCode.Write([]byte(""))
		}

	case viewBinaryData: // Взаимодействие с файлами

		if len(c.data.files) == 0 {
			return nil
		}

		el := fileByIndex(c) // получение записи по индексу
		incrIndexFile(c)     // увеличение значения индекса

		// отображение содержимого поля Код.
		fieldCode, err := gui.View("fieldShowFor")
		if err != nil || fieldCode == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowFor: <%v>", err))
			return nil
		}
		if el != "" {
			fieldCode.Clear()
			fieldCode.Write([]byte(el))

		} else {
			fieldCode.Clear()
			fieldCode.Write([]byte(""))
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

	case viewTextData: // Взаимодействие с текст

		decrIndexText(c)     // уменьшение значения индекса
		el := textByIndex(c) // получение записи по индексу

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
		fieldText, err := gui.View("fieldShowText")
		if err != nil || fieldText == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowText: <%v>", err))
			return nil
		}
		if el.text != "" {
			fieldText.Clear()
			fieldText.Write([]byte(el.text))

		} else {
			fieldText.Clear()
			fieldText.Write([]byte(""))
		}
	case viewBankCardData: // Взаимодействие с банковскими картами

		decrIndexBankCard(c)     // уменьшение значения индекса
		el := bankCardByIndex(c) // получение записи по индексу

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

		// отображение содержимого поля Владелец.
		fieldOwner, err := gui.View("fieldShowOwner")
		if err != nil || fieldOwner == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowOwner: <%v>", err))
			return nil
		}
		if el.owner != "" {
			fieldOwner.Clear()
			fieldOwner.Write([]byte(el.owner))

		} else {
			fieldOwner.Clear()
			fieldOwner.Write([]byte(""))
		}

		// отображение содержимого поля Номер.
		fieldNumb, err := gui.View("fieldShowNumber")
		if err != nil || fieldNumb == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowNumber: <%v>", err))
			return nil
		}
		if el.numb != "" {
			fieldNumb.Clear()
			fieldNumb.Write([]byte(el.numb))

		} else {
			fieldNumb.Clear()
			fieldNumb.Write([]byte(""))
		}

		// отображение содержимого поля Валидность.
		fieldValid, err := gui.View("fieldShowValid")
		if err != nil || fieldValid == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowValid: <%v>", err))
			return nil
		}
		if el.valid != "" {
			fieldValid.Clear()
			fieldValid.Write([]byte(el.valid))

		} else {
			fieldValid.Clear()
			fieldValid.Write([]byte(""))
		}

		// отображение содержимого поля Код.
		fieldCode, err := gui.View("fieldShowCode")
		if err != nil || fieldCode == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowCode: <%v>", err))
			return nil
		}
		if el.code != "" {
			fieldCode.Clear()
			fieldCode.Write([]byte(el.code))

		} else {
			fieldCode.Clear()
			fieldCode.Write([]byte(""))
		}

	case viewBinaryData: // Взаимодействие с файлами.

		decrIndexFile(c)     // уменьшение значения индекса
		el := fileByIndex(c) // получение записи по индексу

		// отображение содержимого.
		fieldCode, err := gui.View("fieldShowFor")
		if err != nil || fieldCode == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowFor: <%v>", err))
			return nil
		}
		if el != "" {
			fieldCode.Clear()
			fieldCode.Write([]byte(el))

		} else {
			fieldCode.Clear()
			fieldCode.Write([]byte(""))
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

	case viewTextData: // Взаимодействие с текстом

		c.flag.delTextPassed = true
		c.flag.delTextSUCCESS = false

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
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := c.conf.DB.DelTextContext(ctx, textEl); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция DelTextContext, вернула ошибку: <%v>", err))
			return nil
		}

		c.flag.delTextSUCCESS = true
		c.conf.PtrLoggerFile.Write(("Debug: данные текста, успешно удалены"))

	case viewBankCardData: // Взаимодействие с банковскими картами

		c.flag.delBankCardPassed = true
		c.flag.delBankCardSUCCESS = false

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
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := c.conf.DB.DelBankCardContext(ctx, textEl); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция DelBankCardContext, вернула ошибку: <%v>", err))
			return nil
		}

		c.flag.delBankCardSUCCESS = true
		c.conf.PtrLoggerFile.Write(("Debug: данные карты, успешно удалены"))

	case viewBinaryData: // Окно работы с файлами.

		c.flag.delFilePassed = true
		c.flag.delFileSUCCESS = false

		// Чтение буфера.
		v, err := gui.View("fieldShowFor")
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка доступа к элементу fieldShowFor: <%v>", err))
			return nil
		}
		name := v.ViewBuffer()
		name = strings.ReplaceAll(name, "\n", "") // удаление символа

		// Удаление файла.
		if err := c.conf.Container.RemoveFileFromContainer(name, c.secret.secretKey); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция RemoveFileFromContainer, вернула ошибку: <%v>", err))
			return nil
		}
		c.flag.delFileSUCCESS = true

	default:
	}

	return nil
}

// Извлечение.
func (c *handlerUI) doExtract(gui *gocui.Gui, v *gocui.View) error {

	switch c.view.activeView {
	case viewBinaryData: // Окно работы с файлами

		c.flag.extractFilePassed = true
		c.flag.extractFileSUCCESS = false

		// Чтение буфера.
		v, err := gui.View("fieldShowFor")
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка доступа к элементу fieldShowFor: <%v>", err))
			return nil
		}
		name := v.ViewBuffer()
		name = strings.ReplaceAll(name, "\n", "") // удаления символа

		// Получение файла из хранилища.
		contentFile, err := c.conf.Container.GetFileFromContainer(name, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка получения файла: <%s> их контейнера: <%v>", name, err))
			return nil
		}

		// Сохранение файла.
		path := c.typed.dataPathTrg + name

		err = os.WriteFile(path, contentFile, 0644)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка получения файла: <%s> из контейнера: <%v>", c.typed.dataPathTrg, err))
			return nil
		}
		c.flag.extractFileSUCCESS = true
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

	// Создание экземпляра контейнера.
	//
	// Создаётся экземпляр в этом месте, т.к. ключ шифрования формируется после запроса дополнительного ключа.
	inst := container.New(containerName, c.secret.secretKey)
	c.conf.Container = inst

	c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: стартовая обработка контейнера <%s> пройдена", containerName))

	// Логика обработчика
	//
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
	c.view.currentFocus = "fieldAddFor" // Установка фокуса на элемент окна.

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

	c.conf.PtrLoggerFile.Write("Debug: выполнен вход в окно typeText")

	c.flag.readTextPassed = false // Сброс признака.
	c.flag.addTextPassed = false
	c.flag.delTextPassed = false
	c.flag.readTextSUCCESS = false
	c.index.text = 0 // Сброс индекса навигации по массиву логин/пароль.

	// Логика.
	//
	c.view.activeView = "" // Сброс признака активного окна

	// Удаляем все зависимые виды
	if err := deleteViews(g); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}

	// Сброс состояния
	layoutInitialized = false
	c.view.currentFocus = "fieldAddFor" // Установка фокуса на элемент окна.

	// Создание контейнера запроса ввода дополнительного секретного ключа.
	view, err := g.SetView(viewTextData, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	view.Title = "Текст"
	view.Wrap = true
	view.Clear()

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
	// --- Отображение разделов ---
	//

	// Просмотр.
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
	if v, err := g.SetView("fieldShowText", 33, 4, inputWidth+48, inputHeight+1+3); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Title = "Текст"
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault

		v.Write([]byte("....."))
		c.setFocusStyle(v, "fieldShowText")
	}

	// Добавление
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
	if v, err := g.SetView("fieldAddText", 33, 12, inputWidth+48, inputHeight+1+11); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Title = "Текст"
		v.Editable = true
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorBlack
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack

		v.Write([]byte("....."))
		c.setFocusStyle(v, "fieldAddText")
	}

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

	// Нижний ряд
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

	// Получение сохранённых значений текста.
	c.flag.readTextPassed = true // Установка признака, что был запущен процесс получения значений текста.

	_, err = showTextWorkDB(c)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция showTextWorkDB, вернула ошибку: <%v>", err))
	} else {
		c.conf.PtrLoggerFile.Write("Debug: текстовые данные успешно прочитаны")
		c.flag.readTextSUCCESS = true
	}

	// Установка фокуса.
	if _, err := g.SetCurrentView(c.view.currentFocus); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	layoutInitialized = true
	c.view.activeView = viewTextData // Установка признака активного окна

	return nil
}

// Окно для взаимодействия с логин/пароль.
func (c *handlerUI) showBinary(g *gocui.Gui, _ *gocui.View) error {

	c.view.activeView = "" // Сброс
	c.index.file = 0

	c.flag.addFilePassed = false
	c.flag.addFileSUCCESS = false

	c.flag.readFilePassed = false
	c.flag.readFileSUCCESS = false

	c.flag.delFilePassed = false
	c.flag.delFileSUCCESS = false

	c.flag.extractFilePassed = false
	c.flag.extractFileSUCCESS = false

	// Удаляем все зависимые виды.
	if err := deleteViews(g); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}

	// Сброс состояния.
	layoutInitialized = false
	c.view.currentFocus = "fieldPathSource" // Установка фокуса на поле ввода.

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
	// --- Вывод надписей ---
	//

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 3))
	fmt.Fprintf(view, "%sФайл в хранилище:\n", strings.Repeat(" ", 2))

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 3))
	fmt.Fprintf(view, "%sОткуда:\n", strings.Repeat(" ", 2))

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 3))
	fmt.Fprintf(view, "%sКуда:\n", strings.Repeat(" ", 2))

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 2))
	fmt.Fprintf(view, "%sДобавление файла: - указать путь к файлу <Откуда> и выполнить Crl+F.\n", strings.Repeat(" ", 2))

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 2))
	fmt.Fprintf(view, "%sИзвлечение файла: - указать путь к директории <Куда>.%sПри изменении данных, выполнить Ctrl+U\n", strings.Repeat(" ", 2), strings.Repeat(" ", 30))
	fmt.Fprintf(view, "%s                  - используя Crl+E и Ctrl+G, выбрать файл в хранилище.\n", strings.Repeat(" ", 2))
	fmt.Fprintf(view, "%s                  - выполнить извлечение Ctrl+K.\n", strings.Repeat(" ", 2))

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 2))
	fmt.Fprintf(view, "%sУдаление файла:   - выбрать файл через Ctrl+E, Ctrl+G и нажать Crl+J.\n", strings.Repeat(" ", 2))

	//
	// --- Индикаторы ---
	//

	// Результат чтения данных.
	indicatorY := 2
	indicatorX := 105
	vRead, err := g.SetView("indicatorReadStatus", indicatorX, indicatorY, indicatorX+20, indicatorY+2)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
	}
	vRead.Frame = false
	vRead.BgColor = gocui.ColorDefault
	vRead.FgColor = gocui.ColorDefault

	//
	// --- Поля вывода ---
	//

	// Отображение имени файла
	if v, err := g.SetView("fieldShowFor", 22, 2, inputWidth+20, inputHeight+1+1); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault

		v.Write([]byte("....."))
		c.setFocusStyle(v, "fieldShowFor")
	}

	//
	// --- Поля ввода ---
	//

	// Полный путь к файлу.
	if v, err := g.SetView("fieldPathSource", 22, 6, inputWidth+48, inputHeight+1+5); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = true
		v.Editable = true
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorBlack
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack

		v.Write([]byte("....."))
		c.setFocusStyle(v, "fieldPathSource")
	}
	// Директория назначения.
	if v, err := g.SetView("fieldPathTarget", 22, 10, inputWidth+48, inputHeight+1+9); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = true
		v.Editable = true
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorBlack
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack

		v.Write([]byte("....."))
		c.setFocusStyle(v, "fieldPathTarget")
	}

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
	if v, err := g.SetView("Extraction", 104, 23, inputWidth+48, inputHeight+1+22); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+K - Извлечение"))
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
		v.Write([]byte("Enter - Фиксация"))
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

	// Получение списка названий файлов.
	c.flag.readFilePassed = true // Установка признака, что был запущен процесс получения значений текста.

	c.data.files, err = c.conf.Container.ListFilesInContainer(c.secret.secretKey)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция ListFilesInContainer, вернула ошибку: <%v>", err))
	} else {
		c.conf.PtrLoggerFile.Write("Debug: имена файлов в контейнере, успешно прочитаны")
		c.flag.readFileSUCCESS = true
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

// Окно для взаимодействия с банковской картой.
func (c *handlerUI) showBankCard(g *gocui.Gui, _ *gocui.View) error {

	c.conf.PtrLoggerFile.Write("Debug: выполнен вход в окно typeBankCard")

	c.flag.readBankCardPassed = false // Сброс признака.
	c.flag.addBankCardPassed = false
	c.flag.delBankCardPassed = false
	c.index.bankCard = 0 // Сброс индекса навигации по массиву логин/пароль.
	c.flag.readBankCardSUCCESS = false

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
	c.view.currentFocus = "fieldAddFor" // Установка фокуса на элемент окна.

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
	// --- Отображение разделов ---
	//

	// Просмотр.
	fmt.Fprintf(view, "%s", strings.Repeat("\n", 1))
	fmt.Fprintf(view, "%sПросмотр.\n", strings.Repeat(" ", 56))

	if v, err := g.SetView("fieldShowFor", 1, 2, inputWidth-48, inputHeight+1+1); err != nil {
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
	if v, err := g.SetView("fieldShowOwner", 33, 2, inputWidth-4, inputHeight+1+1); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Title = "Владелец"
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault

		v.Write([]byte("....."))
		c.setFocusStyle(v, "fieldShowOwner")
	}
	if v, err := g.SetView("fieldShowNumber", 33, 5, inputWidth-4, inputHeight+1+4); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Title = "Номер"
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault

		v.Write([]byte("....."))
		c.setFocusStyle(v, "fieldShowNumber")
	}
	if v, err := g.SetView("fieldShowValid", 33, 8, inputWidth-30, inputHeight+1+7); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Title = "Дата"
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault

		v.Write([]byte("....."))
		c.setFocusStyle(v, "fieldShowValid")
	}
	if v, err := g.SetView("fieldShowCode", 59, 8, inputWidth-4, inputHeight+1+7); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Title = "Код"
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault

		v.Write([]byte("....."))
		c.setFocusStyle(v, "fieldShowCode")
	}

	// Добавление
	fmt.Fprintf(view, "%s", strings.Repeat("\n", 10))
	fmt.Fprintf(view, "%sДобавление. %sПри изменении данных, выполните Crl+U\n", strings.Repeat(" ", 55), strings.Repeat(" ", 15))

	if v, err := g.SetView("fieldAddFor", 1, 13, inputWidth-48, inputHeight+1+12); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Title = "Для"
		v.Editable = true
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault

		v.Write([]byte("....."))
		c.setFocusStyle(v, "fieldAddFor")
	}
	if v, err := g.SetView("fieldAddOwner", 33, 13, inputWidth-4, inputHeight+1+12); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Title = "Владелец"
		v.Editable = true
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault

		v.Write([]byte("....."))
		c.setFocusStyle(v, "fieldAddOwner")
	}
	if v, err := g.SetView("fieldAddNumber", 33, 16, inputWidth-4, inputHeight+1+15); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Title = "Номер"
		v.Editable = true
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault

		v.Write([]byte("....."))
		c.setFocusStyle(v, "fieldAddNumber")
	}
	if v, err := g.SetView("fieldAddValid", 33, 19, inputWidth-30, inputHeight+1+18); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Title = "Дата"
		v.Editable = true
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault

		v.Write([]byte("....."))
		c.setFocusStyle(v, "fieldAddValid")
	}
	if v, err := g.SetView("fieldAddCode", 59, 19, inputWidth-4, inputHeight+1+18); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Title = "Код"
		v.Editable = true
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault

		v.Write([]byte("....."))
		c.setFocusStyle(v, "fieldAddCode")
	}

	//
	// --- Индикаторы ---
	//

	// Результат чтения данных.

	indicatorY := inputHeight*3 - 1
	indicatorX := 80
	vRead, err := g.SetView("indicatorReadStatus", indicatorX, indicatorY, indicatorX+20, indicatorY+2)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
	}
	vRead.Frame = false
	vRead.BgColor = gocui.ColorDefault
	vRead.FgColor = gocui.ColorDefault

	// Результат добавления данных.
	indicatorY = inputHeight*3 + 10
	indicatorX = 80
	vAdd, err := g.SetView("indicatorAddSuccess", indicatorX, indicatorY, indicatorX+20, indicatorY+2)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
	}
	vAdd.Frame = false
	vAdd.BgColor = gocui.ColorDefault
	vAdd.FgColor = gocui.ColorDefault

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

	// Получение сохранённых значений банковских карт.
	c.flag.readBankCardPassed = true // Установка признака, что был запущен процесс получения значений банковских карт.

	_, err = showBankCardWorkDB(c)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция showBankCardWorkDB, вернула ошибку: <%v>", err))
	} else {
		c.conf.PtrLoggerFile.Write("Debug: данные банковских карт успешно прочитаны")
		c.flag.readBankCardSUCCESS = true
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
