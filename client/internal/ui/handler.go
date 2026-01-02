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
type status struct {
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
	restore                  int  // Статус процесса воостановления из резервной копии.
	backUp                   int  // Статус процесса создания резервной копии.
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

// Мьютексы.
type mutex struct {
	restoreBackup sync.Mutex // для резервного копирования и восстановления.
	processTxRx   sync.Mutex // для данных процесса Tx Rx файлов.
}

// Отправка-приём файлов.
type txrx struct {
	percentTxRx float32 // Процент выполнения процесса передачи файлов.
	totalSizeB  int64   // Передаваемый размер (Байт).
	passedB     int64   // обработано данных (Байт).
}

// Общий тип для CLI UI.
type handlerUI struct {
	conf   *udt.Configuration // конфигурация сервиса.
	typed  typeData           // введённые пользователем данные.
	status status             // признаки сервиса.
	view   screens            // взаимодействие с окнами.
	secret encrKey            // секретность.
	data   data               // данные.
	index  indexes            // индексы для обхода массивов.
	mutex  mutex              // мьютексы.
	txrx   txrx               // данные по Tx-Rx файлов.
}

var inst *handlerUI

// Конструктор.
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
			},
			index: indexes{
				loginPassword: 0,
			},
			mutex: mutex{
				restoreBackup: sync.Mutex{},
				processTxRx:   sync.Mutex{},
			},
			txrx: txrx{},
		}
	})
	return inst
}

// Главное окно.
func layout(g *gocui.Gui) error {

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

	// Запрет активности при активности процессов передачи файлов.
	if c.status.backUp == stageActive || c.status.restore == stageActive {
		return nil
	}

	c.view.activeView = "" // Сброс признака активного окна

	// Удаление видов.
	if err := deleteViews(g); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция deleteViews вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews вернула ошибку: <%w>", err)
	}

	// Сброс флагов.
	layoutInitialized = false
	c.view.currentFocus = ""

	// Отображение главного меню.
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

	// Ограничение вызова окна.
	if c.view.activeView == viewAutentification ||
		c.view.activeView == viewBankCardData ||
		c.view.activeView == viewBinaryData ||
		c.view.activeView == viewLoginPasswordData ||
		c.view.activeView == viewRegistration ||
		c.view.activeView == viewRequestSecretKey ||
		c.view.activeView == viewSelectType ||
		c.view.activeView == viewSettings ||
		c.view.activeView == viewTextData {
		return nil
	}

	// Логика
	//

	// Сбросы.
	c.view.activeView = ""
	c.typed.login = ""
	c.typed.password1 = ""
	c.typed.password2 = ""
	c.status.addUserPassed = false
	c.status.addUserSUCCESS = false

	// Удаляем все зависимые виды (включая поля ввода).
	if err := deleteViews(g); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}

	// Инициализация.
	layoutInitialized = false
	c.view.currentFocus = "Login" // Установка фокуса

	// Создание контейнера регистрации.
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

	// Ограничение вызова окна.
	if c.view.activeView == viewAutentification ||
		c.view.activeView == viewBankCardData ||
		c.view.activeView == viewBinaryData ||
		c.view.activeView == viewLoginPasswordData ||
		c.view.activeView == viewRegistration ||
		c.view.activeView == viewRequestSecretKey ||
		c.view.activeView == viewSelectType ||
		c.view.activeView == viewSettings ||
		c.view.activeView == viewTextData {
		return nil
	}

	// Логика
	//

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

	// Ограничение вызова окна.
	if c.view.activeView == viewAutentification ||
		c.view.activeView == viewBankCardData ||
		c.view.activeView == viewBinaryData ||
		c.view.activeView == viewLoginPasswordData ||
		c.view.activeView == viewRegistration ||
		c.view.activeView == viewRequestSecretKey ||
		c.view.activeView == viewSelectType ||
		c.view.activeView == viewSettings ||
		c.view.activeView == viewTextData {
		return nil
	}

	// Логика
	//
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

	// Запрет активности при активности процессов передачи файлов.
	if c.status.backUp == stageActive || c.status.restore == stageActive {
		return nil
	}

	switch c.view.activeView {

	// Окно регистрации.
	case viewRegistration:
		if err := enterViewRegistration(v, c); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция enterViewRegistration, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция enterViewRegistration, вернула ошибку: <%v>", err)
		}

	// Окно аутентификации.
	case viewAutentification:
		if err := enterViewAutentification(v, c); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция enterViewAutentification, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция enterViewAutentification, вернула ошибку: <%v>", err)
		}

	// Окно настроек.
	case viewSettings:
		if err := enterViewSettings(v, g, c); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция enterViewSettings, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция enterViewSettings, вернула ошибку: <%v>", err)
		}

	// Окно с запросом дополнительного ключа шифрования.
	case viewRequestSecretKey:
		if err := enterViewRequestSecretKey(v, g, c); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция enterViewRequestSecretKey, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция enterViewRequestSecretKey, вернула ошибку: <%v>", err)
		}

	// Окно с выбором типа данных.
	case viewSelectType:
		if err := enterViewSelectType(v, g, c); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция enterViewSelectType, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция enterViewSelectType, вернула ошибку: <%v>", err)
		}

	// Окно взаимодействия с логин/пароль.
	case viewLoginPasswordData:
		if err := enterViewLoginPasswordData(v, c); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция enterViewLoginPasswordData, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция enterViewLoginPasswordData, вернула ошибку: <%v>", err)
		}

	// Окно взаимодействия с текстом.
	case viewTextData:
		if err := enterViewTextData(v, c); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция enterViewTextData, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция enterViewTextData, вернула ошибку: <%v>", err)
		}

	// Окно взаимодействия с банковскими картами.
	case viewBankCardData:
		if err := enterViewBankCardData(v, c); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция enterViewBankCardData, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция enterViewBankCardData, вернула ошибку: <%v>", err)
		}

	// Окно взаимодействия с файлами.
	case viewBinaryData:
		if err := enterViewBinaryData(v, c); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция enterViewBinaryData, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция enterViewBinaryData, вернула ошибку: <%v>", err)
		}

	default:
	}

	return nil
}

// Проверка совпадения паролей при регистрации. Возвращается ошибка.
func (c *handlerUI) indicators(g *gocui.Gui) error {

	switch c.view.activeView {

	// Окно регистрации.
	case viewRegistration:
		if err := indicatorViewRegistration(g, c); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция indicatorViewRegistration, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция indicatorViewRegistration, вернула ошибку: <%w>", err)
		}

	// Окно настроек.
	case viewSettings:
		if err := indicatorViewSettings(g, c); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция indicatorViewSettings, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция indicatorViewSettings, вернула ошибку: <%v>", err)
		}

	// Окно логин/пароль
	case viewLoginPasswordData:
		if err := indicatorViewLoginPasswordData(g, c); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция indicatorViewLoginPasswordData, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция indicatorViewLoginPasswordData, вернула ошибку: <%v>", err)
		}

	// Окно текста.
	case viewTextData:
		if err := indicatorViewTextData(g, c); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция indicatorViewTextData, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция indicatorViewTextData, вернула ошибку: <%v>", err)
		}

	// Окно банковских карт.
	case viewBankCardData:
		if err := indicatorViewBankCardData(g, c); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция indicatorViewBankCardData, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция indicatorViewBankCardData, вернула ошибку: <%v>", err)
		}

	// Окно файлов.
	case viewBinaryData:
		if err := indicatorViewBinaryData(g, c); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция indicatorViewBankCardData, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция indicatorViewBankCardData, вернула ошибку: <%v>", err)
		}

	// Окно выбора типов.
	case viewSelectType:
		if err := indicatorViewSelectType(g, c); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция indicatorViewSelectType, вернула ошибку: <%v>", err))
			return fmt.Errorf("Error: Функция indicatorViewSelectType, вернула ошибку: <%v>", err)
		}

	default:
	}
	return nil
}

// Проверка связи с сервером.
func (c *handlerUI) testConnect(g *gocui.Gui, _ *gocui.View) error {

	// Запуск проверки связи с сервером.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Выполнение проверки связи.
	ok, err := pingContext(ctx, c)
	if err != nil {
		c.status.checkConnectStatus = false
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция pingContext, вернула ошибку: <%v>", err))
		return nil
	}

	// Результат.
	c.status.checkConnectStatus = ok
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

	c.status.addUserSUCCESS = false // сброс признака успешности регистрации пользователя.
	c.status.addUserPassed = false  // сброс признака, что процедура регистрации быд запущена.
	c.status.addUserRegBusy = false // сброс признака, что в системе уже есть зарегистрированный пользователь.

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
	busy, err := c.conf.DataBase.UserExistContext(ctx)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция UserExistContext, вернуля ошибку: <%v>", err))
		return nil
	}

	if busy {
		c.status.addUserRegBusy = true // установка признака, что в системе уже есть зарегистрированный пользоатель.
		return nil
	}

	// Добавление пользователя в БД.
	if err := c.conf.DataBase.AddUserContext(ctx, userName, userPwd1); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция AddUserContext, вернуля ошибку: <%v>", err))
		return nil
	}

	c.status.addUserPassed = true  // установка признака, что процедура регистрации была запущена.
	c.status.addUserSUCCESS = true // установка признака, что пользователь зарегистрировался в системе.
	c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: выполнена регистрация пользователя с именем: <%s>", userName))

	return nil
}

// Запуск процесса аутентификации пользователя.
func (c *handlerUI) doAuthenticationUser(gui *gocui.Gui, v *gocui.View) error {

	// Если окно аутентификации.
	if c.view.activeView == viewAutentification {
		userName := c.typed.login
		userPwd1 := c.typed.password1

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Выполнение запроса.
		ok, err := c.conf.DataBase.AuthenticateUserContext(ctx, userName, userPwd1)
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
	}

	return nil
}

// Запуск процесса сохранения данных.
func (c *handlerUI) doStore(gui *gocui.Gui, v *gocui.View) error {

	c.conf.PtrLoggerFile.Write("Debug: запущена функция doStoreLoginPasswordDB")

	switch c.view.activeView {
	case viewLoginPasswordData: // Если окно - логин/пароль

		c.status.addLoginPaaswordPassed = true
		c.status.addLoginPaaswordSUCCESS = false

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
		if err := c.conf.DataBase.AddDataLoginPasswordContext(ctx, encrFor, encrLogin, encrPassword, encrCreatedAt); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка добавления пары логин/пароль в БД: <%v>", err))
			return nil
		}

		c.conf.PtrLoggerFile.Write("Debug: пара логин/пароль добавлена в БД")
		c.status.addLoginPaaswordSUCCESS = true
		return nil

	case viewTextData: // если окно - текст.

		c.conf.PtrLoggerFile.Write(fmt.Sprintf("--- Debug: Добавляются данные For:<%s> Text:<%s>", c.typed.dataFor, c.typed.dataText)) //====================

		c.status.addTextPassed = true
		c.status.addTextSUCCESS = false

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
		if err := c.conf.DataBase.AddDataTextContext(ctx, encrFor, encrText, encrCreatedAt); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка добавления текста в БД: <%v>", err))
			return nil
		}

		c.conf.PtrLoggerFile.Write("Debug: текст добавлен в БД")
		c.status.addTextSUCCESS = true
		return nil

	case viewBankCardData: // если окно - банковские карты.

		c.status.addBankCardPassed = true
		c.status.addBankCardSUCCESS = false

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
		if err := c.conf.DataBase.AddDataBankCardContext(ctx, encrFor, encrOwner, encrNumb, encrValid, encrCode, encrCreatedAt); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка добавления карты в БД: <%v>", err))
			return nil
		}

		c.conf.PtrLoggerFile.Write("Debug: карта добавлена в БД")
		c.status.addBankCardSUCCESS = true
		return nil

	case viewBinaryData: // Окно для работы с файлами.

		c.status.addFilePassed = true
		c.status.addFileSUCCESS = false

		if err := c.conf.Container.AddFileToContainer(c.typed.dataPathSrc, c.secret.secretKey); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка при добавлении файла <%s>, в контейнер", c.typed.dataPathSrc))
			return nil
		}
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: файл <%s>, добавлен в контейнер", c.typed.dataPathSrc))
		c.status.addFileSUCCESS = true

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

		c.status.delLoginPaaswordPassed = true
		c.status.delLoginPaaswordSUCCESS = false

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

		if err := c.conf.DataBase.DelDataLoginPasswordContext(ctx, textEl); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция DelDataLoginPasswordContext, вернула ошибку: <%v>", err))
			return nil
		}

		c.status.delLoginPaaswordSUCCESS = true
		c.conf.PtrLoggerFile.Write(("Debug: данные логин/пароль, успешно удалены"))

	case viewTextData: // Взаимодействие с текстом

		c.status.delTextPassed = true
		c.status.delTextSUCCESS = false

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

		if err := c.conf.DataBase.DelTextContext(ctx, textEl); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция DelTextContext, вернула ошибку: <%v>", err))
			return nil
		}

		c.status.delTextSUCCESS = true
		c.conf.PtrLoggerFile.Write(("Debug: данные текста, успешно удалены"))

	case viewBankCardData: // Взаимодействие с банковскими картами

		c.status.delBankCardPassed = true
		c.status.delBankCardSUCCESS = false

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

		if err := c.conf.DataBase.DelBankCardContext(ctx, textEl); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция DelBankCardContext, вернула ошибку: <%v>", err))
			return nil
		}

		c.status.delBankCardSUCCESS = true
		c.conf.PtrLoggerFile.Write(("Debug: данные карты, успешно удалены"))

	case viewBinaryData: // Окно работы с файлами.

		c.status.delFilePassed = true
		c.status.delFileSUCCESS = false

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
		c.status.delFileSUCCESS = true

	default:
	}

	return nil
}

// Извлечение.
func (c *handlerUI) doExtract(gui *gocui.Gui, v *gocui.View) error {

	switch c.view.activeView {
	case viewBinaryData: // Окно работы с файлами

		c.status.extractFilePassed = true
		c.status.extractFileSUCCESS = false

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
		c.status.extractFileSUCCESS = true
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

	// Ограничение.
	if c.view.activeView != viewLoginPasswordData &&
		c.view.activeView != viewTextData &&
		c.view.activeView != viewBankCardData &&
		c.view.activeView != viewBinaryData &&
		c.view.activeView != viewRequestSecretKey {
		return nil
	}

	// Сброс статусных признаков.
	c.status.backUp = stageNotActive
	c.status.restore = stageNotActive

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
	// --- Индикаторы ---
	//

	// процент выполнения.
	if v, err := g.SetView("indicatorPercent", 50, 20, inputWidth+1, inputHeight+1+19); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = false
		v.Frame = false
		v.BgColor = gocui.ColorDefault
		v.SelBgColor = gocui.ColorDefault
		v.SelFgColor = gocui.ColorDefault
	}

	//
	// --- Пояснение к действию ---
	//
	fmt.Fprintf(view, "%s", strings.Repeat("\n", 15))
	fmt.Fprintf(view, "%sВыберите нужный раздел через Tab и нажмите Enter.\n", strings.Repeat(" ", 40))

	//
	// --- Нижняя часть экрана ---
	//

	// Верхний ряд.
	if v, err := g.SetView("Backup", 1, 23, inputWidth-55, inputHeight+1+22); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+O - ---> сервер"))
	}

	// Нижний ряд.
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
	if v, err := g.SetView("Restore", 104, 26, inputWidth+48, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+P - <--- сервер"))
	}

	// Установка фокуса.
	if _, err := g.SetCurrentView(c.view.currentFocus); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	layoutInitialized = true
	c.view.activeView = viewSelectType // Установка признака активного окна

	return nil
}

// Окно для взаимодействия с логин/пароль.
func (c *handlerUI) showLoginPassword(g *gocui.Gui, _ *gocui.View) error {

	// ограничение.
	if c.view.activeView != viewSelectType {
		return nil
	}

	c.conf.PtrLoggerFile.Write("Debug: выполнен вход в окно typeLoginPassword")

	c.status.readLoginPaaswordPassed = false // Сброс признака.
	c.status.addLoginPaaswordPassed = false
	c.status.delLoginPaaswordPassed = false
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
	c.status.readLoginPaaswordPassed = true // Установка признака, что был запущен процесс получения значений логин/пароль.

	_, err = showLoginPasswordWorkDB(c)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция funcshowLoginPasswordWorkDB, вернула ошибку: <%v>", err))
		c.status.readLoginPaaswordSUCCESS = false
	} else {
		c.conf.PtrLoggerFile.Write("Debug: данные логин/пароль успешно прочитаны")
		c.status.readLoginPaaswordSUCCESS = true
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

	// ограничение.
	if c.view.activeView != viewSelectType {
		return nil
	}

	c.conf.PtrLoggerFile.Write("Debug: выполнен вход в окно typeText")

	c.status.readTextPassed = false // Сброс признака.
	c.status.addTextPassed = false
	c.status.delTextPassed = false
	c.status.readTextSUCCESS = false
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
	c.status.readTextPassed = true // Установка признака, что был запущен процесс получения значений текста.

	_, err = showTextWorkDB(c)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция showTextWorkDB, вернула ошибку: <%v>", err))
	} else {
		c.conf.PtrLoggerFile.Write("Debug: текстовые данные успешно прочитаны")
		c.status.readTextSUCCESS = true
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

	// ограничение.
	if c.view.activeView != viewSelectType {
		return nil
	}

	c.view.activeView = "" // Сброс
	c.index.file = 0

	c.status.addFilePassed = false
	c.status.addFileSUCCESS = false

	c.status.readFilePassed = false
	c.status.readFileSUCCESS = false

	c.status.delFilePassed = false
	c.status.delFileSUCCESS = false

	c.status.extractFilePassed = false
	c.status.extractFileSUCCESS = false

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
	fmt.Fprintf(view, "%sОткуда (файл):\n", strings.Repeat(" ", 2))

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 3))
	fmt.Fprintf(view, "%sКуда (директория):\n", strings.Repeat(" ", 2))

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 5))
	fmt.Fprintf(view, "%sПри внесении изменений, выполните Crl+U.\n", strings.Repeat(" ", 45))

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
	c.status.readFilePassed = true // Установка признака, что был запущен процесс получения значений текста.

	c.data.files, err = c.conf.Container.ListFilesInContainer(c.secret.secretKey)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция ListFilesInContainer, вернула ошибку: <%v>", err))
	} else {
		c.conf.PtrLoggerFile.Write("Debug: имена файлов в контейнере, успешно прочитаны")
		c.status.readFileSUCCESS = true
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

	// ограничение.
	if c.view.activeView != viewSelectType {
		return nil
	}

	c.conf.PtrLoggerFile.Write("Debug: выполнен вход в окно typeBankCard")

	c.status.readBankCardPassed = false // Сброс признака.
	c.status.addBankCardPassed = false
	c.status.delBankCardPassed = false
	c.index.bankCard = 0 // Сброс индекса навигации по массиву логин/пароль.
	c.status.readBankCardSUCCESS = false

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
	c.status.readBankCardPassed = true // Установка признака, что был запущен процесс получения значений банковских карт.

	_, err = showBankCardWorkDB(c)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция showBankCardWorkDB, вернула ошибку: <%v>", err))
	} else {
		c.conf.PtrLoggerFile.Write("Debug: данные банковских карт успешно прочитаны")
		c.status.readBankCardSUCCESS = true
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

// Передача данных клиента, на сервер.
func (c *handlerUI) doBackup(gui *gocui.Gui, v *gocui.View) (err error) {

	// Запрет активности при активности процессов передачи файлов.
	if c.status.backUp == stageActive || c.status.restore == stageActive {
		return nil
	}

	// Логика работает только из окна выбора типа.
	if c.view.activeView == viewSelectType {

		c.mutex.restoreBackup.Lock()
		defer c.mutex.restoreBackup.Unlock()

		c.status.backUp = stageActive // Установка признака, что запущен процесс передачи файлов на сервер.

		// Логика процесса.
		//

		fileNameDB := "manager.db"
		fileNameContainer := "container.data"

		// Определение общего размера файлов.
		c.txrx.totalSizeB, err = totalFileSize(fileNameDB, fileNameContainer)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция totalFileSize, вернула ошибку: <%v>", err))
			c.status.backUp = stageFault
			return nil
		}

		// БД.
		if err := backUpDB(c, fileNameDB); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция backUpDB, вернула ошибку: <%v>", err))
			c.status.backUp = stageFault
			return nil
		}
		c.conf.PtrLoggerFile.Write("Info: Резервное копирование БД, выполнено")

		// Контейнер.
		if err := backUpContainer(c, fileNameContainer); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция backUpContainer, вернула ошибку: <%v>", err))
			c.status.backUp = stageFault
			return nil
		}

		c.conf.PtrLoggerFile.Write("Info: Резервное копирование контейнера, выполнено")
		c.status.backUp = stageOk

	}
	return nil
}

// Получение данных клиента, от сервер.
func (c *handlerUI) doRestore(gui *gocui.Gui, v *gocui.View) error {

	// Запрет активности при активности процессов передачи файлов.
	if c.status.backUp == stageActive || c.status.restore == stageActive {
		return nil
	}

	// Логика работает только из окна выбора типа.
	if c.view.activeView == viewSelectType {

		c.mutex.restoreBackup.Lock()
		defer c.mutex.restoreBackup.Unlock()

		c.status.restore = stageActive // Установка признака, что запущен процесс приёма файлов от сервера.

		if err := restoreDB(c); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция restoreDB, вернула ошибку: <%v>", err))
			c.status.restore = stageFault // Установка признака ошибки процесса получения резервной копии.
			return nil
		}
		c.conf.PtrLoggerFile.Write("Info: Восстановление БД, выполнено")

		if err := restoreContainer(c); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция restoreContainer, вернула ошибку: <%v>", err))
			c.status.restore = stageFault // Установка признака ошибки процесса получения резервной копии.
			return nil
		}

		c.conf.PtrLoggerFile.Write("Info: Восстановление контейнера, выполнено")
		c.status.restore = stageOk // Установка признака, что восстановление выполнено.
	}
	return nil
}
