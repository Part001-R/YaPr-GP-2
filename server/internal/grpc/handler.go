package grpc

// Интерсептор аворизации. Проверяется токен. Возвращается принятый запрос и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	req - запрос.
//	info -
//	handler - обработчик.
/*
func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	method := info.FullMethod

	if strings.Contains(method, "Authenticate") {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	tokens, exists := md["token"]
	if !exists || len(tokens) == 0 {
		return nil, status.Error(codes.Unauthenticated, "token not provided")
	}

	tokenString := tokens[0]
	if strings.TrimSpace(tokenString) == "" {
		return nil, status.Error(codes.Unauthenticated, "empty token")
	}

	// Проверка токена
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, status.Error(codes.Unauthenticated, "unexpected signing method")
		}
		return SecretKey, nil
	})

	if err != nil {
		// Уточняем причину ошибки при парсинге токена
		return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("error parsing token: %v", err))
	}

	// Проверка на истечение срока действия токена
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && claims.ExpiresAt != nil {
		if time.Now().After(claims.ExpiresAt.Time) {
			return nil, status.Error(codes.Unauthenticated, "token has expired")
		}
	}

	// Токен валиден — передаём управление дальше
	return handler(ctx, req)
}
*/
