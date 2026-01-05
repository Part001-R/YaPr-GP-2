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
// --- backUp ---
//

// Передача файла на сервер.
func layerBackUpTx(client proto.PasswordManagerClient, fileName, token string, c *handlerUI) (resp *pb.UploadResponse, rxHash, rxToken string, err error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Установка метаданных с токеном
	md := metadata.Pairs("token", token)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Инициация стрима для загрузки файла
	stream, err := client.BackupFile(ctx)
	if err != nil {
		return nil, "", "", fmt.Errorf("ошибка создания stream, для передачи данных: <%w>", err)
	}

	// Открытие файла для отправки
	file, err := os.Open(fileName)
	if err != nil {
		return nil, "", "", fmt.Errorf("ошибка открытия файла передачи: <%w>", err)
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
				return nil, "", "", fmt.Errorf("ошибка при чтении файла: <%w>", err)
			}
			break
		}

		req := &pb.UploadRequest{
			FileName: fileName,
			Content:  buf[:n],
		}

		if err := stream.Send(req); err != nil {
			return nil, "", "", fmt.Errorf("ошибка отправки данных файла: <%w>", err)
		}

		// Обновление статистики процесса.
		updateDataBackUpRestoreProcess(c, n)
	}

	// Закрытие потока передачи и ожидание ответа от сервера.
	resp, err = stream.CloseAndRecv()
	if err != nil {
		return nil, "", "", fmt.Errorf("ошибка получения ответа от сервера: <%w>", err)
	}

	// Получение трейлера хэша
	rxTrailer := stream.Trailer()
	if hash, ok := rxTrailer["hash"]; ok {
		rxHash = hash[0]
	} else {
		return nil, "", "", fmt.Errorf("Сервер не предоставил трейлер с данными хэша, для файла: <%s>", fileName)
	}

	if token, ok := rxTrailer["token"]; ok {
		rxToken = token[0]
	} else {
		return nil, "", "", fmt.Errorf("Сервер не предоставил трейлер с данными токена, для файла: <%s>", fileName)
	}

	// Результат.
	return resp, rxHash, rxToken, nil
}

// Проверка ответа от сервера.
func layerBackUpCheckResult(resp *proto.UploadResponse, fileName, fileHash, rxFileHash, token, rxToken string) error {

	// проверка аргументов.
	if fileName == "" {
		return EmptyDataArgumentFileName
	}

	// Проверка содержимого ответа сервера.
	respFileName := resp.FileName

	if fileName != respFileName {
		return fmt.Errorf("ошибка подтверждения сервером, для файла:<%s>. Ожидалось имя:<%s>, а принято:<%s>", fileName, fileName, respFileName)
	}
	if fileHash != rxFileHash {
		return fmt.Errorf("ошибка подтверждения сервером, для файла:<%s> Ожидался хэш:<%s>, а принято:<%s>", fileName, fileHash, rxFileHash)
	}
	if token != rxToken {
		return fmt.Errorf("ошибка подтверждения сервером, для файла:<%s>, нет соответствия токенов", fileName)
	}

	return nil
}

//
// --- restore ---
//

// Приём файла.
func layerRx(client proto.PasswordManagerClient, fileName, token string, c *handlerUI) (content []byte, rxFileHash, rxToken string, err error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Установка метаданных с токеном
	md := metadata.Pairs("token", token)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Запрос
	req := &pb.DownloadRequest{FileName: fileName}
	stream, err := client.RestoreFile(ctx, req)
	if err != nil {
		return nil, "", "", fmt.Errorf("Функция client.RestoreFile, вернула ошибку: <%w>", err)
	}

	// Чтение потоком.
	for {
		res, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, "", "", fmt.Errorf("Функция stream.Recv, вернула ошибку: <%w>", err)
		}
		if res.FileName != fileName {
			return nil, "", "", fmt.Errorf("Приняты данные для другого файла: <%s>", res.FileName)
		}

		// Обновление статистики процесса.
		updateDataBackUpRestoreProcess(c, len(res.Content))

		content = append(content, res.Content...)
	}

	// Получение трейлера после завершения потока
	rxTrailer := stream.Trailer()
	if hash, ok := rxTrailer["hash"]; ok {
		rxFileHash = hash[0]
	} else {
		return nil, "", "", fmt.Errorf("Сервер не предоставил трейлер с данными хэша, для файла: <%s>", fileName)
	}
	if token, ok := rxTrailer["token"]; ok {
		rxToken = token[0]
	} else {
		return nil, "", "", fmt.Errorf("Сервер не предоставил трейлер с данными токена, для файла: <%s>", fileName)
	}

	// Результат
	return content, rxFileHash, rxToken, nil
}

// Сохранение файла.
func saveFile(content []byte, fileName string) error {

	if err := os.WriteFile(fileName, content, 0644); err != nil {
		return fmt.Errorf("Функция os.WriteFile, вернула ошибку: <%v>", err)
	}

	return nil
}

// Проверка ответа от сервера.
func layerRestoreCheckResult(fileName, fileHash, rxFileHash, token, rxToken string) error {

	// Анализ данных ответа от сервера.
	fileHash, err := hashFile(fileName)
	if err != nil {
		return fmt.Errorf("Функция hashFile, вернула ошибку:<%w>", err)
	}
	if rxFileHash != fileHash {
		return fmt.Errorf("Для файла:<%s>, нет соответствия хэша. Ожидался:<%s>, а принято:<%s>", fileName, rxFileHash, fileHash)
	}
	if token != rxToken {
		return fmt.Errorf("Для файла:<%s>, нет соответствия токенов", fileName)
	}

	return nil
}

//
// --- FilesInfo ---
//

func layerFilesInfoRequest(client proto.PasswordManagerClient, token string) (files []infoByFiles, rxToken string, err error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Установка метаданных с токеном
	md := metadata.Pairs("token", token)
	ctx = metadata.NewOutgoingContext(ctx, md)
	var trailer metadata.MD

	// Запрос у сервера информации по файлам.
	emptyRequest := &emptypb.Empty{}
	infoResp, err := client.FilesInfo(
		ctx,
		emptyRequest,
		grpc.Trailer(&trailer),
	)
	if err != nil {
		return nil, "", fmt.Errorf("Функция client.FilesInfo, вернула ошибку: <%v>", err)
	}

	// Получение трейлера
	if trailer != nil {
		val, ok := trailer["token"]
		if !ok {
			return nil, "", ErrMetadata
		}
		rxToken = val[0]
	}

	// Обработка результата запроса.
	for _, f := range infoResp.FileInfo {
		var el infoByFiles

		el.name = f.FileName
		el.volume = f.Size

		files = append(files, el)
	}

	// Результат.
	return files, rxToken, nil
}

// Заполнение данных по ожидаемому объёму приема.
func layerFilesInfoFillData(files []infoByFiles, c *handlerUI) error {

	// Обновление данных.
	for _, f := range files {

		if f.volume < 0 {
			return IncorrectData
		}

		c.txrx.totalSizeKB += f.volume / 1024
	}

	return nil
}
