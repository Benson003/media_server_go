package logger

import (
	"go.uber.org/zap"
	"sync"
)

var (
	log *zap.Logger
	once sync.Once
)

func InitLogger(isDev bool){
	once.Do(func(){
		var err error
		if isDev{
			log,err = zap.NewDevelopment()
		} else {
			log,err = zap.NewProduction()
		}
		if err != nil{
			panic("Failed to initalise logger: "+ err.Error())
		}
	})
}

func Log() *zap.Logger{
	if log == nil {
		panic("Logger Not initaliseed. Call loggger.InitLogger()")
	}
	return log
}
