package ui

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Part001-R/YaPr-GP-2/client/internal/adapters/container"
	"github.com/Part001-R/YaPr-GP-2/client/internal/adapters/server"
	"github.com/Part001-R/YaPr-GP-2/client/internal/service/udt"
	service "github.com/Part001-R/YaPr-GP-2/client/internal/service/udt"
	"github.com/Part001-R/YaPr-GP-2/client/internal/utils/flags"
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
	checkConnectStatus           bool // Результат процедуры проверки связи с сервером.
	checkConnectPassed           bool // Признак, что проверка связи была запущена.
	addUserSUCCESS               bool // Признак успешного добавления пользователя.
	addUserPassed                bool // Признак, что была запущена процедура регистрации пользователя.
	addUserRegBusy               bool // Признак, чот уже есть заргистрированный пользователь
	addLoginPaaswordSUCCESS      bool // Признак успешного добавления пары логин/пароль.
	addLoginPaaswordPassed       bool // Признак, что выполнена процедура добавления пары логин/пароль.
	readLoginPaaswordSUCCESS     bool // Признак, успешного получения данных логин/пароль.
	readLoginPaaswordPassed      bool // Признак, что процедура чтения логин/пароль, пройдена.
	readNameLoginPaaswordSUCCESS bool // Признак, успешного получения имён логин/пароль.
	readNameLoginPaaswordPassed  bool // Признак, что процедура получения имён логин/пароль, пройдена.
	readNameTextSUCCESS          bool // Признак, успешного получения имён текста.
	readNameTextPassed           bool // Признак, что процедура получения имён текста, пройдена.
	delLoginPaaswordSUCCESS      bool // Признак, успешного удаления данных логин/пароль.
	delLoginPaaswordPassed       bool // Признак, что процедура удаления логин/пароль, пройдена.
	addTextSUCCESS               bool // Признак успешного добавления текста.
	addTextPassed                bool // Признак, что выполнена процедура добавления текста.
	readTextSUCCESS              bool // Признак, успешного получения данных текста.
	readTextPassed               bool // Признак, что процедура получения текста, пройдена.
	delTextSUCCESS               bool // Признак, успешного удаления данных текста.
	delTextPassed                bool // Признак, что процедура удаления текста, пройдена.
	addBankCardSUCCESS           bool // Признак успешного добавления карты.
	addBankCardPassed            bool // Признак, что выполнена процедура добавления карты.
	readBankCardSUCCESS          bool // Признак, успешного получения данных карт.
	readBankCardPassed           bool // Признак, что процедура получения данных карт, пройдена.
	readNameBankCardSUCCESS      bool // Признак, успешного получения данных карт.
	readNameBankCardPassed       bool // Признак, что процедура получения данных карт, пройдена.
	readNameFileSUCCESS          bool // Признак, успешного получения данных файлов.
	readNameFilePassed           bool // Признак, что процедура получения данных файлов, пройдена.
	delBankCardSUCCESS           bool // Признак, успешного удаления данных карты.
	delBankCardPassed            bool // Признак, что процедура удаления карты, пройдена.
	addFileSUCCESS               bool // Признак успешного добавления файла.
	addFilePassed                bool // Признак, что выполнена процедура добавления файла.
	readFileSUCCESS              bool // Признак, успешного получения данных файла.
	readFilePassed               bool // Признак, что процедура получения данных файла, пройдена.
	delFileSUCCESS               bool // Признак, успешного удаления файла.
	delFilePassed                bool // Признак, что процедура удаления файла, пройдена.
	extractFileSUCCESS           bool // Признак, успешного извлечения файла.
	extractFilePassed            bool // Признак, что процедура извлечения файла, пройдена.
	restore                      int  // Статус процесса воостановления из резервной копии.
	backUp                       int  // Статус процесса создания резервной копии.
	pushContainer                int  // Статус процесса передачи в контейнер.
	popContainer                 int  // Статус процесса извлечения из контейнера.
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
	namesLoginPassword   []string        // имена записей логин/пароль
	namesText            []string        // имена записей текст
	namesBankCard        []string        // имена записей банковские карты
	namesFile            []string        // имена файлов
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
	restoreBackup       sync.Mutex // для резервного копирования и восстановления.
	processTxRx         sync.Mutex // для данных процесса Tx Rx файлов.
	statusBackUp        sync.Mutex // для статуса процесса передачи.
	statusRestore       sync.Mutex // для статуса процесса приёма.
	statusPushContainer sync.Mutex // для статуса процесса передачи в контейнер.
	statusPopContainer  sync.Mutex // для статуса процесса извлечения из контейнера.
}

// Отправка-приём файлов.
type txrx struct {
	percentTxRx float32 // Процент выполнения процесса передачи файлов.
	totalSizeKB int64   // Передаваемый размер (Байт).
	passedKB    int64   // обработано данных (Байт).
}

// Общий тип для CLI UI.
type handlerUI struct {
	conf       *udt.Configuration // конфигурация сервиса.
	typed      typeData           // введённые пользователем данные.
	status     status             // признаки сервиса.
	view       screens            // взаимодействие с окнами.
	secret     encrKey            // секретность.
	data       data               // данные.
	index      indexes            // индексы для обхода массивов.
	mutex      mutex              // мьютексы.
	txrx       txrx               // данные по Tx-Rx файлов.
	clientName string             // имя клиента.
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
			},
			txrx: txrx{},
		}
	})
	return inst
}

// Главное окно в режиме - локальный. Возвращается ошибка.
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

// Главное окно.
func (c *handlerUI) showMain(g *gocui.Gui, v *gocui.View) error {

	c.conf.PtrLoggerFile.Write("Info: Нажата комбинация Ctrl+H")

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
	err := layoutLocal(g)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция layout вернула ошибку: <%v>", err))
		return fmt.Errorf("функция layout вернула ошибку: <%w>", err)
	}
	c.view.activeView = viewMain // Установка признака активного окна
	return nil
}

// Регистрация.
func (c *handlerUI) showRegistration(g *gocui.Gui, _ *gocui.View) error {

	c.conf.PtrLoggerFile.Write("Info: Нажата комбинация Ctrl+A")

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

	c.conf.PtrLoggerFile.Write("Info: Нажата комбинация Ctrl+B")

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

	// Проверка запуска в режиме - удалённый. Подключение и создание/обновление экземпляра.
	if c.conf.Flag.Mode == flags.ModeRemote {

		srv, err := server.New(c.typed.ip, c.typed.port)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Не удалось создать экземпляр сервера: <%v>", err))
			return fmt.Errorf("Не удалось создать экземпляр сервера: <%v>", err)
		}
		service.NewServer(srv)

		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: Соединение с сервером установлено: <%v>", err))
	}

	return nil
}

// Настройки.
func (c *handlerUI) showSettings(g *gocui.Gui, _ *gocui.View) error {

	c.conf.PtrLoggerFile.Write("Info: Нажата комбинация Ctrl+D")

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

	c.conf.PtrLoggerFile.Write("Info: Нажат Tab")

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
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Ошибка в функции SetCurrentView: <%v>", err))
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

	c.conf.PtrLoggerFile.Write("Info: Нажата комбинация Ctrl+C")

	// Ожидание завершения активных процессов.
	for {
		if c.getStatusBackUp() != stageActive &&
			c.getStatusRestore() != stageActive &&
			c.getStatusPopContainer() != stageActive &&
			c.getStatusPushContainer() != stageActive {
			break
		}
		time.Sleep(100 * time.Millisecond) // Ограничить использование ЦПУ.
	}

	return gocui.ErrQuit
}

// Обработка нажатия Enter.
func (c *handlerUI) handleEnter(g *gocui.Gui, v *gocui.View) error {

	c.conf.PtrLoggerFile.Write("Info: Нажат Enter")

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

	c.conf.PtrLoggerFile.Write("Info: Нажата комбинация Ctrl+N")

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

	// Если режим - Локальный.
	if c.conf.Flag.Mode == flags.ModeLocal {
		err := doRegistrationUserLocal(c)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция doRegistrationUserLocal, вернула ошибку: <%v>", err))
			return nil
		}
	}

	// Если режим - Удалённый.
	if c.conf.Flag.Mode == flags.ModeRemote {
		err := doRegistrationUserRemote(c)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция doRegistrationUserRemote, вернула ошибку: <%v>", err))
			return nil
		}
	}

	return nil
}

// Запуск процесса аутентификации пользователя.
func (c *handlerUI) doAuthenticationUser(gui *gocui.Gui, v *gocui.View) error {

	c.conf.PtrLoggerFile.Write("Info: Нажата комбинация Ctrl+L")

	// Если окно аутентификации.
	if c.view.activeView == viewAutentification {

		// Если режим - локальный.
		if c.conf.Flag.Mode == flags.ModeLocal {
			if err := doAuthenticationUserModeLocal(c); err != nil {
				c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция doAuthenticationUserModeLocal, вернуля ошибку: <%v>", err))
				return nil
			}
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: пользователь <%s>, прошел аутентификацию. Режим - локальный", c.typed.login))
		}

		// Если режим - удалённый.
		if c.conf.Flag.Mode == flags.ModeRemote {
			if err := doAuthenticationUserModeRemote(c); err != nil {
				c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция doAuthenticationUserModeRemote, вернуля ошибку: <%v>", err))
				return nil
			}
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: пользователь <%s>, прошел аутентификацию. Режим - удалённый", c.typed.login))
		}

		// Открытие окна, с запросом ввода дополнительного кода шифрования.
		if err := c.showRequestEncryptKey(gui, v); err != nil {
			return fmt.Errorf("Error: функция showRequestEncryptKey, вернуля ошибку: <%v>", err)
		}
	}

	return nil
}

// Запуск процесса сохранения данных.
func (c *handlerUI) doStore(gui *gocui.Gui, v *gocui.View) error {

	c.conf.PtrLoggerFile.Write("Info: Нажата комбинация Ctrl+F")

	switch c.view.activeView {
	case viewLoginPasswordData: // Окно - логин/пароль
		if err := doStoreViewLoginPasswordData(c); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция doStoreViewLoginPasswordData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewTextData: // Окно - текст.
		if err := doStoreViewTextData(c); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция doStoreViewTextData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewBankCardData: // Окно - банковские карты.
		if err := doStoreViewBankCardData(c); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция doStoreViewBankCardData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewBinaryData: // Окно - файлы.
		if err := doStoreViewBinaryData(c); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция doStoreViewBinaryData, вернула ошибку: <%v>", err))
			return nil
		}

	default:
	}

	return nil
}

// Отображение слудующего элемента.
func (c *handlerUI) doShowNextElement(gui *gocui.Gui, v *gocui.View) error {

	c.conf.PtrLoggerFile.Write("Info: Нажата комбинация Ctrl+E")

	switch c.view.activeView {
	case viewLoginPasswordData: // Окно логин/пароль
		if err := doShowNextElementViewLoginPasswordData(c, gui); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция doShowNextElementViewLoginPasswordData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewTextData: // Окно текста
		if err := doShowNextElementViewTextData(c, gui); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция doShowNextElementViewTextData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewBankCardData: // Окно банковских карт
		if err := doShowNextElementViewBankCardData(c, gui); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция doShowNextElementViewBankCardData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewBinaryData: // Окно файлов
		if err := doShowNextElementViewBinaryData(c, gui); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция doShowNextElementViewBinaryData, вернула ошибку: <%v>", err))
			return nil
		}

	default:
	}

	return nil
}

// Отображение предыдущего элемента.
func (c *handlerUI) doShowPrevElement(gui *gocui.Gui, v *gocui.View) error {

	c.conf.PtrLoggerFile.Write("Info: Нажата комбинация Ctrl+G")

	switch c.view.activeView {
	case viewLoginPasswordData: // Окно логин/пароль.
		if err := doShowPrevElementViewLoginPasswordData(c, gui); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция doShowPrevElementViewLoginPasswordData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewTextData: // Окно текста.
		if err := doShowPrevElementViewTextData(c, gui); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция doShowPrevElementViewTextData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewBankCardData: // Окно банковских карт.
		if err := doShowPrevElementViewBankCardData(c, gui); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция doShowPrevElementViewBankCardData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewBinaryData: // Окно файлов.
		if err := doShowPrevElementViewBinaryData(c, gui); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция doShowPrevElementViewBinaryData, вернула ошибку: <%v>", err))
			return nil
		}

	default:
	}

	return nil
}

// Удаление записи.
func (c *handlerUI) doDeleteElement(gui *gocui.Gui, v *gocui.View) error {

	c.conf.PtrLoggerFile.Write("Info: Нажата комбинация Ctrl+J")

	switch c.view.activeView {
	case viewLoginPasswordData: // Окно логин/пароль
		if err := doDeleteElementViewLoginPasswordData(c, gui); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция doDeleteElementViewLoginPasswordData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewTextData: // Окно текста
		if err := doDeleteElementViewTextData(c, gui); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция doDeleteElementViewTextData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewBankCardData: // Окно банковских карт
		if err := doDeleteElementViewBankCardData(c, gui); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция doDeleteElementViewBankCardData, вернула ошибку: <%v>", err))
			return nil
		}

	case viewBinaryData: // Окно файлов.
		if err := doDeleteElementViewBinaryData(c, gui); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция doDeleteElementViewBinaryData, вернула ошибку: <%v>", err))
			return nil
		}

	default:
	}

	return nil
}

// Извлечение.
func (c *handlerUI) doExtract(gui *gocui.Gui, v *gocui.View) error {

	c.conf.PtrLoggerFile.Write("Info: Нажата комбинация Ctrl+K")

	switch c.view.activeView {
	case viewBinaryData: // Окно работы с файлами

		if c.getStatusPopContainer() == stageNotActive && c.getStatusPushContainer() == stageNotActive {

			c.updateStatusPopContainer(stageActive)

			c.txrx.passedKB = 0
			c.txrx.percentTxRx = 0
			c.txrx.totalSizeKB = 0

			// Чтение буфера.
			v, err := gui.View("fieldShowFor")
			if err != nil {
				c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка доступа к элементу fieldShowFor: <%v>", err))
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

	c.conf.PtrLoggerFile.Write("Info: Нажата комбинация Ctrl+U")

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
	if c.conf.Flag.Mode == flags.ModeLocal {
		inst := container.New(containerName, c.secret.secretKey)
		c.conf.Container = inst

		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: стартовая обработка контейнера <%s> пройдена", containerName))
	}

	// Логика обработчика
	//
	c.view.activeView = "" // Сброс признака активного окна

	// Очистка видов.
	if err := deleteViews(g); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}

	// Сброс состояния
	layoutInitialized = false
	c.view.currentFocus = "selectLoginPassword" // Установка фокуса

	//
	// --- Поля ввода ---
	//

	// Для дополнительного секретного ключа.
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

	// имя клиента.
	if v, err := g.SetView("indicatorNameClient", 44, 18, inputWidth+12, inputHeight+1+17); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = false
		v.Frame = false
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorCyan
	}

	// процент выполнения.
	if v, err := g.SetView("indicatorPercent", 56, 20, inputWidth+7, inputHeight+1+19); err != nil {
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
	fmt.Fprintf(view, "%sПроцесс, может быть продолжительным. Дождитесь открытия окна.\n", strings.Repeat(" ", 35))

	//
	// --- Нижняя часть экрана ---
	//

	if c.conf.Flag.Mode == flags.ModeLocal { // Отбразить элемент, если режим - локальный.
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
	if c.conf.Flag.Mode == flags.ModeLocal { // Отбразить элемент, если режим - локальный.
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
	}

	// Установка фокуса.
	if _, err := g.SetCurrentView(c.view.currentFocus); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}

	// Формирование уникального ID клиента.
	if c.conf.Flag.Mode == flags.ModeRemote && c.clientName == "" {
		t := time.Now().UTC().Format("20060102150405.000")
		randStr, err := generateRandomString(10)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция generateRandomString, вернула ошибку: <%v>", err))
			return fmt.Errorf("ошибка при генерации случайной строки, для ID клиента: <%w>", err)
		}
		c.clientName = t + "-" + randStr
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: Создан ID клиента: <%s>", c.clientName))
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
	c.status.readNameLoginPaaswordPassed = false
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

	// Обработка режима - локальный.
	// Получение сохранённых значений логин/пароль.
	if c.conf.Flag.Mode == flags.ModeLocal {
		c.status.readLoginPaaswordPassed = true // Установка признака, что был запущен процесс получения значений логин/пароль.

		_, err = showLoginPasswordWorkDB(c)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция funcshowLoginPasswordWorkDB, вернула ошибку: <%v>", err))
			c.status.readLoginPaaswordSUCCESS = false
		} else {
			c.conf.PtrLoggerFile.Write("Debug: данные логин/пароль успешно прочитаны")
			c.status.readLoginPaaswordSUCCESS = true
		}
	}

	// Обработка режима - локальный.
	// Получение сохранённых значений логин/пароль.
	if c.conf.Flag.Mode == flags.ModeRemote {
		c.status.readNameLoginPaaswordPassed = true // Установка признака, что был запущен процесс получения имён логин/пароль.

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Запрос у сервера имен записей
		rxData, err := c.conf.Server.RequestLoginPasswordNames(ctx, c.conf.Server.GetTokenAuthentication(), c.clientName, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция RequestLoginPasswordNames, вернула ошибку: <%v>", err))
			c.status.readNameLoginPaaswordSUCCESS = false
		} else {
			c.data.namesLoginPassword = rxData // передача результата
			c.conf.PtrLoggerFile.Write("Debug: данные логин/пароль успешно прочитаны")
			c.status.readNameLoginPaaswordSUCCESS = true
		}
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
	c.status.readNameTextPassed = false
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
	if c.conf.Flag.Mode == flags.ModeLocal {
		c.status.readTextPassed = true // Установка признака, что был запущен процесс получения значений текста.

		_, err = showTextWorkDB(c)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция showTextWorkDB, вернула ошибку: <%v>", err))
		} else {
			c.conf.PtrLoggerFile.Write("Debug: данные текста успешно прочитаны")
			c.status.readTextSUCCESS = true
		}
	}

	if c.conf.Flag.Mode == flags.ModeRemote {

		c.status.readNameTextPassed = true // Установка признака, что был запущен процесс получения имён текста.

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Запрос у сервера имен записей
		rxData, err := c.conf.Server.RequestTextNames(ctx, c.conf.Server.GetTokenAuthentication(), c.clientName, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция RequestTextNames, вернула ошибку: <%v>", err))
			c.status.readNameTextSUCCESS = false
		} else {
			c.data.namesText = rxData // передача результата
			c.conf.PtrLoggerFile.Write("Debug: данные текста успешно прочитаны")
			c.status.readNameTextSUCCESS = true
		}
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

	c.updateStatusPopContainer(stageNotActive)
	c.updateStatusPushContainer(stageNotActive)

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

	// Процент выполнения.
	if v, err := g.SetView("indicatorPercent", 56, 20, inputWidth+7, inputHeight+1+19); err != nil {
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
		v.Wrap = false
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
		v.Wrap = false
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

	// Если режим - локальный
	if c.conf.Flag.Mode == flags.ModeLocal {
		c.data.files, err = c.conf.Container.ListFilesInContainer(c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция ListFilesInContainer, вернула ошибку: <%v>", err))
		} else {
			c.conf.PtrLoggerFile.Write("Debug: имена файлов в контейнере, успешно прочитаны")
			c.status.readFileSUCCESS = true
		}
	}

	// Если режим - удалённый
	if c.conf.Flag.Mode == flags.ModeRemote {

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
	c.status.readNameBankCardPassed = false
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

	// Окно для банковской карты.
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

	// Если режим - локальный.
	if c.conf.Flag.Mode == flags.ModeLocal {

		// Получение сохранённых значений банковских карт.
		c.status.readBankCardPassed = true // Установка признака, что был запущен процесс получения значений банковских карт.

		_, err = showBankCardWorkDB(c)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция showBankCardWorkDB, вернула ошибку: <%v>", err))
		} else {
			c.conf.PtrLoggerFile.Write("Debug: данные банковских карт успешно прочитаны")
			c.status.readBankCardSUCCESS = true
		}
	}

	// Если режим - удалённый.
	if c.conf.Flag.Mode == flags.ModeRemote {

		c.status.readNameBankCardPassed = true // Установка признака, что был запущен процесс получения имён банковских карт.

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Запрос у сервера имен записей
		rxData, err := c.conf.Server.RequestBankCardNames(ctx, c.conf.Server.GetTokenAuthentication(), c.clientName, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция RequestTextNames, вернула ошибку: <%v>", err))
			c.status.readNameBankCardSUCCESS = false
		} else {
			c.data.namesBankCard = rxData // передача результата
			c.conf.PtrLoggerFile.Write("Debug: данные банковской карты, успешно прочитаны")
			c.status.readNameBankCardSUCCESS = true
		}
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

	c.conf.PtrLoggerFile.Write("Info: Нажата комбинация Ctrl+O")

	// Запрет отработки, если уже есть активный процесс.
	if c.getStatusBackUp() == stageActive || c.getStatusRestore() == stageActive {
		return nil
	}

	// Логика работает только из окна выбора типа.
	if c.view.activeView == viewSelectType {

		// Установка признака, что запущен процесс передачи файлов на сервер.
		c.updateStatusRestore(stageNotActive) // Сброс состояния, чтобы убрать подсветку.
		c.updateStatusBackUp(stageActive)

		// Сброс данных.
		c.txrx.passedKB = 0
		c.txrx.percentTxRx = 0
		c.txrx.totalSizeKB = 0

		// Логика процесса.
		//
		files := []string{"manager.db", "container.data"}

		// Определение общего размера файлов.
		c.txrx.totalSizeKB, err = totalFileSize(files)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция totalFileSize, вернула ошибку: <%v>", err))
			c.updateStatusBackUp(stageFault)
			return nil
		}

		// Передача файлов.
		go doBackupProcess(files, c)
	}
	return nil
}

// Получение данных клиента, от сервер.
func (c *handlerUI) doRestore(gui *gocui.Gui, v *gocui.View) error {

	c.conf.PtrLoggerFile.Write("Info: Нажата комбинация Ctrl+P")

	// Запрет отработки, если уже есть активный процесс.
	if c.getStatusBackUp() == stageActive || c.getStatusRestore() == stageActive {
		return nil
	}

	// Логика работает только из окна выбора типа.
	if c.view.activeView == viewSelectType {

		// Установка признака, что запущен процесс приёма файлов от сервера.
		c.updateStatusBackUp(stageNotActive) // сброс признака, чтобы убрать подсветку.
		c.updateStatusRestore(stageActive)

		// Сброс данных.
		c.txrx.passedKB = 0
		c.txrx.percentTxRx = 0
		c.txrx.totalSizeKB = 0

		// Приём файлов.
		files := []string{"manager.db", "container.data"}
		go doRestoreProcess(files, c)
	}
	return nil
}

// Обновление статуса процесса передачи.
func (c *handlerUI) updateStatusBackUp(st int) {

	c.mutex.statusBackUp.Lock()
	defer c.mutex.statusBackUp.Unlock()

	c.status.backUp = st
}

// Получение текущего статуса процесса передачи.
func (c *handlerUI) getStatusBackUp() int {

	c.mutex.statusBackUp.Lock()
	defer c.mutex.statusBackUp.Unlock()

	return c.status.backUp
}

// Обновление статуса процесса передачи.
func (c *handlerUI) updateStatusRestore(st int) {

	c.mutex.statusRestore.Lock()
	defer c.mutex.statusRestore.Unlock()

	c.status.restore = st
}

// Получение текущего статуса процесса передачи.
func (c *handlerUI) getStatusRestore() int {

	c.mutex.statusRestore.Lock()
	defer c.mutex.statusRestore.Unlock()

	return c.status.restore
}

// Получение процента выполения TxRx.
func (c *handlerUI) getPercentTxRx() float32 {

	c.mutex.processTxRx.Lock()
	defer c.mutex.processTxRx.Unlock()

	return c.txrx.percentTxRx
}

// Установка процента выполения TxRx.
func (c *handlerUI) setPercentTxRx(percent float32) {

	c.mutex.processTxRx.Lock()
	defer c.mutex.processTxRx.Unlock()

	c.txrx.percentTxRx = percent
}

// Обновление статуса процесса передачи.
func (c *handlerUI) updateStatusPushContainer(st int) {

	c.mutex.statusPushContainer.Lock()
	defer c.mutex.statusPushContainer.Unlock()

	c.status.pushContainer = st
}

// Получение текущего статуса процесса передачи.
func (c *handlerUI) getStatusPushContainer() int {

	c.mutex.statusPushContainer.Lock()
	defer c.mutex.statusPushContainer.Unlock()

	return c.status.pushContainer
}

// Обновление статуса процесса передачи.
func (c *handlerUI) updateStatusPopContainer(st int) {

	c.mutex.statusPopContainer.Lock()
	defer c.mutex.statusPopContainer.Unlock()

	c.status.popContainer = st
}

// Получение текущего статуса процесса передачи.
func (c *handlerUI) getStatusPopContainer() int {

	c.mutex.statusPopContainer.Lock()
	defer c.mutex.statusPopContainer.Unlock()

	return c.status.popContainer
}
