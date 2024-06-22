package log

import (
    "os"
    "sync"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

var (
    logger     *zap.Logger
    once       sync.Once
    infoFile   *os.File
    errorFile  *os.File
    initLogger = initZapLogger
)

func initZapLogger() {
    var err error
    infoFile, err = os.OpenFile("logs/info.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        panic(err)
    }

    errorFile, err = os.OpenFile("logs/error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        panic(err)
    }

    infoWriteSyncer := zapcore.AddSync(infoFile)
    errorWriteSyncer := zapcore.AddSync(errorFile)

    encoderConfig := zap.NewProductionEncoderConfig()
    encoderConfig.TimeKey = "timestamp"
    encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

    infoCore := zapcore.NewCore(
        zapcore.NewJSONEncoder(encoderConfig),
        zapcore.NewMultiWriteSyncer(infoWriteSyncer, zapcore.AddSync(os.Stdout)),
        zapcore.InfoLevel,
    )

    errorCore := zapcore.NewCore(
        zapcore.NewJSONEncoder(encoderConfig),
        errorWriteSyncer,
        zapcore.ErrorLevel,
    )

    logger = zap.New(zapcore.NewTee(infoCore, errorCore), zap.AddCaller())
}

// Init initializes the logger. It should be called once.
func Init() {
    once.Do(initLogger)
}

// GetLogger returns the initialized logger instance.
func GetLogger() *zap.Logger {
    if logger == nil {
        Init()
    }
    return logger
}

// Cleanup closes the log files and syncs the logger.
func Cleanup() {
    if logger != nil {
        logger.Sync()
    }
    if infoFile != nil {
        infoFile.Close()
    }
    if errorFile != nil {
        errorFile.Close()
    }
}
