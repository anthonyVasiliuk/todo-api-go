package logger

import (
	"os"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func InitLogger() error {
	Log = logrus.New()

	// Устанавливаем уровень логирования
	Log.SetLevel(logrus.InfoLevel)

	// Формат логов
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Путь к файлу логов
	logFilePath := filepath.Join(".", "/logs", "app-%Y%m%d.log")

	// Настройка ротации
	rl, err := rotatelogs.New( // Используем fileRotatelogs как пакет, а не переменную
		logFilePath,
		rotatelogs.WithRotationTime(24*time.Hour), // Новый файл каждый день
		rotatelogs.WithRotationCount(7),           // Ограничение по количеству файлов
	)
	if err != nil {
		return err
	}

	// Подключаем ротацию к logrus через хук
	hook := lfshook.NewHook(
		lfshook.WriterMap{
			logrus.InfoLevel:  rl,
			logrus.WarnLevel:  rl,
			logrus.ErrorLevel: rl,
		},
		&logrus.TextFormatter{
			FullTimestamp: true,
		},
	)
	Log.AddHook(hook)

	// Вывод в консоль (опционально)
	Log.SetOutput(os.Stdout)

	Log.Info("Логгер с ротацией инициализирован")
	return nil
}
