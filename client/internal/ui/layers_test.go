// Тесты пакета.
package ui

import (
	"os"
	"testing"

	"github.com/Part001-R/YaPr-GP-2/client/internal/adapters/server"
	"github.com/Part001-R/YaPr-GP-2/client/internal/service/udt"
	"github.com/Part001-R/YaPr-GP-2/client/internal/utils/flags"
	"github.com/Part001-R/YaPr-GP-2/client/internal/utils/logfile"
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

func TestLayerShowSelectTypeIsRestraintRun(t *testing.T) {

	// Данные тестов.
	dataTest := []struct {
		nameTest   string
		conf       *handlerUI
		wantResult bool
	}{
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
			nameTest: "Окно текст",
			conf: &handlerUI{
				view: screens{
					activeView: viewTextData,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно банковсой карты",
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
			nameTest: "Окно секретного ключа",
			conf: &handlerUI{
				view: screens{
					activeView: viewRequestSecretKey,
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

func TestLayerShowSelectTypeCloseDB(t *testing.T) {

	inst := &handlerUI{
		conf: &udt.Configuration{
			Flag: &flags.Config{
				DSN: "file:Foo.db?cache=shared&foreign_keys=on&mode=rwc",
			},
		},
	}

	// БД.
	err := layerShowAuthenticationIsNewInstDB(inst)
	require.NoErrorf(t, err, "Ошибка функции")
	assert.Truef(t, inst.conf.ActionsDB != nil, "Нет указателя")

	// Тест.
	err = layerShowSelectTypeCloseDB(inst)
	require.NoErrorf(t, err, "Ошибка функции")

	// Удаление данных.
	_ = os.Remove("Foo.db") // нет проверки err, т.к. при запуске go test -v ./..., формируется ошибка по отсутствию файла.
}

func TestLayerShowSelectTypeReset(t *testing.T) {

	layoutInitialized = true

	dataTest := &handlerUI{
		status: status{
			restore: stageActive,
			backUp:  stageActive,
		},
		view: screens{
			activeView: "Foo",
		},
	}

	// Тест.
	err := layerShowSelectTypeReset(dataTest)
	require.NoErrorf(t, err, "Ошибка функции")
	assert.Equalf(t, "", dataTest.view.activeView, "Нет сброса имени активного окна.")
	assert.Equalf(t, stageNotActive, dataTest.status.restore, "Нет сброса статуса restore.")
	assert.Equalf(t, stageNotActive, dataTest.status.backUp, "Нет сброса статуса backUp.")
	assert.Falsef(t, layoutInitialized, "Нет сброса инициализации")
}

func TestLayerShowSelectTypeNewInstContainer(t *testing.T) {

	nameContainer := "TestContainer"

	lgrFile, err := logfile.New("log.txt")
	require.NoErrorf(t, err, "Ошибка содания логгера")

	// Данные теста.
	inst := &handlerUI{
		conf: &udt.Configuration{
			LgrFile: lgrFile,
			Flag: &flags.Config{
				Mode:               flags.ModeLocal,
				LocalNameContainer: nameContainer,
			},
		},
		secret: encrKey{
			secretKey: [32]byte{},
		},
	}

	// Тест.
	err = layerShowSelectTypeNewInstContainer(inst)
	require.NoErrorf(t, err, "Ошибка содания контейнера")

	// Очистка.
	err = os.Remove(nameContainer)
	require.NoErrorf(t, err, "Ошибка удаления тестового контейнера")

	err = os.Remove("log.txt")
	require.NoErrorf(t, err, "Ошибка удаления файла логов")
}

// ----------------------------
//
//      showLoginPassword
//
// ----------------------------

func TestLayerShowLoginPasswordRestraintRun(t *testing.T) {

	// Данные тестов.
	dataTest := []struct {
		nameTest   string
		conf       *handlerUI
		wantResult bool
	}{
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
			nameTest: "Окно текст",
			conf: &handlerUI{
				view: screens{
					activeView: viewTextData,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно банковсой карты",
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
			nameTest: "Окно секретного ключа",
			conf: &handlerUI{
				view: screens{
					activeView: viewRequestSecretKey,
				},
			},
			wantResult: true,
		},
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
			nameTest: "Окно главное",
			conf: &handlerUI{
				view: screens{
					activeView: viewMain,
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
			nameTest: "Окно выбора типа данных",
			conf: &handlerUI{
				view: screens{
					activeView: viewSelectType,
				},
			},
			wantResult: false,
		},
	}

	// Тесты.
	for _, tt := range dataTest {
		t.Run(tt.nameTest, func(t *testing.T) {

			result := layerShowLoginPasswordRestraintRun(tt.conf)
			require.Equalf(t, tt.wantResult, result, "Нет соответствия при <%s>", tt.conf.view.activeView)
		})
	}
}

func TestLayerShowLoginPasswordNewInstDB(t *testing.T) {

	lgrFile, err := logfile.New("log.txt")
	require.NoErrorf(t, err, "Ошибка содания логгера")

	inst := &handlerUI{
		conf: &udt.Configuration{
			LgrFile: lgrFile,
			Flag: &flags.Config{
				DSN:  "file:Foo.db?cache=shared&foreign_keys=on&mode=rwc",
				Mode: flags.ModeLocal,
			},
		},
	}

	// Тест.
	err = layerShowLoginPasswordNewInstDB(inst)
	require.NoErrorf(t, err, "Ошибка функции")
	assert.Truef(t, inst.conf.ActionsDB != nil, "Нет указателя")

	// Удаление данных.
	_ = os.Remove("log.txt") // нет проверки err, т.к. при запуске go test -v ./..., формируется ошибка по отсутствию файла.
	_ = os.Remove("Foo.db")  // нет проверки err, т.к. при запуске go test -v ./..., формируется ошибка по отсутствию файла.
}

func TestLayerShowLoginPasswordReset(t *testing.T) {

	layoutInitialized = true

	dataTest := &handlerUI{
		view: screens{
			activeView: "Foo",
		},
		status: status{
			readLoginPaaswordPassed:     true,
			readNameLoginPaaswordPassed: true,
			addLoginPaaswordPassed:      true,
			delLoginPaaswordPassed:      true,
		},
		index: indexes{
			loginPassword: 1,
		},
	}

	// Тест.
	err := layerShowLoginPasswordReset(dataTest)
	require.NoErrorf(t, err, "Ошибка функции")
	assert.Equalf(t, "", dataTest.view.activeView, "Нет сброса имени активного окна.")
	assert.Falsef(t, dataTest.status.readLoginPaaswordPassed, "Нет сброса процесса чтения логин/пароль.")
	assert.Falsef(t, dataTest.status.readNameLoginPaaswordPassed, "Нет сброса процесса имён логин/пароль.")
	assert.Falsef(t, dataTest.status.addLoginPaaswordPassed, "Нет сброса процесса добавления логин/пароль.")
	assert.Falsef(t, dataTest.status.delLoginPaaswordPassed, "Нет сброса процесса удаления логин/пароль.")
	assert.Falsef(t, layoutInitialized, "Нет сброса инициализации")
}

// ----------------------------
//
//          showText
//
// ----------------------------

func TestLayerShowTextRestraintRun(t *testing.T) {

	// Данные тестов.
	dataTest := []struct {
		nameTest   string
		conf       *handlerUI
		wantResult bool
	}{
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
			nameTest: "Окно текст",
			conf: &handlerUI{
				view: screens{
					activeView: viewTextData,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно банковсой карты",
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
			nameTest: "Окно секретного ключа",
			conf: &handlerUI{
				view: screens{
					activeView: viewRequestSecretKey,
				},
			},
			wantResult: true,
		},
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
			nameTest: "Окно главное",
			conf: &handlerUI{
				view: screens{
					activeView: viewMain,
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
			nameTest: "Окно выбора типа данных",
			conf: &handlerUI{
				view: screens{
					activeView: viewSelectType,
				},
			},
			wantResult: false,
		},
	}

	// Тесты.
	for _, tt := range dataTest {
		t.Run(tt.nameTest, func(t *testing.T) {

			result := layerShowTextRestraintRun(tt.conf)
			require.Equalf(t, tt.wantResult, result, "Нет соответствия при <%s>", tt.conf.view.activeView)
		})
	}
}

func TestLayerShowTextNewInstDB(t *testing.T) {

	lgrFile, err := logfile.New("log.txt")
	require.NoErrorf(t, err, "Ошибка содания логгера")

	inst := &handlerUI{
		conf: &udt.Configuration{
			LgrFile: lgrFile,
			Flag: &flags.Config{
				DSN:  "file:Foo.db?cache=shared&foreign_keys=on&mode=rwc",
				Mode: flags.ModeLocal,
			},
		},
	}

	// Тест.
	err = layerShowTextNewInstDB(inst)
	require.NoErrorf(t, err, "Ошибка функции")
	assert.Truef(t, inst.conf.ActionsDB != nil, "Нет указателя")

	// Удаление данных.
	_ = os.Remove("log.txt") // нет проверки err, т.к. при запуске go test -v ./..., формируется ошибка по отсутствию файла.
	_ = os.Remove("Foo.db")  // нет проверки err, т.к. при запуске go test -v ./..., формируется ошибка по отсутствию файла.
}

func TestLayerShowTextReset(t *testing.T) {

	layoutInitialized = true

	dataTest := &handlerUI{
		status: status{
			readTextPassed:     true,
			readNameTextPassed: true,
			addTextPassed:      true,
			delTextPassed:      true,
			readTextSUCCESS:    true,
		},
		index: indexes{
			text: 1,
		},
		view: screens{
			activeView: "Foo",
		},
	}

	// Тест.
	err := layerShowTextReset(dataTest)
	require.NoErrorf(t, err, "Ошибка функции")
	assert.Equalf(t, "", dataTest.view.activeView, "Нет сброса activeView.")
	assert.Falsef(t, dataTest.status.readTextPassed, "Нет сброса readTextPassed.")
	assert.Falsef(t, dataTest.status.readNameTextPassed, "Нет сброса readNameTextPassed.")
	assert.Falsef(t, dataTest.status.addTextPassed, "Нет сброса addTextPassed.")
	assert.Falsef(t, dataTest.status.delTextPassed, "Нет сброса delTextPassed.")
	assert.Falsef(t, dataTest.status.readTextSUCCESS, "Нет сброса readTextSUCCESS.")
	assert.Falsef(t, layoutInitialized, "Нет сброса инициализации.")
}

// ----------------------------
//
//         showBankCard
//
// ----------------------------

func TestLayerShowBankCardRestraintRun(t *testing.T) {

	// Данные тестов.
	dataTest := []struct {
		nameTest   string
		conf       *handlerUI
		wantResult bool
	}{
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
			nameTest: "Окно текст",
			conf: &handlerUI{
				view: screens{
					activeView: viewTextData,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно банковсой карты",
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
			nameTest: "Окно секретного ключа",
			conf: &handlerUI{
				view: screens{
					activeView: viewRequestSecretKey,
				},
			},
			wantResult: true,
		},
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
			nameTest: "Окно главное",
			conf: &handlerUI{
				view: screens{
					activeView: viewMain,
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
			nameTest: "Окно выбора типа данных",
			conf: &handlerUI{
				view: screens{
					activeView: viewSelectType,
				},
			},
			wantResult: false,
		},
	}

	// Тесты.
	for _, tt := range dataTest {
		t.Run(tt.nameTest, func(t *testing.T) {

			result := layerShowBankCardRestraintRun(tt.conf)
			require.Equalf(t, tt.wantResult, result, "Нет соответствия при <%s>", tt.conf.view.activeView)
		})
	}
}

func TestLayerShowBankCardNewInstDB(t *testing.T) {

	lgrFile, err := logfile.New("log.txt")
	require.NoErrorf(t, err, "Ошибка содания логгера")

	inst := &handlerUI{
		conf: &udt.Configuration{
			LgrFile: lgrFile,
			Flag: &flags.Config{
				DSN:  "file:Foo.db?cache=shared&foreign_keys=on&mode=rwc",
				Mode: flags.ModeLocal,
			},
		},
	}

	// Тест.
	err = layerShowBankCardNewInstDB(inst)
	require.NoErrorf(t, err, "Ошибка функции")
	assert.Truef(t, inst.conf.ActionsDB != nil, "Нет указателя")

	// Удаление данных.
	_ = os.Remove("log.txt") // нет проверки err, т.к. при запуске go test -v ./..., формируется ошибка по отсутствию файла.
	_ = os.Remove("Foo.db")  // нет проверки err, т.к. при запуске go test -v ./..., формируется ошибка по отсутствию файла.
}

func TestLayerShowBankCardReset(t *testing.T) {

	layoutInitialized = true

	dataTest := &handlerUI{
		status: status{
			readBankCardPassed:      true,
			readNameBankCardPassed:  true,
			addBankCardPassed:       true,
			delBankCardPassed:       true,
			readBankCardSUCCESS:     true,
			readNameBankCardSUCCESS: true,
		},
		index: indexes{
			text: 1,
		},
		view: screens{
			activeView: "Foo",
		},
	}

	// Тест.
	err := layerShowBankCardReset(dataTest)
	require.NoErrorf(t, err, "Ошибка функции")
	assert.Equalf(t, "", dataTest.view.activeView, "Нет сброса activeView.")
	assert.Falsef(t, dataTest.status.readBankCardPassed, "Нет сброса readBankCardPassed.")
	assert.Falsef(t, dataTest.status.readNameBankCardPassed, "Нет сброса readNameBankCardPassed.")
	assert.Falsef(t, dataTest.status.addBankCardPassed, "Нет сброса addBankCardPassed.")
	assert.Falsef(t, dataTest.status.delBankCardPassed, "Нет сброса delBankCardPassed.")
	assert.Falsef(t, dataTest.status.readBankCardSUCCESS, "Нет сброса readBankCardSUCCESS.")
	assert.Falsef(t, dataTest.status.readNameBankCardSUCCESS, "Нет сброса readNameBankCardSUCCESS.")
	assert.Falsef(t, layoutInitialized, "Нет сброса инициализации.")
}

// ----------------------------
//
//          showBinary
//
// ----------------------------

func TestLayerShowBinaryRestraintRun(t *testing.T) {

	// Данные тестов.
	dataTest := []struct {
		nameTest   string
		conf       *handlerUI
		wantResult bool
	}{
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
			nameTest: "Окно текст",
			conf: &handlerUI{
				view: screens{
					activeView: viewTextData,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Окно банковсой карты",
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
			nameTest: "Окно секретного ключа",
			conf: &handlerUI{
				view: screens{
					activeView: viewRequestSecretKey,
				},
			},
			wantResult: true,
		},
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
			nameTest: "Окно главное",
			conf: &handlerUI{
				view: screens{
					activeView: viewMain,
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
			nameTest: "Окно выбора типа данных",
			conf: &handlerUI{
				view: screens{
					activeView: viewSelectType,
				},
			},
			wantResult: false,
		},
	}

	// Тесты.
	for _, tt := range dataTest {
		t.Run(tt.nameTest, func(t *testing.T) {

			result := layerShowBinaryRestraintRun(tt.conf)
			require.Equalf(t, tt.wantResult, result, "Нет соответствия при <%s>", tt.conf.view.activeView)
		})
	}
}

func TestLayerShowBinaryNewInstDB(t *testing.T) {

	lgrFile, err := logfile.New("log.txt")
	require.NoErrorf(t, err, "Ошибка содания логгера")

	inst := &handlerUI{
		conf: &udt.Configuration{
			LgrFile: lgrFile,
			Flag: &flags.Config{
				DSN:  "file:Foo.db?cache=shared&foreign_keys=on&mode=rwc",
				Mode: flags.ModeLocal,
			},
		},
	}

	// Тест.
	err = layerShowBinaryNewInstDB(inst)
	require.NoErrorf(t, err, "Ошибка функции")
	assert.Truef(t, inst.conf.ActionsDB != nil, "Нет указателя")

	// Удаление данных.
	_ = os.Remove("log.txt") // нет проверки err, т.к. при запуске go test -v ./..., формируется ошибка по отсутствию файла.
	_ = os.Remove("Foo.db")  // нет проверки err, т.к. при запуске go test -v ./..., формируется ошибка по отсутствию файла.
}

func TestLayerShowBinaryReset(t *testing.T) {

	layoutInitialized = true

	dataTest := &handlerUI{
		status: status{
			pushContainer:       stageActive,
			popContainer:        stageActive,
			fileTx:              stageActive,
			fileRx:              stageActive,
			readFilePassed:      true,
			readFileSUCCESS:     true,
			readNameFilePassed:  true,
			readNameFileSUCCESS: true,
			delFilePassed:       true,
			delFileSUCCESS:      true,
			extractFilePassed:   true,
			extractFileSUCCESS:  true,
		},
		index: indexes{
			file: 1,
		},
		view: screens{
			activeView: "Foo",
		},
	}

	// Тест.
	err := layerShowBinaryReset(dataTest)
	require.NoErrorf(t, err, "Ошибка функции")
	assert.Equalf(t, "", dataTest.view.activeView, "Нет сброса activeView.")
	assert.Equalf(t, 0, dataTest.index.file, "Нет сброса file.")
	assert.Equalf(t, stageNotActive, dataTest.status.pushContainer, "Нет сброса pushContainer.")
	assert.Equalf(t, stageNotActive, dataTest.status.popContainer, "Нет сброса popContainer.")
	assert.Equalf(t, stageNotActive, dataTest.status.fileTx, "Нет сброса fileTx.")
	assert.Equalf(t, stageNotActive, dataTest.status.fileRx, "Нет сброса fileRx.")
	assert.Falsef(t, dataTest.status.readFilePassed, "Нет сброса readFilePassed.")
	assert.Falsef(t, dataTest.status.readFileSUCCESS, "Нет сброса readFileSUCCESS.")
	assert.Falsef(t, dataTest.status.readNameFilePassed, "Нет сброса readNameFilePassed.")
	assert.Falsef(t, dataTest.status.readNameFileSUCCESS, "Нет сброса readNameFileSUCCESS.")
	assert.Falsef(t, dataTest.status.delFilePassed, "Нет сброса delFilePassed.")
	assert.Falsef(t, dataTest.status.delFileSUCCESS, "Нет сброса delFileSUCCESS.")
	assert.Falsef(t, dataTest.status.extractFilePassed, "Нет сброса extractFilePassed.")
	assert.Falsef(t, dataTest.status.extractFileSUCCESS, "Нет сброса extractFileSUCCESS.")
	assert.Falsef(t, layoutInitialized, "Нет сброса инициализации.")
}

// ----------------------------
//
//     showRequestEncryptKey
//
// ----------------------------

func TestLayerShowRequestEncryptKeyReset(t *testing.T) {

	// Данные тестов.
	dataTest := []struct {
		nameTest   string
		conf       *handlerUI
		wantResult string
	}{
		{
			nameTest: "Окно логин/пароль",
			conf: &handlerUI{
				view: screens{
					activeView: viewLoginPasswordData,
				},
			},
			wantResult: "",
		},
	}

	// Тесты.
	for _, tt := range dataTest {
		t.Run(tt.nameTest, func(t *testing.T) {

			err := layerShowRequestEncryptKeyReset(tt.conf)
			require.NoErrorf(t, err, "Ошибка функции")
			require.Equalf(t, tt.wantResult, tt.conf.view.activeView, "Нет соответствия activeView")
		})
	}
}

// ----------------------------
//
//          doRestore
//
// ----------------------------

func TestLayerDoRestoreRestraintRun(t *testing.T) {

	srv, err := server.New("localhost", "50001")
	require.NoErrorf(t, err, "ошибка создания экземпляра сервера")

	// Данные тестов.
	dataTest := []struct {
		nameTest   string
		conf       *handlerUI
		wantResult bool
	}{
		{
			nameTest: "Окно логин/пароль",
			conf: &handlerUI{
				conf: &udt.Configuration{
					Flag: &flags.Config{
						Mode: flags.ModeLocal,
					},
				},
				status: status{
					restore: stageNotActive,
					backUp:  stageNotActive,
				},
				view: screens{
					activeView: viewLoginPasswordData,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Идёт процесс backUp",
			conf: &handlerUI{
				conf: &udt.Configuration{
					Flag: &flags.Config{
						Mode: flags.ModeLocal,
					},
				},
				status: status{
					restore: stageNotActive,
					backUp:  stageActive,
				},
				view: screens{
					activeView: viewSelectType,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Идёт процесс restore",
			conf: &handlerUI{
				conf: &udt.Configuration{
					Flag: &flags.Config{
						Mode: flags.ModeLocal,
					},
				},
				status: status{
					restore: stageActive,
					backUp:  stageNotActive,
				},
				view: screens{
					activeView: viewSelectType,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Режим удалённый",
			conf: &handlerUI{
				conf: &udt.Configuration{
					Flag: &flags.Config{
						Mode: flags.ModeRemote,
					},
				},
				status: status{
					restore: stageNotActive,
					backUp:  stageNotActive,
				},
				view: screens{
					activeView: viewSelectType,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Режим локальный, без подключения.",
			conf: &handlerUI{
				conf: &udt.Configuration{
					Flag: &flags.Config{
						Mode: flags.ModeLocal,
					},
					Server: srv,
				},
				status: status{
					restore: stageNotActive,
					backUp:  stageNotActive,
				},
				view: screens{
					activeView: viewSelectType,
				},
			},
			wantResult: false,
		},
	}

	// Тесты.
	for _, tt := range dataTest {
		t.Run(tt.nameTest, func(t *testing.T) {

			result := layerDoRestoreRestraintRun(tt.conf)
			require.Equalf(t, tt.wantResult, result, "Нет соответствия при <%s>", tt.conf.view.activeView)
		})
	}
}

func TestLayerDoRestoreCloseDB(t *testing.T) {

	inst := &handlerUI{
		conf: &udt.Configuration{
			Flag: &flags.Config{
				DSN: "file:Foo.db?cache=shared&foreign_keys=on&mode=rwc",
			},
		},
	}

	// БД.
	err := layerShowAuthenticationIsNewInstDB(inst)
	require.NoErrorf(t, err, "Ошибка функции")
	assert.Truef(t, inst.conf.ActionsDB != nil, "Нет указателя")

	// Тест.
	err = layerDoRestoreCloseDB(inst)
	require.NoErrorf(t, err, "Ошибка функции")

	// Удаление данных.
	_ = os.Remove("Foo.db") // нет проверки err, т.к. при запуске go test -v ./..., формируется ошибка по отсутствию файла.
}

// ----------------------------
//
//          doBackUp
//
// ----------------------------

func TestLayerDoBackUpRestraintRun(t *testing.T) {

	srv, err := server.New("localhost", "50001")
	require.NoErrorf(t, err, "ошибка создания экземпляра сервера")

	// Данные тестов.
	dataTest := []struct {
		nameTest   string
		conf       *handlerUI
		wantResult bool
	}{
		{
			nameTest: "Окно логин/пароль",
			conf: &handlerUI{
				conf: &udt.Configuration{
					Flag: &flags.Config{
						Mode: flags.ModeLocal,
					},
				},
				status: status{
					restore: stageNotActive,
					backUp:  stageNotActive,
				},
				view: screens{
					activeView: viewLoginPasswordData,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Идёт процесс restore",
			conf: &handlerUI{
				conf: &udt.Configuration{
					Flag: &flags.Config{
						Mode: flags.ModeLocal,
					},
				},
				status: status{
					restore: stageActive,
					backUp:  stageNotActive,
				},
				view: screens{
					activeView: viewSelectType,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Идёт процесс backUp",
			conf: &handlerUI{
				conf: &udt.Configuration{
					Flag: &flags.Config{
						Mode: flags.ModeLocal,
					},
				},
				status: status{
					restore: stageNotActive,
					backUp:  stageActive,
				},
				view: screens{
					activeView: viewSelectType,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Режим удалённый",
			conf: &handlerUI{
				conf: &udt.Configuration{
					Flag: &flags.Config{
						Mode: flags.ModeRemote,
					},
				},
				status: status{
					restore: stageNotActive,
					backUp:  stageNotActive,
				},
				view: screens{
					activeView: viewSelectType,
				},
			},
			wantResult: true,
		},
		{
			nameTest: "Режим локальный, без подключения.",
			conf: &handlerUI{
				conf: &udt.Configuration{
					Flag: &flags.Config{
						Mode: flags.ModeLocal,
					},
					Server: srv,
				},
				status: status{
					restore: stageNotActive,
					backUp:  stageNotActive,
				},
				view: screens{
					activeView: viewSelectType,
				},
			},
			wantResult: false,
		},
	}

	// Тесты.
	for _, tt := range dataTest {
		t.Run(tt.nameTest, func(t *testing.T) {

			result := layerDoBackUpRestraintRun(tt.conf)
			require.Equalf(t, tt.wantResult, result, "Нет соответствия при <%s>", tt.conf.view.activeView)
		})
	}
}

func TestLayerDoBackUpCloseDB(t *testing.T) {

	inst := &handlerUI{
		conf: &udt.Configuration{
			Flag: &flags.Config{
				DSN: "file:Foo.db?cache=shared&foreign_keys=on&mode=rwc",
			},
		},
	}

	// БД.
	err := layerShowAuthenticationIsNewInstDB(inst)
	require.NoErrorf(t, err, "Ошибка функции")
	assert.Truef(t, inst.conf.ActionsDB != nil, "Нет указателя")

	// Тест.
	err = layerDoBackUpCloseDB(inst)
	require.NoErrorf(t, err, "Ошибка функции")

	// Удаление данных.
	_ = os.Remove("Foo.db") // нет проверки err, т.к. при запуске go test -v ./..., формируется ошибка по отсутствию файла.

}
