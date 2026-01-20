// Статические ошибки пакета.
package server

import "errors"

var (
	// В аргументе ip, нет содержимого
	NilPtrArgumentIP = errors.New("В аргументе ip, нет содержимого")

	// В аргументе port, нет содержимого
	NilPtrArgumentPort = errors.New("В аргументе port, нет содержимого")

	// В аргументе tokenStr, нет данных.
	MissingDataArgumentTokenStr = errors.New("в аргументе tokenStr, нет данных")

	// Длинна ключа не соответутсвует 32-м.
	LenSecretKey = errors.New("длина <secretKey> должна быть не менее 32 символов")

	// Неизвестный метод подписи.
	SigningMethodUnknown = errors.New("неизвестный метод подписи")

	// Время валидности токена истекло
	ValidTokenExpired = errors.New("время валидности токена истекло")

	// В аргументе subjectName, нет данных.
	EmptyDataArgumentSubjectName = errors.New("в аргументе <subjectName>, нет данных")

	// В аргументе userName, нет данных.
	EmptyDataArgumentUserName = errors.New("в аргументе <userName>, нет данных")

	// В аргументе userPwd, нет данных.
	EmptyDataArgumentUserPwd = errors.New("в аргументе <userPwd>, нет данных")

	// Нет указателя connect
	NilPtrConnect = errors.New("Нет указателя connect")

	// Отсутствует токен, отправленный серверу.
	MissingTokenData = errors.New("Отсутствует токен, отправленный серверу")

	// Отсутствуют данные токена от сервера.
	MissingTokenSrvData = errors.New("Отсутствуют данные токена от сервера")

	// В аргументе txID, нет данных.
	EmptyDataArgumentTxID = errors.New("в аргументе <txID>, нет данных")

	// В аргументе txFor, нет данных.
	EmptyDataArgumentTxFor = errors.New("в аргументе <txFor>, нет данных")

	// В аргументе txLogin, нет данных.
	EmptyDataArgumentTxLogin = errors.New("в аргументе <txLogin>, нет данных")

	// В аргументе txPassword, нет данных.
	EmptyDataArgumentTxPassword = errors.New("в аргументе <txPassword>, нет данных")

	// В аргументе txCreatedAt, нет данных.
	EmptyDataArgumentTxCreatedAt = errors.New("в аргументе <txCreatedAt>, нет данных")

	// В аргументе txOwner, нет данных.
	EmptyDataArgumentTxOwner = errors.New("в аргументе <txOwner>, нет данных")

	// В аргументе txNumb, нет данных.
	EmptyDataArgumentTxNumb = errors.New("в аргументе <txNumb>, нет данных")

	// В аргументе txValidData, нет данных.
	EmptyDataArgumentTxValidData = errors.New("в аргументе <txValidData>, нет данных")

	// В аргументе txCode, нет данных.
	EmptyDataArgumentTxCode = errors.New("в аргументе <txCode>, нет данных")

	// В аргументе txText, нет данных.
	EmptyDataArgumentTxText = errors.New("в аргументе <txText>, нет данных")

	// Некорректная длина зашифрованных данных
	NotCorrectLenData = errors.New("некорректная длина зашифрованных данных")

	// Некорректное заполнение
	NotCorrectDataFill = errors.New("некорректное заполнение")

	// Некорректное заполнение номера
	NotCorrectDataNumb = errors.New("некорректное заполнение номера")

	// Ошибка паддинга
	ErrInvalidPadding = errors.New("Ошибка паддинга")

	// В аргументе fileName, нет данных.
	EmptyDataArgumentFileName = errors.New("в аргументе <fileName>, нет данных")

	// В аргументе filePath, нет данных.
	EmptyDataArgumentFilePath = errors.New("в аргументе <filePath>, нет данных")

	// В аргументе tokenAuth, нет данных.
	EmptyDataArgumentTokenAuth = errors.New("в аргументе <tokenAuth>, нет данных")

	// В аргументе clientID, нет данных.
	EmptyDataArgumentClientID = errors.New("в аргументе <clientID>, нет данных")

	// Не удалось извлечь метаданные из контекста ответа
	ErrMetadata = errors.New("не удалось извлечь метаданные из контекста ответа")

	// В аргументе rxFileHash, нет данных.
	EmptyDataArgumentRxFileHash = errors.New("в аргументе <rxFileHash>, нет данных")

	// В аргументе srcFileHash, нет данных.
	EmptyDataArgumentSrcFileHash = errors.New("в аргументе <srcFileHash>, нет данных")

	// В аргументе rxToken, нет данных.
	EmptyDataArgumentRxToken = errors.New("в аргументе <rxToken>, нет данных")

	// В аргументе secretKey, нет данных.
	EmptyDataArgumentSecretKey = errors.New("в аргументе <secretKey>, нет данных")

	// В аргументе resp, нет данных.
	NilPtrArgumentResp = errors.New("в аргументе <resp>, нет данных")

	// В аргументе txFileName, нет данных.
	EmptyDataArgumentTxFileName = errors.New("в аргументе <txFileName>, нет данных")

	// В аргументе txFileHash, нет данных.
	EmptyDataArgumentTxFileHash = errors.New("в аргументе <txFileHash>, нет данных")

	// Некорректные данные в аргументе listFiles.
	NotCorrectLenLestFiles = errors.New("в аргументе <listFiles>, нет данных")

	// В аргументе sizeSendFile, нет данных.
	EmptyDataArgumentSizeSendFile = errors.New("в аргументе <sizeSendFile>, нет данных")

	// Подключение клиента закрыто.
	ErrConnectIsClosing = errors.New("rpc error: code = Canceled desc = grpc: the client connection is closing")

	// Недопустимое значение размера файла
	IncorrectSizeFile = errors.New("недопустимое значение размера файла")

	// недопустимое значение процентов процесса
	IncorrectSizePassed = errors.New("недопустимое значение процентов процесса")
)
