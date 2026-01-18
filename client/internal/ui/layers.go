// Слои обработчиков пакета.
package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Part001-R/YaPr-GP-2/client/internal/adapters/container"
	"github.com/Part001-R/YaPr-GP-2/client/internal/adapters/server"
	"github.com/Part001-R/YaPr-GP-2/client/internal/domain"
	service "github.com/Part001-R/YaPr-GP-2/client/internal/service/udt"
	"github.com/Part001-R/YaPr-GP-2/client/internal/utils/flags"
	"github.com/Part001-R/YaPr-GP-2/proto"
	"github.com/jroimartin/gocui"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

//
// --- pingContext ---
//

// Додготовка данных для запроса. Возвращаются метаданные, имя токена, ключ и ошибка.
//
// Параметры:
//
//	c - экземпляр интерфейса.
func layerDataPingContextPrepare(c *handlerUI) (txMD metadata.MD, nameToken string, secretKey string, err error) {

	// Проверка аргументов.
	if c == nil {
		return nil, "", "", NilPtrArgumentC
	}

	// Создание ключа.
	secretKey, err = generateRandomString(50)
	if err != nil {
		return nil, "", "", fmt.Errorf("функция generateRandomString, вернула ошибку: <%w>", err)
	}

	// Создание токена.
	timeValidToken := time.Duration(5 * time.Second)
	txToken, err := createToken("clientManager", secretKey, timeValidToken)
	if err != nil {
		return nil, "", "", fmt.Errorf("функция createToken, вернула ошибку: <%w>", err)
	}

	// Заполнение метаданных.
	nameToken = "token"
	txMD = metadata.Pairs(nameToken, txToken)

	// Результат.
	return txMD, nameToken, secretKey, nil
}

// Запрос. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	txMD - метаданные.
//	client - клиент.
//	nameToken - имя токена.
//	key - ключ.
func layerPingContextRequest(ctx context.Context, txMD metadata.MD, client proto.PasswordManagerClient, nameToken string, key string) error {

	emptyRequest := &emptypb.Empty{}
	var header metadata.MD

	// Запрос.
	ctx = metadata.NewOutgoingContext(ctx, txMD) // добавление метаданных к контексту.
	_, err := client.Ping(ctx, emptyRequest, grpc.Header(&header))
	if err != nil {
		return fmt.Errorf("функция client.Ping, вернула ошибку: <%w>", err)
	}

	// Получение токена из метаданных ответа.
	token := header[nameToken]
	if len(token) == 0 || token[0] == "" {
		return MissingTokenData
	}
	rxToken := token[0]

	// Проверка токена.
	if err := checkToken(rxToken, key); err != nil {
		return fmt.Errorf("функция checkToken, вернула ошибку: <%w>", err)
	}

	return nil
}

//
// --- showRegistration ---
//

// Ограничение на запуск функционала. Возвращается true - есть ограничение.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowRegistrationIsRestraintRun(c *handlerUI) bool {

	if c.view.activeView == viewAutentification ||
		c.view.activeView == viewBankCardData ||
		c.view.activeView == viewBinaryData ||
		c.view.activeView == viewLoginPasswordData ||
		c.view.activeView == viewRegistration ||
		c.view.activeView == viewRequestSecretKey ||
		c.view.activeView == viewSelectType ||
		c.view.activeView == viewSettings ||
		c.view.activeView == viewTextData {
		return true
	}

	return false
}

// Сброс переменных. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowRegistrationReset(c *handlerUI) error {

	layoutInitialized = false
	c.view.activeView = ""
	c.typed.login = ""
	c.typed.password1 = ""
	c.typed.password2 = ""
	c.status.addUserPassed = false
	c.status.addUserSUCCESS = false

	return nil
}

// Отрисовка окна. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowRegistrationDrawWindow(c *handlerUI, g *gocui.Gui) (*gocui.View, error) {

	view, err := g.SetView(viewRegistration, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
		return nil, fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
	}
	view.Title = "Регистрация"
	view.Wrap = true
	view.Clear()

	return view, nil
}

// Отрисовка полей ввода. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
//	_ - заглушка на View.
func layerShowRegistrationDrawInput(c *handlerUI, g *gocui.Gui, _ *gocui.View) error {

	if v, err := g.SetView("Login", 50, 2, inputWidth+1, inputHeight+1+1); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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

	return nil
}

// Отрисовка индикаторов. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowRegistrationDrawIndicator(c *handlerUI, g *gocui.Gui) error {

	// Признак корректных данных пользователя
	indicatorY := inputHeight*3 + 7
	indicatorX := 58
	v, err := g.SetView("indicator-match", indicatorX, indicatorY, indicatorX+20, indicatorY+2)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
		c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
	}
	v.Frame = false
	v.BgColor = gocui.ColorDefault
	v.FgColor = gocui.ColorGreen

	return nil
}

// Отрисовка пояснений. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowRegistrationDrawGuide(c *handlerUI, g *gocui.Gui) error {

	if v, err := g.SetView("TAB", 1, 26, inputWidth-55, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
			return fmt.Errorf("Не удалось установить фокус: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+W - регистрация"))
	}

	return nil
}

// Установка фокуса. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
//	name - Имя элемента установки фокуса.
func layerShowRegistrationSetFocus(c *handlerUI, g *gocui.Gui, name string) error {

	if _, err := g.SetCurrentView(name); err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}

	layoutInitialized = true
	c.view.currentFocus = name
	c.view.activeView = viewRegistration

	// Установка стиля фокуса
	if v, err := g.View(name); err == nil {
		c.setFocusStyle(v, name)
	} else {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось получить вид для %s: <%v>", name, err))
		return fmt.Errorf("Не удалось получить вид для %s: <%w>", name, err)
	}

	// Принудительное обновление интерфейса
	g.Update(func(*gocui.Gui) error {
		return nil
	})

	return nil
}

//
// --- showAuthentication ---
//

// Ограничение на запуск функционала. Возвращается true - есть ограничение.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowAuthenticationIsRestraintRun(c *handlerUI) bool {

	if c.view.activeView == viewAutentification ||
		c.view.activeView == viewBankCardData ||
		c.view.activeView == viewBinaryData ||
		c.view.activeView == viewLoginPasswordData ||
		c.view.activeView == viewRegistration ||
		c.view.activeView == viewRequestSecretKey ||
		c.view.activeView == viewSelectType ||
		c.view.activeView == viewSettings ||
		c.view.activeView == viewTextData {
		return true
	}
	return false
}

// Создание экземпляра БД. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowAuthenticationIsNewInstDB(c *handlerUI) error {

	storage, err := domain.NewStorage(c.conf.Flag.DSN)
	if err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("Error: Ошибка подключения к БД:<%v>", err))
		return fmt.Errorf("Error: Ошибка подключения к БД:<%v>", err)
	}
	c.conf.ActionsDB = storage

	return nil
}

// Сброс переменных. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowAuthenticationReset(c *handlerUI) error {

	layoutInitialized = false
	c.view.activeView = ""
	c.typed.login = ""
	c.typed.password1 = ""
	c.typed.password2 = ""

	return nil
}

// Очистка экрана. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowAuthenticationClear(c *handlerUI, g *gocui.Gui) error {

	if err := deleteViews(g); err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}

	return nil
}

// Отрисовка окна. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowAuthenticationDrawWindow(c *handlerUI, g *gocui.Gui) (*gocui.View, error) {

	view, err := g.SetView(viewAutentification, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
		return nil, fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
	}
	view.Title = "Аутентификация"
	view.Wrap = true
	view.Clear()

	return view, nil
}

// Отрисовка полей ввода. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
//	_ - заглушка на View.
func layerShowAuthenticationDrawInput(c *handlerUI, g *gocui.Gui, _ *gocui.View) error {

	if v, err := g.SetView("Login", 50, 2, inputWidth+1, inputHeight+1+1); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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

	return nil
}

// Отрисовка индикаторов. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowAuthenticationDrawIndicator(c *handlerUI, g *gocui.Gui) error {

	indicatorY := inputHeight*3 + 7
	indicatorX := 58
	v, err := g.SetView("indicator-match", indicatorX, indicatorY, indicatorX+20, indicatorY+2)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
	}
	v.Frame = false
	v.BgColor = gocui.ColorDefault
	v.FgColor = gocui.ColorGreen

	return nil
}

// Отрисовка пояснений. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowAuthenticationDrawGuide(c *handlerUI, g *gocui.Gui) error {

	if v, err := g.SetView("TAB", 1, 26, inputWidth-55, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
			return fmt.Errorf("Не удалось установить фокус: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+L - Подключение"))
	}

	return nil
}

// Установка фокуса. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
//	name - Имя элемента установки фокуса.
func layerShowAuthenticationSetFocus(c *handlerUI, g *gocui.Gui, name string) error {

	if _, err := g.SetCurrentView(name); err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}

	layoutInitialized = true
	c.view.currentFocus = name
	c.view.activeView = viewAutentification

	// Установка стиля фокуса
	if v, err := g.View(name); err == nil {
		c.setFocusStyle(v, name)
	} else {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось получить вид для %s: <%v>", name, err))
		return fmt.Errorf("Не удалось получить вид для %s: <%w>", name, err)
	}

	// Принудительное обновление интерфейса
	g.Update(func(*gocui.Gui) error {
		return nil
	})

	return nil
}

// Создание экземпляра сервера. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowAuthenticationNewInst(c *handlerUI) error {

	srv, err := server.New(c.typed.ip, c.typed.port)
	if err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("Error: Не удалось создать экземпляр сервера: <%v>", err))
		return fmt.Errorf("Не удалось создать экземпляр сервера: <%v>", err)
	}
	service.NewServer(srv)

	c.conf.LgrFile.Write("Info: Экземпляр сервера создан")

	return nil
}

//
// --- showSettings ---
//

// Ограничение на запуск функционала. Возвращается true - есть ограничение.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowSettingsIsRestraintRun(c *handlerUI) bool {

	if c.view.activeView == viewAutentification ||
		c.view.activeView == viewBankCardData ||
		c.view.activeView == viewBinaryData ||
		c.view.activeView == viewLoginPasswordData ||
		c.view.activeView == viewRegistration ||
		c.view.activeView == viewRequestSecretKey ||
		c.view.activeView == viewSelectType ||
		c.view.activeView == viewSettings ||
		c.view.activeView == viewTextData {
		return true
	}

	return false
}

// Сброс переменных. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowSettingsReset(c *handlerUI) error {

	layoutInitialized = false
	c.view.activeView = "" // Сброс признака активного окна

	return nil
}

// Очистка экрана. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowSettingsClear(c *handlerUI, g *gocui.Gui) error {

	if err := deleteViews(g); err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}

	return nil
}

// Отрисовка окна. Возвращается указатель на окно и ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowSettingsDrawWindow(c *handlerUI, g *gocui.Gui) (*gocui.View, error) {

	view, err := g.SetView(viewSettings, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return nil, fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	view.Title = "Настройки"
	view.Wrap = true
	view.Clear()

	return view, nil
}

// Отрисовка полей ввода. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
//	_ - заглушка на View.
func layerShowSettingsDrawInput(c *handlerUI, g *gocui.Gui, _ *gocui.View) error {

	if v, err := g.SetView("IP", 50, 2, inputWidth+1, inputHeight+1+1); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
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

	return nil
}

// Отрисовка пояснений. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowSettingsDrawGuide(c *handlerUI, g *gocui.Gui, view *gocui.View) error {

	if v, err := g.SetView("TAB", 1, 26, inputWidth-55, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
			return fmt.Errorf("Не удалось установить фокус: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+N - тест связи"))
	}

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 12))
	fmt.Fprintf(view, "%sУкажите данные сервера и выполните тест.\n", strings.Repeat(" ", 45))

	return nil
}

// Установка фокуса. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
//	name - Имя элемента установки фокуса.
func layerShowSettingsSetFocus(c *handlerUI, g *gocui.Gui, name string) error {

	if _, err := g.SetCurrentView(name); err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}

	layoutInitialized = true
	c.view.currentFocus = name
	c.view.activeView = viewSettings

	// Установка стиля фокуса
	if v, err := g.View(name); err == nil {
		c.setFocusStyle(v, name)
	} else {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось получить вид для %s: <%v>", name, err))
		return fmt.Errorf("Не удалось получить вид для %s: <%w>", name, err)
	}

	// Принудительное обновление интерфейса
	g.Update(func(*gocui.Gui) error {
		return nil
	})

	return nil
}

//
// --- showMain ---
//

// Ограничение на запуск функционала. Возвращается true - есть ограничение.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowMainIsRestraintRun(c *handlerUI) bool {

	if c.status.backUp == stageActive ||
		c.status.restore == stageActive {
		return true
	}

	return false
}

// Сброс переменных. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowMainReset(c *handlerUI) error {

	c.view.activeView = ""
	layoutInitialized = false
	c.view.currentFocus = ""

	return nil
}

// Очистка экрана. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowMainClear(c *handlerUI, g *gocui.Gui) error {

	if err := deleteViews(g); err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("Error: функция deleteViews вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews вернула ошибку: <%w>", err)
	}

	return nil
}

// Установка фокуса. Возвращается ошибка.
func layerShowMainSetFocus(c *handlerUI) error {

	c.view.activeView = viewMain

	return nil
}

//
// --- showSelectType ---
//

// Ограничение на запуск функционала. Возвращается true - есть ограничение.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowSelectTypeIsRestraintRun(c *handlerUI) bool {

	if c.view.activeView != viewLoginPasswordData &&
		c.view.activeView != viewTextData &&
		c.view.activeView != viewBankCardData &&
		c.view.activeView != viewBinaryData &&
		c.view.activeView != viewRequestSecretKey {
		return true
	}

	return false
}

// Закрытие подключения к БД. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowSelectTypeCloseDB(c *handlerUI) error {

	if c.conf.Flag.Mode == flags.ModeLocal {
		if err := c.conf.ActionsDB.Close(); err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Ошибка закрытия подключения к БД:<%v>, перед выходом из раздела логин/пароль", err))
			return fmt.Errorf("Ошибка закрытия подключения к БД: <%w>", err)
		}
	}

	return nil
}

// Сброс переменных. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowSelectTypeReset(c *handlerUI) error {

	c.status.backUp = stageNotActive
	c.status.restore = stageNotActive
	c.view.activeView = ""
	layoutInitialized = false

	return nil
}

// Создание экземпляра контейнера. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowSelectTypeNewInstContainer(c *handlerUI) error {

	if c.conf.Flag.Mode == flags.ModeLocal {
		inst, err := container.New(c.conf.Flag.LocalNameContainer, c.secret.secretKey)
		if err != nil {
			return fmt.Errorf("Ошибка создания контейнера:<%w>", err)
		}
		c.conf.Container = inst
		c.conf.LgrFile.Write(fmt.Sprintf("Info: стартовая обработка контейнера <%s> пройдена", c.conf.Flag.LocalNameContainer))
	}

	return nil
}

// Очистка экрана. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowSelectTypeClear(c *handlerUI, g *gocui.Gui) error {

	if err := deleteViews(g); err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}

	return nil
}

// Отрисовка окна. Возвращается указатель на окно и ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowSelectDrawWindow(c *handlerUI, g *gocui.Gui) (*gocui.View, error) {

	view, err := g.SetView(viewSelectType, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return nil, fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	view.Title = "Тип данных"
	view.Wrap = true
	view.Clear()

	return view, nil
}

// Отрисовка типов данных. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
//	_ - заглушка на View.
func layerShowSelectDrawTypes(c *handlerUI, g *gocui.Gui, _ *gocui.View) error {

	if v, err := g.SetView("selectLoginPassword", 50, 2, inputWidth+1, inputHeight+1+1); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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

	return nil
}

// Отрисовка индикаторов. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowSelectDrawIndicator(c *handlerUI, g *gocui.Gui) error {

	// имя клиента.
	if v, err := g.SetView("indicatorNameClient", 44, 18, inputWidth+12, inputHeight+1+17); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = false
		v.Frame = false
		v.BgColor = gocui.ColorDefault
		v.SelBgColor = gocui.ColorDefault
		v.SelFgColor = gocui.ColorDefault
	}

	return nil
}

// Отрисовка пояснений. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowSelectDrawGuide(c *handlerUI, g *gocui.Gui, view *gocui.View) error {

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 15))
	fmt.Fprintf(view, "%sВыберите нужный раздел через Tab и нажмите Enter.\n", strings.Repeat(" ", 40))
	fmt.Fprintf(view, "%sПроцесс, может быть продолжительным. Дождитесь открытия окна.\n", strings.Repeat(" ", 35))

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 1))
	fmt.Fprintf(view, "%sПосле выполнения процесса Ctrl+P, перезапустите приложение.\n", strings.Repeat(" ", 36))

	//
	// --- Нижняя часть экрана ---
	//

	if c.conf.Flag.Mode == flags.ModeLocal { // Отбразить элемент, если режим - локальный.
		// Верхний ряд.
		if v, err := g.SetView("Backup", 1, 23, inputWidth-55, inputHeight+1+22); err != nil {
			if err != gocui.ErrUnknownView {
				c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
				c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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

	return nil
}

// Создание уникального id клиента. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowSelectCreateClientID(c *handlerUI) error {

	if c.conf.Flag.Mode == flags.ModeRemote && c.clientName == "" {
		t := time.Now().UTC().Format("20060102150405.000")
		randStr, err := generateRandomString(10)
		if err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция generateRandomString, вернула ошибку: <%v>", err))
			return fmt.Errorf("ошибка при генерации случайной строки, для ID клиента: <%w>", err)
		}
		c.clientName = t + "-" + randStr
		c.conf.LgrFile.Write(fmt.Sprintf("Info: Создан ID клиента: <%s>", c.clientName))
	}

	return nil
}

// Установка фокуса. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
//	name - Имя элемента установки фокуса.
func layerShowSelectSetFocus(c *handlerUI, g *gocui.Gui, name string) error {

	if _, err := g.SetCurrentView(name); err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}

	layoutInitialized = true
	c.view.currentFocus = name
	c.view.activeView = viewSelectType

	// Установка стиля фокуса
	if v, err := g.View(name); err == nil {
		c.setFocusStyle(v, name)
	} else {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось получить вид для %s: <%v>", name, err))
		return fmt.Errorf("Не удалось получить вид для %s: <%w>", name, err)
	}

	// Принудительное обновление интерфейса
	g.Update(func(*gocui.Gui) error {
		return nil
	})

	return nil
}

//
// --- showLoginPassword ---
//

// Ограничение на запуск функционала. Возвращается true - есть ограничение.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowLoginPasswordRestraintRun(c *handlerUI) bool {

	if c.view.activeView != viewSelectType {
		return true
	}

	return false
}

// Создание экземпляра БД. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowLoginPasswordNewInstDB(c *handlerUI) error {

	if c.conf.Flag.Mode == flags.ModeLocal {
		storage, err := domain.NewStorage(c.conf.Flag.DSN)
		if err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Ошибка подключения к БД:<%v>", err))
			return fmt.Errorf("Функция NewStorage, вернула ошибку:<%w>", err)
		}
		c.conf.ActionsDB = storage
	}

	return nil
}

// Сброс переменных. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowLoginPasswordReset(c *handlerUI) error {

	c.status.readLoginPaaswordPassed = false
	c.status.readNameLoginPaaswordPassed = false
	c.status.addLoginPaaswordPassed = false
	c.status.delLoginPaaswordPassed = false
	c.index.loginPassword = 0
	c.view.activeView = ""
	layoutInitialized = false

	return nil
}

// Очистка экрана. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowLoginPasswordClear(c *handlerUI, g *gocui.Gui) error {

	if err := deleteViews(g); err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}

	return nil
}

// Отрисовка окна. Возвращается указатель на окно и ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowLoginPasswordWindow(c *handlerUI, g *gocui.Gui) (*gocui.View, error) {

	view, err := g.SetView(viewLoginPasswordData, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return nil, fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	view.Title = "Логины - пароли"
	view.Wrap = true
	view.Clear()
	return view, nil
}

// Отрисовка полей ввода. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
//	view - указатель на окно.
func layerShowLoginPasswordDrawInput(c *handlerUI, g *gocui.Gui, view *gocui.View) error {

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 3))
	fmt.Fprintf(view, "%sПросмотр.\n", strings.Repeat(" ", 56))

	if v, err := g.SetView("fieldShowFor", 1, 4, inputWidth-48, inputHeight+1+3); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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

	return nil
}

// Отрисовка индикаторов. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowLoginPasswordDrawIndicator(c *handlerUI, g *gocui.Gui) error {

	indicatorY := inputHeight * 3
	indicatorX := 51
	vRead, err := g.SetView("indicatorReadStatus", indicatorX, indicatorY, indicatorX+20, indicatorY+2)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
		c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
	}
	vAdd.Frame = false
	vAdd.BgColor = gocui.ColorDefault
	vAdd.FgColor = gocui.ColorDefault

	return nil
}

// Отрисовка пояснений. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowLoginPasswordDrawGuide(c *handlerUI, g *gocui.Gui, view *gocui.View) error {

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 8))
	fmt.Fprintf(view, "%sПри изменении данных, выполните Crl+U.\n", strings.Repeat(" ", 42))

	//
	// --- Нижняя часть экрана ---
	//

	// Верхний ряд.
	if v, err := g.SetView("Save", 1, 23, inputWidth-55, inputHeight+1+22); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+U - назад"))
	}

	return nil
}

// Действия обработчика. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowLoginPasswordActions(c *handlerUI) error {

	// Если режим - локальный.
	if c.conf.Flag.Mode == flags.ModeLocal {
		c.status.readNameLoginPaaswordPassed = true // Установка признака, что был запущен процесс получения записей.

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Запрос у сервера имен записей
		rxData, err := c.conf.ActionsDB.ReadNamesTableLoginPasswordContext(ctx)
		if err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: функция RequestLoginPasswordNames, вернула ошибку: <%v>", err))
			c.status.readNameLoginPaaswordSUCCESS = false
		} else {
			c.data.namesLoginPassword = rxData // передача результата
			c.conf.LgrFile.Write("Debug: имена записей логин/пароль, прочитаны")
			c.status.readNameLoginPaaswordSUCCESS = true
		}
		return nil
	}

	// Если режим - удалённый.
	if c.conf.Flag.Mode == flags.ModeRemote {
		c.status.readNameLoginPaaswordPassed = true // Установка признака, что был запущен процесс получения имён логин/пароль.

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Запрос у сервера имен записей
		rxData, err := c.conf.Server.RequestLoginPasswordNames(ctx, c.conf.Server.GetTokenAuthentication(), c.secret.secretKey)
		if err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: функция RequestLoginPasswordNames, вернула ошибку: <%v>", err))
			c.status.readNameLoginPaaswordSUCCESS = false
		} else {
			c.data.namesLoginPassword = rxData // передача результата
			c.conf.LgrFile.Write("Debug: данные логин/пароль успешно прочитаны")
			c.status.readNameLoginPaaswordSUCCESS = true
		}
	}

	return nil
}

// Установка фокуса. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
//	name - Имя элемента установки фокуса.
func layerShowLoginPasswordSetFocus(c *handlerUI, g *gocui.Gui, name string) error {

	if _, err := g.SetCurrentView(name); err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}

	layoutInitialized = true
	c.view.currentFocus = name
	c.view.activeView = viewLoginPasswordData // Установка признака активного окна

	// Установка стиля фокуса
	if v, err := g.View(name); err == nil {
		c.setFocusStyle(v, name)
	} else {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось получить вид для %s: <%v>", name, err))
		return fmt.Errorf("Не удалось получить вид для %s: <%w>", name, err)
	}

	// Принудительное обновление интерфейса
	g.Update(func(*gocui.Gui) error {
		return nil
	})

	return nil
}

//
// --- showText ---
//

// Ограничение на запуск функционала. Возвращается true - есть ограничение.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowTextRestraintRun(c *handlerUI) bool {

	if c.view.activeView != viewSelectType {
		return true
	}

	return false
}

// Создание экземпляра БД. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowTextNewInstDB(c *handlerUI) error {

	if c.conf.Flag.Mode == flags.ModeLocal {
		storage, err := domain.NewStorage(c.conf.Flag.DSN)
		if err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Ошибка подключения к БД:<%v>", err))
			return fmt.Errorf("Функция NewStorage, вернула ошибку:<%w>", err)
		}
		c.conf.ActionsDB = storage
	}

	return nil
}

// Сброс переменных. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowTextReset(c *handlerUI) error {

	c.status.readTextPassed = false
	c.status.readNameTextPassed = false
	c.status.addTextPassed = false
	c.status.delTextPassed = false
	c.status.readTextSUCCESS = false
	c.index.text = 0
	c.view.activeView = ""
	layoutInitialized = false

	return nil
}

// Очистка экрана. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowTextClear(c *handlerUI, g *gocui.Gui) error {

	if err := deleteViews(g); err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}

	return nil
}

// Отрисовка окна. Возвращается указатель на окно и ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowTextdWindow(c *handlerUI, g *gocui.Gui) (*gocui.View, error) {

	view, err := g.SetView(viewTextData, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return nil, fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	view.Title = "Текст"
	view.Wrap = true
	view.Clear()

	return view, nil
}

// Отрисовка индикаторов. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowTextDrawIndicator(c *handlerUI, g *gocui.Gui) error {

	// Результат чтения данных.
	indicatorY := inputHeight * 3
	indicatorX := 51
	vRead, err := g.SetView("indicatorReadStatus", indicatorX, indicatorY, indicatorX+20, indicatorY+2)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
		c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
	}
	vAdd.Frame = false
	vAdd.BgColor = gocui.ColorDefault
	vAdd.FgColor = gocui.ColorDefault

	return nil
}

// Отрисовка полей ввода. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowTextDrawInput(c *handlerUI, g *gocui.Gui) error {

	if v, err := g.SetView("fieldShowFor", 1, 4, inputWidth-48, inputHeight+1+3); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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

	if v, err := g.SetView("fieldAddFor", 1, 12, inputWidth-48, inputHeight+1+11); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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

	return nil
}

// Отрисовка пояснений. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowTextDrawGuide(c *handlerUI, g *gocui.Gui, view *gocui.View) error {

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 3))
	fmt.Fprintf(view, "%sПросмотр.\n", strings.Repeat(" ", 56))

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 7))
	fmt.Fprintf(view, "%sДобавление.\n", strings.Repeat(" ", 55))

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 8))
	fmt.Fprintf(view, "%sПри изменении данных, выполните Crl+U.\n", strings.Repeat(" ", 42))

	// Верхний ряд.
	if v, err := g.SetView("Save", 1, 23, inputWidth-55, inputHeight+1+22); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+U - назад"))
	}

	return nil
}

// Действия обработчика. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowTextActions(c *handlerUI) error {

	// Если режим - локальный.
	if c.conf.Flag.Mode == flags.ModeLocal {
		c.status.readNameTextPassed = true // Установка признака, что был запущен процесс получения значений текста.

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		rxData, err := c.conf.ActionsDB.ReadNamesTableTextContext(ctx)
		if err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: функция ReadNamesTableTextContext, вернула ошибку: <%v>", err))
		} else {
			c.data.namesText = rxData
			c.conf.LgrFile.Write("Debug: данные текста успешно прочитаны")
			c.status.readNameTextSUCCESS = true
		}
	}

	// Если режим - удалённый.
	if c.conf.Flag.Mode == flags.ModeRemote {
		c.status.readNameTextPassed = true // Установка признака, что был запущен процесс получения имён текста.

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Запрос у сервера имен записей
		rxData, err := c.conf.Server.RequestTextNames(ctx, c.conf.Server.GetTokenAuthentication(), c.secret.secretKey)
		if err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: функция RequestTextNames, вернула ошибку: <%v>", err))
		} else {
			c.data.namesText = rxData // передача результата
			c.conf.LgrFile.Write("Debug: данные текста успешно прочитаны")
			c.status.readNameTextSUCCESS = true
		}
	}

	return nil
}

// Установка фокуса. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
//	name - Имя элемента установки фокуса.
func layerShowTextSetFocus(c *handlerUI, g *gocui.Gui, name string) error {

	if _, err := g.SetCurrentView(name); err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}

	layoutInitialized = true
	c.view.currentFocus = name
	c.view.activeView = viewTextData // Установка признака активного окна

	// Установка стиля фокуса
	if v, err := g.View(name); err == nil {
		c.setFocusStyle(v, name)
	} else {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось получить вид для %s: <%v>", name, err))
		return fmt.Errorf("Не удалось получить вид для %s: <%w>", name, err)
	}

	// Принудительное обновление интерфейса
	g.Update(func(*gocui.Gui) error {
		return nil
	})

	return nil
}

//
// --- showBankCard ---
//

// Ограничение на запуск функционала. Возвращается true - есть ограничение.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowBankCardRestraintRun(c *handlerUI) bool {

	if c.view.activeView != viewSelectType {
		return true
	}
	return false
}

// Создание экземпляра БД. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowBankCardNewInstDB(c *handlerUI) error {

	if c.conf.Flag.Mode == flags.ModeLocal {
		storage, err := domain.NewStorage(c.conf.Flag.DSN)
		if err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Ошибка подключения к БД:<%v>", err))
			return fmt.Errorf("Функция NewStorage, вернула ошибку:<%w>", err)
		}
		c.conf.ActionsDB = storage
	}

	return nil
}

// Сброс переменных. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowBankCardReset(c *handlerUI) error {

	c.status.readBankCardPassed = false
	c.status.readNameBankCardPassed = false
	c.status.addBankCardPassed = false
	c.status.delBankCardPassed = false
	c.index.bankCard = 0
	c.status.readBankCardSUCCESS = false
	c.view.activeView = ""
	layoutInitialized = false
	c.status.readNameBankCardSUCCESS = false

	return nil
}

// Очистка экрана. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowBankCardClear(c *handlerUI, g *gocui.Gui) error {

	if err := deleteViews(g); err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}
	return nil
}

// Отрисовка окна. Возвращается указатель на окно и ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowBankCardWindow(c *handlerUI, g *gocui.Gui) (*gocui.View, error) {

	view, err := g.SetView(viewBankCardData, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return nil, fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	view.Title = "Банковские карты"
	view.Wrap = true
	view.Clear()

	return view, nil
}

// Отрисовка пояснений. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowBankCardDrawGuide(c *handlerUI, g *gocui.Gui, view *gocui.View) error {

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 1))
	fmt.Fprintf(view, "%sПросмотр.\n", strings.Repeat(" ", 56))

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 10))
	fmt.Fprintf(view, "%sДобавление. %sПри изменении данных, выполните Crl+U\n", strings.Repeat(" ", 55), strings.Repeat(" ", 15))

	// Верхний ряд.
	if v, err := g.SetView("Save", 1, 23, inputWidth-55, inputHeight+1+22); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+U - назад"))
	}

	return nil

}

// Отрисовка полей вывода. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowBankCardDrawOut(c *handlerUI, g *gocui.Gui) error {

	if v, err := g.SetView("fieldAddFor", 1, 13, inputWidth-48, inputHeight+1+12); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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

	return nil
}

// Отрисовка полей ввода. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowBankCardDrawInput(c *handlerUI, g *gocui.Gui) error {

	if v, err := g.SetView("fieldShowFor", 1, 2, inputWidth-48, inputHeight+1+1); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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

	return nil
}

// Отрисовка индикаторов. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowBankCardIndicator(c *handlerUI, g *gocui.Gui) error {

	// Результат чтения данных.
	indicatorY := inputHeight*3 - 1
	indicatorX := 80
	vRead, err := g.SetView("indicatorReadStatus", indicatorX, indicatorY, indicatorX+20, indicatorY+2)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
		c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
	}
	vAdd.Frame = false
	vAdd.BgColor = gocui.ColorDefault
	vAdd.FgColor = gocui.ColorDefault

	return nil
}

// Действия обработчика. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowBankCardActions(c *handlerUI) error {

	// Если режим - локальный.
	if c.conf.Flag.Mode == flags.ModeLocal {

		c.status.readNameBankCardPassed = true // Установка признака, что был запущен процесс получения имён банковских карт.

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Запрос у БД имен записей
		rxData, err := c.conf.ActionsDB.ReadNamesTableBankCardContext(ctx)
		if err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: функция ReadNamesTableBankCardContext, вернула ошибку: <%v>", err))
		} else {
			c.data.namesBankCard = rxData // передача результата
			c.conf.LgrFile.Write("Debug: данные банковской карты, успешно прочитаны")
			c.status.readNameBankCardSUCCESS = true
		}
	}

	// Если режим - удалённый.
	if c.conf.Flag.Mode == flags.ModeRemote {

		c.status.readNameBankCardPassed = true // Установка признака, что был запущен процесс получения имён банковских карт.

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Запрос у сервера имен записей
		rxData, err := c.conf.Server.RequestBankCardNames(ctx, c.conf.Server.GetTokenAuthentication(), c.secret.secretKey)
		if err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: функция RequestTextNames, вернула ошибку: <%v>", err))
		} else {
			c.data.namesBankCard = rxData // передача результата
			c.conf.LgrFile.Write("Debug: данные банковской карты, успешно прочитаны")
			c.status.readNameBankCardSUCCESS = true
		}
	}

	return nil
}

// Установка фокуса. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
//	name - Имя элемента установки фокуса.
func layerShowBankCardSetFocus(c *handlerUI, g *gocui.Gui, name string) error {

	if _, err := g.SetCurrentView(name); err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}

	layoutInitialized = true
	c.view.currentFocus = name
	c.view.activeView = viewBankCardData

	// Установка стиля фокуса
	if v, err := g.View(name); err == nil {
		c.setFocusStyle(v, name)
	} else {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось получить вид для %s: <%v>", name, err))
		return fmt.Errorf("Не удалось получить вид для %s: <%w>", name, err)
	}

	// Принудительное обновление интерфейса
	g.Update(func(*gocui.Gui) error {
		return nil
	})

	return nil
}

//
// --- showBinary ---
//

// Ограничение на запуск функционала. Возвращается true - есть ограничение.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowBinaryRestraintRun(c *handlerUI) bool {

	if c.view.activeView != viewSelectType {
		return true
	}

	return false
}

// Создание экземпляра БД. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowBinaryNewInstDB(c *handlerUI) error {

	if c.conf.Flag.Mode == flags.ModeLocal {
		storage, err := domain.NewStorage(c.conf.Flag.DSN)
		if err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: Ошибка подключения к БД:<%v>", err))
			return fmt.Errorf("Функция NewStorage, вернула ошибку:<%w>", err)
		}
		c.conf.ActionsDB = storage
	}

	return nil
}

// Сброс переменных. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowBinaryReset(c *handlerUI) error {

	c.view.activeView = "" // Сброс
	c.index.file = 0

	c.updateStatusPopContainer(stageNotActive)
	c.updateStatusPushContainer(stageNotActive)
	c.updateStatusFileRx(stageNotActive)
	c.updateStatusFileTx(stageNotActive)

	c.status.readFilePassed = false // Для локального режима
	c.status.readFileSUCCESS = false

	c.status.readNameFilePassed = false // Для удалённого режима
	c.status.readNameFileSUCCESS = false

	c.status.delFilePassed = false
	c.status.delFileSUCCESS = false

	c.status.extractFilePassed = false
	c.status.extractFileSUCCESS = false

	layoutInitialized = false

	return nil
}

// Очистка экрана. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowBinaryClear(c *handlerUI, g *gocui.Gui) error {

	if err := deleteViews(g); err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}
	return nil
}

// Отрисовка окна. Возвращается указатель на окно и ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowBinaryWindow(c *handlerUI, g *gocui.Gui) (*gocui.View, error) {

	view, err := g.SetView(viewBinaryData, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return nil, fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	view.Title = "Файлы"
	view.Wrap = true
	view.Clear()

	return view, nil
}

// Отрисовка пояснений. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowBinaryDrawGuide(c *handlerUI, g *gocui.Gui, view *gocui.View) error {

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 3))
	fmt.Fprintf(view, "%sФайл в хранилище:\n", strings.Repeat(" ", 2))

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 3))
	fmt.Fprintf(view, "%sОткуда (файл):\n", strings.Repeat(" ", 2))

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 3))
	fmt.Fprintf(view, "%sКуда (директория):\n", strings.Repeat(" ", 2))

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 5))
	fmt.Fprintf(view, "%sПри внесении изменений, выполните Crl+U.\n", strings.Repeat(" ", 45))

	//
	// --- Нижняя часть экрана ---
	//

	// Верхний ряд.
	if v, err := g.SetView("Save", 1, 23, inputWidth-55, inputHeight+1+22); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+U - назад"))
	}

	return nil
}

// Отрисовка индикаторов. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowBinaryIndicator(c *handlerUI, g *gocui.Gui) error {

	// Результат чтения данных.
	indicatorY := 2
	indicatorX := 105
	vRead, err := g.SetView("indicatorReadStatus", indicatorX, indicatorY, indicatorX+20, indicatorY+2)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
	}
	vRead.Frame = false
	vRead.BgColor = gocui.ColorDefault
	vRead.FgColor = gocui.ColorDefault

	// Процент выполнения.
	if v, err := g.SetView("indicatorPercent", 56, 20, inputWidth+7, inputHeight+1+19); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = false
		v.Frame = false
		v.BgColor = gocui.ColorDefault
		v.SelBgColor = gocui.ColorDefault
		v.SelFgColor = gocui.ColorDefault
	}

	return nil
}

// Отрисовка полей вывода. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowBinaryDrawOut(c *handlerUI, g *gocui.Gui) error {

	// Отображение имени файла
	if v, err := g.SetView("fieldShowFor", 22, 2, inputWidth+20, inputHeight+1+1); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault

		v.Write([]byte("....."))
		c.setFocusStyle(v, "fieldShowFor")
	}

	return nil
}

// Отрисовка полей ввода. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowBinaryDrawInput(c *handlerUI, g *gocui.Gui) error {

	// Полный путь к файлу.
	if v, err := g.SetView("fieldPathSource", 22, 6, inputWidth+48, inputHeight+1+5); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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

	return nil
}

// Действия обработчика. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowBinaryActions(c *handlerUI) (err error) {

	// Получение списка названий файлов.
	c.status.readFilePassed = true // Установка признака, что был запущен процесс получения значений текста.

	// Если режим - локальный
	if c.conf.Flag.Mode == flags.ModeLocal {
		c.data.files, err = c.conf.Container.ListFilesInContainer(c.secret.secretKey)
		if err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: функция ListFilesInContainer, вернула ошибку: <%v>", err))
		} else {
			c.conf.LgrFile.Write("Debug: имена файлов в контейнере, успешно прочитаны")
			c.status.readFileSUCCESS = true
		}
	}

	// Если режим - удалённый
	if c.conf.Flag.Mode == flags.ModeRemote {

		c.status.readNameFilePassed = true // Установка признака, что был запущен процесс получения имён банковских карт.

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Запрос у сервера имен записей
		rxData, isBusyServer, err := c.conf.Server.RequestFileNames(ctx, c.conf.Server.GetTokenAuthentication(), c.secret.secretKey)
		if err != nil {
			c.conf.LgrFile.Write(fmt.Sprintf("Error: функция RequestFileNames, вернула ошибку: <%v>", err))
			c.status.readNameFileSUCCESS = false
		} else {
			if !isBusyServer {
				c.updateStatusIsBusyServer(false)
				c.data.namesFile = rxData
				c.conf.LgrFile.Write("Debug: данные файлов, успешно прочитаны")
				c.status.readNameFileSUCCESS = true
			} else { // На сервере активна работа с файлами.
				c.updateStatusIsBusyServer(true)
				c.data.namesFile = []string{}
				c.conf.LgrFile.Write("Debug: На сервере активна работа с файлами. Нет актуальных данных.")
			}
		}
	}

	return nil
}

// Установка фокуса. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
//	name - Имя элемента установки фокуса.
func layerShowBinarySetFocus(c *handlerUI, g *gocui.Gui, name string) error {

	if _, err := g.SetCurrentView(name); err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}

	layoutInitialized = true
	c.view.currentFocus = name
	c.view.activeView = viewBinaryData

	// Установка стиля фокуса
	if v, err := g.View(name); err == nil {
		c.setFocusStyle(v, name)
	} else {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось получить вид для %s: <%v>", name, err))
		return fmt.Errorf("Не удалось получить вид для %s: <%w>", name, err)
	}

	// Принудительное обновление интерфейса
	g.Update(func(*gocui.Gui) error {
		return nil
	})

	return nil
}

//
// --- showRequestEncryptKey ---
//

// Сброс переменных. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerShowRequestEncryptKeyReset(c *handlerUI) error {

	c.view.activeView = "" // Сброс признака активного окна

	layoutInitialized = false

	return nil
}

// Очистка экрана. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowRequestEncryptKeyClear(c *handlerUI, g *gocui.Gui) error {

	if err := deleteViews(g); err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("функция deleteViews, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция deleteViews, вернула ошибку: <%w>", err)
	}
	return nil
}

// Отрисовка окна. Возвращается указатель на окно и ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowRequestEncryptKeyWindow(c *handlerUI, g *gocui.Gui) (*gocui.View, error) {

	view, err := g.SetView(viewRequestSecretKey, 0, 0, screenWidth-1, screenHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return nil, fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}
	view.Title = "Секретный ключ"
	view.Wrap = true
	view.Clear()

	return view, nil
}

// Отрисовка полей ввода. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowRequestEncryptKeyDrawInput(c *handlerUI, g *gocui.Gui) error {

	if v, err := g.SetView("scrtKey", 50, 7, inputWidth+1, inputHeight+1+6); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
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

	return nil
}

// Отрисовка пояснений. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerShowRequestEncryptKeyDrawGuide(c *handlerUI, g *gocui.Gui, view *gocui.View) error {

	fmt.Fprintf(view, "%s", strings.Repeat("\n", 15))
	fmt.Fprintf(view, "%sОпционально. Укажите дополнительный ключ шифрования.\n", strings.Repeat(" ", 38))
	fmt.Fprintf(view, "%sИли оставьте поле пустым.", strings.Repeat(" ", 38))
	fmt.Fprintf(view, "%s", strings.Repeat("\n", 2))
	fmt.Fprintf(view, "%sНажмите на Enter.\n", strings.Repeat(" ", 38))

	if v, err := g.SetView("Enter", 1, 26, inputWidth-55, inputHeight+1+25); err != nil {
		if err != gocui.ErrUnknownView {
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
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
			c.conf.LgrFile.Write(fmt.Sprintf("функция SetView, вернула ошибку: <%v>", err))
			return fmt.Errorf("функция SetView, вернула ошибку: <%w>", err)
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Write([]byte("Ctrl+C - выход"))

	}
	return nil
}

// Установка фокуса. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
//	name - Имя элемента установки фокуса.
func layerShowRequestEncryptKeySetFocus(c *handlerUI, g *gocui.Gui, name string) error {

	if _, err := g.SetCurrentView(name); err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось установить фокус: <%v>", err))
		return fmt.Errorf("Не удалось установить фокус: <%w>", err)
	}

	layoutInitialized = true
	c.view.currentFocus = name
	c.view.activeView = viewRequestSecretKey

	// Установка стиля фокуса
	if v, err := g.View(name); err == nil {
		c.setFocusStyle(v, name)
	} else {
		c.conf.LgrFile.Write(fmt.Sprintf("Не удалось получить вид для %s: <%v>", name, err))
		return fmt.Errorf("Не удалось получить вид для %s: <%w>", name, err)
	}

	// Принудительное обновление интерфейса
	g.Update(func(*gocui.Gui) error {
		return nil
	})

	return nil
}

//
// --- doRestore ---
//

// Ограничение на запуск функционала. Возвращается true - есть ограничение.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerDoRestoreRestraintRun(c *handlerUI) bool {

	if c.getStatusBackUp() == stageActive ||
		c.getStatusRestore() == stageActive ||
		c.conf.Flag.Mode == flags.ModeRemote ||
		c.view.activeView != viewSelectType {
		return true
	}

	// Проверка, что установлен коннект с сервером.
	if !c.conf.Server.IsConnectSuccess() {
		c.conf.LgrFile.Write("Warn: Для выполнения процедуры резервного копирования, укажите данные сервера")
		c.updateStatusRestore(stageFault) // Установки признака ошибки процесса
		c.updateStatusBackUp(stageNotActive)
		return true
	}

	return false
}

// Закрытие подключения к БД. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerDoRestoreCloseDB(c *handlerUI) error {

	err := c.conf.ActionsDB.Close()
	if err != nil {
		return fmt.Errorf("функция ActionsDB.Close, вернула ошибку: <%w>", err)
	}

	return nil
}

// Действия обработчика. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerDoRestoreActions(c *handlerUI) (err error) {

	c.conf.LgrFile.Write("Info: Запущен поцесс Restore")

	c.updateStatusBackUp(stageNotActive) // сброс признака, чтобы убрать подсветку.
	c.updateStatusRestore(stageActive)

	// Предварительный запрос у сервера данных по файлам.
	rxFilesInfo, err := c.conf.Server.RestoreRequestFilesInfo()
	if err != nil {
		return fmt.Errorf("Функция RestoreRequestFilesInfo, вернула ошибку:<%w>", err)
	}

	// Проверка, что сервер предоставил данные по всем нужным файлам.
	wantNames := []string{c.conf.Flag.LocalNameDB, c.conf.Flag.LocalNameContainer}

	rxNames := make([]string, 0, 2)
	for _, f := range rxFilesInfo {
		rxNames = append(rxNames, f.Name)
	}
	if err := chechRxNameFiles(rxNames, wantNames); err != nil {
		return fmt.Errorf("Функция chechRxNameFiles, вернула ошибку:<%w>", err)
	}

	// Инициализация данных для выполнения Restore.
	if err := c.conf.Server.InitDataRestore(rxFilesInfo); err != nil {
		return fmt.Errorf("Функция InitDataRestore, вернула ошибку:<%w>", err)
	}

	chProcess := make(chan float32)
	chErr := make(chan error)
	chDone := make(chan struct{})

	// Приём файлов.
	go c.conf.Server.Restore(chProcess, chErr, chDone)

	// Буфер процесса.
	go bufferProcessRestore(c, chProcess, chErr, chDone)

	return nil
}

//
// --- doBackUp ---
//

// Ограничение на запуск функционала. Возвращается true - есть ограничение.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerDoBackUpRestraintRun(c *handlerUI) bool {

	if c.getStatusBackUp() == stageActive ||
		c.getStatusRestore() == stageActive ||
		c.conf.Flag.Mode == flags.ModeRemote ||
		c.view.activeView != viewSelectType {
		return true
	}

	// Проверка, что установлен коннект с сервером.
	if !c.conf.Server.IsConnectSuccess() {
		c.conf.LgrFile.Write("Warn: Для выполнения процедуры восстановления, укажите данные сервера")
		c.updateStatusBackUp(stageFault) // Установки признака ошибки процесса
		c.updateStatusRestore(stageNotActive)
		return true
	}

	return false
}

// Закрытие подключения к БД. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
func layerDoBackUpCloseDB(c *handlerUI) error {

	err := c.conf.ActionsDB.Close()
	if err != nil {
		return fmt.Errorf("функция ActionsDB.Close, вернула ошибку: <%w>", err)
	}

	return nil
}

// Действия обработчика. Возвращается ошибка.
//
// Параметры:
//
//	c - указатель на экземпляр сервиса.
//	g - указатель на Gui
func layerDoBackUpActions(c *handlerUI) (err error) {

	c.updateStatusBackUp(stageActive)
	c.updateStatusRestore(stageNotActive) // сброс признака, чтобы убрать подсветку.

	c.conf.LgrFile.Write("Info: Запущен процесс BackUp")

	// Логика процесса.
	//
	files := []string{c.conf.Flag.LocalNameDB, c.conf.Flag.LocalNameContainer}

	// Определение общего размера файлов.
	totalSizeKB, err := totalFileSize(files)
	if err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция totalFileSize, вернула ошибку: <%v>", err))
		c.updateStatusBackUp(stageFault)
		return nil
	}

	// Инициализация данных процесса BackUp.
	if err := c.conf.Server.InitDataBackUp(files, totalSizeKB, c.secret.secretKey); err != nil {
		c.conf.LgrFile.Write(fmt.Sprintf("Error: Функция InitDataBackUp, вернула ошибку: <%v>", err))
		c.updateStatusBackUp(stageFault)
		return nil
	}

	chProcess := make(chan float32)
	chErr := make(chan error)
	chDone := make(chan struct{})

	// Передача файлов.
	go c.conf.Server.BackUp(chProcess, chErr, chDone)

	// Буфер процесса BackUp.
	go bufferProcessBackUp(c, chProcess, chErr, chDone)

	return nil
}
