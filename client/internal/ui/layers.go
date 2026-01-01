package ui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
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
func layerPrepareDataPingContext(c *handlerUI) (txMD metadata.MD, nameToken string, secretKey string, err error) {

	// Проверка аргументов.
	if c == nil {
		return nil, "", "", NilPtrArgumentC
	}

	// Подготовка данных.
	secretKey, err = generateRandomString(50)
	if err != nil {
		return nil, "", "", fmt.Errorf("функция generateRandomString, вернула ошибку: <%w>", err)
	}

	timeValidToken := time.Duration(5 * time.Second)
	txToken, err := createToken("clientManager", secretKey, timeValidToken)
	if err != nil {
		return nil, "", "", fmt.Errorf("функция createToken, вернула ошибку: <%w>", err)
	}

	nameToken = "token"
	txMD = metadata.Pairs(nameToken, txToken)

	// Результат.
	return txMD, nameToken, secretKey, nil
}

// Запрос.
func layerRequestPingContext(ctx context.Context, txMD metadata.MD, client proto.PasswordManagerClient, nameToken string, secretKey string) error {

	emptyRequest := &emptypb.Empty{}
	var header metadata.MD

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

	// Проверка метаданных ответа.
	if err := checkToken(rxToken, secretKey); err != nil {
		return fmt.Errorf("функция checkToken, вернула ошибку: <%w>", err)
	}

	return nil
}

//
// --- backUpDB ---
//

// Передача файла на сервер.
func layerTx(client proto.PasswordManagerClient, fileName string, c *handlerUI) (*pb.UploadResponse, error) {

	// Инициация стрима для загрузки файла
	stream, err := client.Upload(context.Background())
	if err != nil {
		return nil, fmt.Errorf("ошибка создания stream, для передачи данных: <%w>", err)
	}

	// Открытие файла для отправки
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла передачи: <%w>", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка при закрытии подключения к файлу: <%v>", err))
		}
	}()

	// Чтение файла по частям
	reader := bufio.NewReader(file)
	buf := make([]byte, 1024)

	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err != io.EOF {
				return nil, fmt.Errorf("ошибка при чтении файла: <%w>", err)
			}
			break
		}

		req := &pb.FileRequest{
			Filename: fileName,
			Content:  buf[:n],
		}

		if err := stream.Send(req); err != nil {
			return nil, fmt.Errorf("ошибка отправки данных файла: <%w>", err)
		}
	}

	// Закрытие потока передачи и ожидание ответа от сервера.
	resp, err := stream.CloseAndRecv()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения ответа от сервера: <%w>", err)
	}

	// Результат.
	return resp, nil
}

// Проверка ответа от сервера.
func layerCheckResultBackUp(resp *proto.UploadResponse, fileName string) error {

	// проверка аргументов.
	if fileName == "" {
		return EmptyDataArgumentFileName
	}

	// Проверка содержимого ответа сервера.
	if fileName != resp.Message {
		return NotConfirm
	}

	return nil
}
