package ui

import (
	"os"
	"testing"

	"github.com/Part001-R/YaPr-GP-2/client/internal/service/udt"
	"github.com/Part001-R/YaPr-GP-2/client/internal/utils/flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ----------------------------
//
//        pingContext
//
// ----------------------------

func TestLayerDataPingContextPrepare_SUCCESS(t *testing.T) {

	conf := &handlerUI{
		conf:       nil,
		typed:      typeData{},
		status:     status{},
		view:       screens{},
		secret:     encrKey{},
		data:       data{},
		index:      indexes{},
		mutex:      mutex{},
		txrx:       txrx{},
		clientName: "",
		tokenAuth:  "",
		ch:         ch{},
	}

	// Тест.
	txMD, nameToken, secrKey, err := layerDataPingContextPrepare(conf)
	require.NoErrorf(t, err, "Ошибка layerDataPingContextPrepare")

	// Проверки.
	require.NotEmpty(t, secrKey, "Ожидали не пустой secrKey")
	require.Equal(t, "token", nameToken, "Ожидали nameToken равным 'token'")

	tokenValues := txMD.Get(nameToken)
	require.Equalf(t, 1, len(tokenValues), "Нет соответствия длинны", nameToken)

	token := tokenValues[0]
	require.NotEmpty(t, token, "Пустой токен")
}

func TestLayerDataPingContextPrepare_FAULT(t *testing.T) {

	// Тест.
	_, _, _, err := layerDataPingContextPrepare(nil)
	require.Equalf(t, "в аргументе <c> нет указателя", err.Error(), "Нет соответствия ошибки.")

}

// ----------------------------
//
//      showRegistration
//
// ----------------------------

func TestLayerShowRegistrationIsRestraintRun(t *testing.T) {

	// Данные тестов.
	dataTest := []struct {
		nameTest   string
		conf       *handlerUI
		wantResult bool
	}{
		{
			nameTest: "Окно аутентификации",
			conf: &handlerUI{
				view: screens{
					activeView: viewAutentification,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно банковской карты",
			conf: &handlerUI{
				view: screens{
					activeView: viewBankCardData,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно файлов",
			conf: &handlerUI{
				view: screens{
					activeView: viewBinaryData,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно логин/пароль",
			conf: &handlerUI{
				view: screens{
					activeView: viewLoginPasswordData,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно регистрации",
			conf: &handlerUI{
				view: screens{
					activeView: viewRegistration,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно запроса секретного ключа",
			conf: &handlerUI{
				view: screens{
					activeView: viewRequestSecretKey,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно выбора типа",
			conf: &handlerUI{
				view: screens{
					activeView: viewSelectType,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно настроек",
			conf: &handlerUI{
				view: screens{
					activeView: viewSettings,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно текстовых данных",
			conf: &handlerUI{
				view: screens{
					activeView: viewTextData,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Без имени",
			conf: &handlerUI{
				view: screens{
					activeView: "",
				},
			},
			wantResult: false,
		},
	}

	// Тесты.
	for _, tt := range dataTest {
		t.Run(tt.nameTest, func(t *testing.T) {

			result := layerShowRegistrationIsRestraintRun(tt.conf)
			require.Equalf(t, tt.wantResult, result, "Нет соответствия при <%s>", tt.conf.view.activeView)
		})
	}
}

func TestLayerShowRegistrationReset(t *testing.T) {

	layoutInitialized = true

	dataTest := &handlerUI{
		typed: typeData{
			login:     "Foo",
			password1: "Bar",
			password2: "Bar",
		},
		status: status{
			addUserSUCCESS: true,
			addUserPassed:  true,
		},
		view: screens{
			activeView: "Foo",
		},
	}

	// Тест.
	err := layerShowRegistrationReset(dataTest)
	require.NoErrorf(t, err, "Ошибка функции")
	assert.Equalf(t, "", dataTest.view.activeView, "Нет сброса имени активного окна.")
	assert.Equalf(t, "", dataTest.typed.login, "Нет сброса логина.")
	assert.Equalf(t, "", dataTest.typed.password1, "Нет сброса пароля 1.")
	assert.Equalf(t, "", dataTest.typed.password2, "Нет сброса пароля 2.")
	assert.Falsef(t, dataTest.status.addUserPassed, "Нет сброса Passed")
	assert.Falsef(t, dataTest.status.addUserPassed, "Нет сброса SUCCESS")
	assert.Falsef(t, layoutInitialized, "Нет сброса инициализации")
}

// ----------------------------
//
//     showAuthentication
//
// ----------------------------

func TestLayerShowAuthenticationIsRestraintRun(t *testing.T) {

	// Данные тестов.
	dataTest := []struct {
		nameTest   string
		conf       *handlerUI
		wantResult bool
	}{
		{
			nameTest: "Окно аутентификации",
			conf: &handlerUI{
				view: screens{
					activeView: viewAutentification,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно банковской карты",
			conf: &handlerUI{
				view: screens{
					activeView: viewBankCardData,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно файлов",
			conf: &handlerUI{
				view: screens{
					activeView: viewBinaryData,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно логин/пароль",
			conf: &handlerUI{
				view: screens{
					activeView: viewLoginPasswordData,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно регистрации",
			conf: &handlerUI{
				view: screens{
					activeView: viewRegistration,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно секретного ключа",
			conf: &handlerUI{
				view: screens{
					activeView: viewRequestSecretKey,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно выбора типа",
			conf: &handlerUI{
				view: screens{
					activeView: viewSelectType,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно настроект",
			conf: &handlerUI{
				view: screens{
					activeView: viewSettings,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно текста",
			conf: &handlerUI{
				view: screens{
					activeView: viewTextData,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Без имени",
			conf: &handlerUI{
				view: screens{
					activeView: "",
				},
			},
			wantResult: false,
		},
	}

	// Тесты.
	for _, tt := range dataTest {
		t.Run(tt.nameTest, func(t *testing.T) {

			result := layerShowAuthenticationIsRestraintRun(tt.conf)
			require.Equalf(t, tt.wantResult, result, "Нет соответствия при <%s>", tt.conf.view.activeView)
		})
	}
}

func TestLayerShowAuthenticationIsNewInstDB(t *testing.T) {

	inst := &handlerUI{
		conf: &udt.Configuration{
			Flag: &flags.Config{
				DSN: "file:Foo.db?cache=shared&foreign_keys=on&mode=rwc",
			},
		},
	}

	// Тест.
	err := layerShowAuthenticationIsNewInstDB(inst)
	require.NoErrorf(t, err, "Ошибка функции")
	assert.Truef(t, inst.conf.ActionsDB != nil, "Нет указателя")

	// Удаление данных.
	err = os.Remove("Foo.db")
	require.NoErrorf(t, err, "Ошибка удаления БД")
}

func TestLayerShowAuthenticationReset(t *testing.T) {

	layoutInitialized = true

	dataTest := &handlerUI{
		typed: typeData{
			login:     "Foo",
			password1: "Bar",
			password2: "Bar",
		},
		view: screens{
			activeView: "Foo",
		},
	}

	// Тест.
	err := layerShowAuthenticationReset(dataTest)
	require.NoErrorf(t, err, "Ошибка функции")
	assert.Equalf(t, "", dataTest.view.activeView, "Нет сброса имени активного окна.")
	assert.Equalf(t, "", dataTest.typed.login, "Нет сброса логина.")
	assert.Equalf(t, "", dataTest.typed.password1, "Нет сброса пароля 1.")
	assert.Equalf(t, "", dataTest.typed.password2, "Нет сброса пароля 2.")
	assert.Falsef(t, layoutInitialized, "Нет сброса инициализации")
}

// ----------------------------
//
//         showSettings
//
// ----------------------------

func TestLayerShowSettingsIsRestraintRun(t *testing.T) {

	// Данные тестов.
	dataTest := []struct {
		nameTest   string
		conf       *handlerUI
		wantResult bool
	}{
		{
			nameTest: "Окно аутентификации",
			conf: &handlerUI{
				view: screens{
					activeView: viewAutentification,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно банковской карты",
			conf: &handlerUI{
				view: screens{
					activeView: viewBankCardData,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно файлов",
			conf: &handlerUI{
				view: screens{
					activeView: viewBinaryData,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно логин/пароль",
			conf: &handlerUI{
				view: screens{
					activeView: viewLoginPasswordData,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно регистрации",
			conf: &handlerUI{
				view: screens{
					activeView: viewRegistration,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно секретного ключа",
			conf: &handlerUI{
				view: screens{
					activeView: viewRequestSecretKey,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно выбора типа",
			conf: &handlerUI{
				view: screens{
					activeView: viewSelectType,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно настроект",
			conf: &handlerUI{
				view: screens{
					activeView: viewSettings,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно текста",
			conf: &handlerUI{
				view: screens{
					activeView: viewTextData,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Без имени",
			conf: &handlerUI{
				view: screens{
					activeView: "",
				},
			},
			wantResult: false,
		},
	}

	// Тесты.
	for _, tt := range dataTest {
		t.Run(tt.nameTest, func(t *testing.T) {

			result := layerShowSettingsIsRestraintRun(tt.conf)
			require.Equalf(t, tt.wantResult, result, "Нет соответствия при <%s>", tt.conf.view.activeView)
		})
	}
}

func TestLayerShowSettingsReset(t *testing.T) {

	layoutInitialized = true

	dataTest := &handlerUI{
		view: screens{
			activeView: "Foo",
		},
	}

	// Тест.
	err := layerShowSettingsReset(dataTest)
	require.NoErrorf(t, err, "Ошибка функции")
	assert.Equalf(t, "", dataTest.view.activeView, "Нет сброса имени активного окна.")
	assert.Falsef(t, layoutInitialized, "Нет сброса инициализации")
}

// ----------------------------
//
//          showMain
//
// ----------------------------

func TestLayerShowMainIsRestraintRun(t *testing.T) {

	// Данные тестов.
	dataTest := []struct {
		nameTest   string
		conf       *handlerUI
		wantResult bool
	}{
		{
			nameTest: "Есть активность",
			conf: &handlerUI{
				status: status{
					restore: stageActive,
					backUp:  stageActive,
				},
			},
			wantResult: true,
		},
	}

	// Тесты.
	for _, tt := range dataTest {
		t.Run(tt.nameTest, func(t *testing.T) {

			result := layerShowMainIsRestraintRun(tt.conf)
			require.Equalf(t, tt.wantResult, result, "Нет соответствия")
		})
	}
}

func TestLayerShowMainReset(t *testing.T) {

	layoutInitialized = true

	dataTest := &handlerUI{
		view: screens{
			activeView:   "Foo",
			currentFocus: "Bar",
		},
	}

	// Тест.
	err := layerShowMainReset(dataTest)
	require.NoErrorf(t, err, "Ошибка функции")
	assert.Equalf(t, "", dataTest.view.activeView, "Нет сброса имени активного окна.")
	assert.Equalf(t, "", dataTest.view.currentFocus, "Нет сброса фокуса.")
	assert.Falsef(t, layoutInitialized, "Нет сброса инициализации")
}

// ----------------------------
//
//        showSelectType
//
// ----------------------------
