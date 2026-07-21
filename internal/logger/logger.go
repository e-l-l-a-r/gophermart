package logger

import (
	"bytes"
	"fmt"
	"io"
	console "log"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	// Структура для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// Оберточная структура для http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // оригинальный http.ResponseWriter
		responseData        *responseData
	}

	// Структура самого логгера
	logger struct {
		zap.Logger
	}
)

// Глобальная переменная для реализации работы логгера - синглтона
var singleLogger *logger

func GetLogger() (*logger, error) {
	if singleLogger == nil {
		return nil, fmt.Errorf("no logger inited")
	}
	return singleLogger, nil
}

func InitLogger(level string) (*logger, error) {
	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}
	// создаём новую конфигурацию логера
	cfg := zap.NewProductionConfig()
	// устанавливаем уровень
	cfg.Level = lvl
	cfg.Encoding = "console"
	cfg.EncoderConfig.ConsoleSeparator = "\t| "
	// формат времени
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // или RFC3339TimeEncoder, EpochTimeEncoder
	// формат уровня логирования
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	// настраиваем формат caller
	cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	// создаём логер на основе конфигурации
	log, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}
	singleLogger = &logger{
		*log,
	}
	return GetLogger()
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

func (l *logger) LogHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}
		next.ServeHTTP(&lw, r) // внедряем реализацию http.ResponseWriter

		duration := time.Since(start)

		l.Sugar().Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status, // получаем перехваченный код статуса ответа
			"duration", duration,
			"size", responseData.size, // получаем перехваченный размер ответа
		)
	})
}

func (l *logger) DoRequestWithLog(c *retryablehttp.Client, req *retryablehttp.Request) (resp *http.Response, err error) {
	resp, err = c.Do(req)
	if err != nil {
		l.WarnMsg(err)
	} else {
		bodyBytes, _ := io.ReadAll(resp.Body)
		l.InfoMsg("url:", req.URL, "\tstatus code:", resp.StatusCode, "\n\tBody", string(bodyBytes))
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("status code: %d", resp.StatusCode)
		}
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	return
}

func (l *logger) InfoMsg(args ...interface{}) {
	if l != nil {
		l.Sugar().Infoln(args)
	} else {
		console.Println(args...)
	}
}

func (l *logger) WarnMsg(args ...interface{}) {
	if l != nil {
		l.Sugar().Warnln(args)
	} else {
		console.Println(args...)
	}
}

func Fatal(args ...interface{}) {
	if singleLogger != nil {
		singleLogger.Sugar().Fatalln(args)
	} else {
		console.Fatal(args...)
	}
}

func Err(args ...interface{}) {
	if singleLogger != nil {
		singleLogger.Sugar().Errorln(args)
	} else {
		console.Println(args...)
	}
}

func Warn(args ...interface{}) {
	if singleLogger != nil {
		singleLogger.Sugar().Warnln(args)
	} else {
		console.Println(args...)
	}
}

func Info(args ...interface{}) {
	if singleLogger != nil {
		singleLogger.Sugar().Infoln(args)
	} else {
		console.Println(args...)
	}
}

func Debug(args ...interface{}) {
	if singleLogger != nil {
		singleLogger.Sugar().Debugln(args)
	} else {
		console.Println(args...)
	}
}
