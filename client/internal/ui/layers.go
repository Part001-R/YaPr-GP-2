package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/Part001-R/YaPr-GP-2/proto"
	pb "github.com/Part001-R/YaPr-GP-2/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Подключение к серверу.
func connectSrv(c *handlerUI) (pb.PasswordManagerClient, *grpc.ClientConn, error) {

	// Проверка аргументов
	if c == nil {
		return nil, nil, NilPtrArgumentC
	}

	// Логика
	//
	c.status.checkConnectPassed = true
	srvAddr := c.typed.ip + ":" + c.typed.port

	// Настройка TLS.
	creds, err := credentials.NewClientTLSFromFile("tls/server.crt", "")
	if err != nil {
		return nil, nil, fmt.Errorf("функция credentials.NewClientTLSFromFile, вернула ошибку: <%w>", err)
	}

	// Подключение к серверу.
	conn, err := grpc.NewClient(srvAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, nil, fmt.Errorf("функция grpc.NewClient, вернула ошибку: <%w>", err)
	}

	// создание клиента.
	client := pb.NewPasswordManagerClient(conn)

	return client, conn, nil
}

//
// --- pingContext ---
//

// Додготовка данных для запроса.
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

// Запрос.
func layerPingContextRequest(ctx context.Context, txMD metadata.MD, client proto.PasswordManagerClient, nameToken string, secretKey string) error {

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
	if err := checkToken(rxToken, secretKey); err != nil {
		return fmt.Errorf("функция checkToken, вернула ошибку: <%w>", err)
	}

	return nil
}
