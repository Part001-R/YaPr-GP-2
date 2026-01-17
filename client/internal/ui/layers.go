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

	c.conf.LgrFile.Write(fmt.Sprintf("Info: Соединение с сервером установлено: <%v>", err))

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
		inst := container.New(containerName, c.secret.secretKey)
		c.conf.Container = inst

		c.conf.LgrFile.Write(fmt.Sprintf("Info: стартовая обработка контейнера <%s> пройдена", containerName))
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
			return nil
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
