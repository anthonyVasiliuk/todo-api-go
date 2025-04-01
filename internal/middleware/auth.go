package middleware

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strings"
	pb "todo-api/proto" // Импорт сгенерированного пакета

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var authServiceAddr = "localhost:8081"

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Требуется токен", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// Загружаем сертификат сервера
		certPEM, err := os.ReadFile("certs/cert.pem")
		if err != nil {
			http.Error(w, "Ошибка загрузки сертификата", http.StatusInternalServerError)
			return
		}

		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(certPEM) {
			http.Error(w, "Ошибка добавления сертификата в пул", http.StatusInternalServerError)
			return
		}

		// Настраиваем TLS
		creds := credentials.NewTLS(&tls.Config{
			RootCAs:    certPool,
			ServerName: "localhost", // Должно совпадать с CN в сертификате
			MinVersion: tls.VersionTLS12,
		})

		// Подключаемся к gRPC-серверу с TLS
		conn, err := grpc.Dial(authServiceAddr, grpc.WithTransportCredentials(creds))
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка подключения к auth-service: %v", err), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		client := pb.NewAuthServiceClient(conn)

		// Вызов VerifyToken
		resp, err := client.VerifyToken(context.Background(), &pb.TokenRequest{Token: tokenStr})
		if err != nil {
			http.Error(w, "Ошибка проверки токена", http.StatusUnauthorized)
			return
		}

		r.Header.Set("UserID", fmt.Sprintf("%d", resp.GetUserId()))
		r.Header.Set("Role", resp.GetRole())
		next(w, r)
	}
}
