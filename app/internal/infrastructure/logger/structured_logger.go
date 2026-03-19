package logger

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"ws-int-httr/internal/domain/log_domain"
	"ws-int-httr/internal/infrastructure/session"

	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *log.Logger
var Hostname string

// StructuredLogger es la interfaz para escribir logs estructurados
type StructuredLogger interface {
	LogAvail(log *log_domain.AvailLog)
	LogPreBook(log *log_domain.HotelResBookLog)
	LogBook(log *log_domain.HotelResCommitLog)
	LogCancel(log *log_domain.CancelLog)
	// LogCall es un método genérico que acepta cualquier tipo de log
	LogCall(log log_domain.GenericCallLog)
}

// FileStructuredLogger implementa StructuredLogger escribiendo a archivos JSON
type FileStructuredLogger struct {
	logPath       string
	availLogger   *lumberjack.Logger
	preBookLogger *lumberjack.Logger
	bookLogger    *lumberjack.Logger
	cancelLogger  *lumberjack.Logger
	mu            sync.Mutex
}

// NewFileStructuredLogger crea un nuevo logger que escribe a archivos JSON
// logPath: ruta base donde se crearán los archivos (ej: "/home/guester/ws-int-httr/log/")
func NewFileStructuredLogger(logPath string) (StructuredLogger, error) {
	// Crear directorio logstash si no existe
	logstashPath := filepath.Join(logPath, "logstash")
	if err := os.MkdirAll(logstashPath, 0755); err != nil {
		return nil, err
	}

	// Configurar loggers con lumberjack para rotación de archivos
	logger := &FileStructuredLogger{
		logPath: logPath,
		availLogger: &lumberjack.Logger{
			Filename:   filepath.Join(logstashPath, "HotelAvailJSON.log"),
			MaxSize:    5000, // MB (como en el proyecto original)
			MaxBackups: 7,
			MaxAge:     1, // días
			Compress:   true,
			LocalTime:  true,
		},
		preBookLogger: &lumberjack.Logger{
			Filename:   filepath.Join(logstashPath, "HotelResBookJSON.log"),
			MaxSize:    5000,
			MaxBackups: 7,
			MaxAge:     1,
			Compress:   true,
			LocalTime:  true,
		},
		bookLogger: &lumberjack.Logger{
			Filename:   filepath.Join(logstashPath, "HotelResCommitJSON.log"),
			MaxSize:    5000,
			MaxBackups: 7,
			MaxAge:     1,
			Compress:   true,
			LocalTime:  true,
		},
		cancelLogger: &lumberjack.Logger{
			Filename:   filepath.Join(logstashPath, "CancelJSON.log"),
			MaxSize:    5000,
			MaxBackups: 7,
			MaxAge:     1,
			Compress:   true,
			LocalTime:  true,
		},
	}

	// Establecer permisos de archivo
	if err := ensureFileWithPerms(logger.availLogger.Filename, 0655); err != nil {
		return nil, err
	}
	if err := ensureFileWithPerms(logger.preBookLogger.Filename, 0655); err != nil {
		return nil, err
	}
	if err := ensureFileWithPerms(logger.bookLogger.Filename, 0655); err != nil {
		return nil, err
	}
	if err := ensureFileWithPerms(logger.cancelLogger.Filename, 0655); err != nil {
		return nil, err
	}

	return logger, nil
}

func ensureFileWithPerms(filename string, mode os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, mode)
	if err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Chmod(filename, mode)
}

func (l *FileStructuredLogger) LogAvail(log *log_domain.AvailLog) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.availLogger.Write([]byte(log.ToJsonString() + "\n"))
}

func (l *FileStructuredLogger) LogPreBook(log *log_domain.HotelResBookLog) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.preBookLogger.Write([]byte(log.ToJsonString() + "\n"))
}

func (l *FileStructuredLogger) LogBook(log *log_domain.HotelResCommitLog) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.bookLogger.Write([]byte(log.ToJsonString() + "\n"))
}

func (l *FileStructuredLogger) LogCancel(log *log_domain.CancelLog) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cancelLogger.Write([]byte(log.ToJsonString() + "\n"))
}

// LogCall es un método genérico que acepta cualquier tipo de log estructurado
// y lo escribe al logger correspondiente según su tipo
func (l *FileStructuredLogger) LogCall(genLog log_domain.GenericCallLog) {
	switch genLog.CallType() {
	case "Avail":
		if availLog, ok := genLog.(*log_domain.AvailLog); ok {
			l.LogAvail(availLog)
		}
	case "HotelResBook":
		if preBookLog, ok := genLog.(*log_domain.HotelResBookLog); ok {
			l.LogPreBook(preBookLog)
		}
	case "HotelResCommit":
		if bookLog, ok := genLog.(*log_domain.HotelResCommitLog); ok {
			l.LogBook(bookLog)
		}
	case "Cancel":
		if cancelLog, ok := genLog.(*log_domain.CancelLog); ok {
			l.LogCancel(cancelLog)
		}
	}
}

// InitLog inicializa el sistema de logging
func InitLog(level string, path string) {
	var err error
	Hostname, err = os.Hostname()
	if err != nil {
		Hostname = "unknown"
	}

	Log = log.New(os.Stdout, "", log.LstdFlags)

	logFile := filepath.Join(path, "ws-int-httr_"+Hostname+".log")
	lumberjackLogger := &lumberjack.Logger{
		Filename: logFile,
		// MaxSize:    100,
		// MaxBackups: 7,
		MaxAge:    1,
		Compress:  true,
		LocalTime: true,
	}

	Log.SetOutput(io.MultiWriter(os.Stdout, lumberjackLogger))

	if err := ensureFileWithPerms(logFile, 0655); err != nil {
		Log.Printf("Failed to set file permissions: %v", err)
	} else {
		Log.Printf("Given permissions to %v", logFile)
	}

	Log.Printf("Log configured, level: %s", level)
}

func Infof(echoToken string, format string, v ...interface{}) {
	if Log != nil {
		msg := format
		if echoToken != "" {
			msg = "[" + echoToken + "] " + format
		}
		Log.Printf(msg, v...)
	}
}

func Errorf(echoToken string, format string, v ...interface{}) {
	if Log != nil {
		msg := format
		if echoToken != "" {
			msg = "[" + echoToken + "] " + format
		}
		Log.Printf("ERROR: "+msg, v...)
	}
}

func Infoln(echoToken string, v ...interface{}) {
	if Log != nil {
		if echoToken != "" {
			Log.Println(append([]interface{}{"[" + echoToken + "]"}, v...)...)
		} else {
			Log.Println(v...)
		}
	}
}

// Logger crea un middleware de logging para las peticiones HTTP
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Excluir endpoints de health del logging
		if path == "/ws-int-httr/health" || path == "/ws-int-httr/healthcheck" {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()
		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		dataLength := c.Writer.Size()
		if dataLength < 0 {
			dataLength = 0
		}

		// Obtener echoToken desde Gin context o sesión
		sessionCtx := session.FromContext()
		echoToken := ""
		if val, exists := c.Get("echoToken"); exists {
			if token, ok := val.(string); ok {
				echoToken = token
			}
		}
		if echoToken == "" {
			echoToken = sessionCtx.Data().EchoToken
		}

		// Intentar obtener requestType de la sesión si está disponible
		requestType := ""
		if requestTypeVal, ok := sessionCtx.Get("requestType"); ok {
			if rType, ok := requestTypeVal.(string); ok {
				requestType = rType
			}
		}
		requestType = normalizeRequestType(requestType)

		// Formato requerido: time="..." level=info EchoToken=... clientIP=... dataLength=...
		// latency=... method=... path=... requestType=... statusCode=...
		if Log != nil {
			_, _ = fmt.Fprintf(
				Log.Writer(),
				"time=%q level=info EchoToken=%s clientIP=%s dataLength=%d latency=%d method=%s path=%s requestType=%s statusCode=%d\n",
				time.Now().Format(time.RFC3339),
				echoToken,
				clientIP,
				dataLength,
				latency,
				c.Request.Method,
				path,
				requestType,
				statusCode,
			)
		}
	}
}

func normalizeRequestType(requestType string) string {
	switch requestType {
	case "GIOTAHotelAvailRQ":
		return "GIOTAHotelAvail"
	case "GIOTAHotelResRQ_Book":
		return "GIOTAHotelResBook"
	case "GIOTAHotelResRQ_COMMIT":
		return "GIOTAHotelResCommit"
	case "GIOTACancelRQ":
		return "GIOTACancel"
	default:
		return requestType
	}
}
