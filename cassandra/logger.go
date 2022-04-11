package cassandra

import hlog "github.com/InVisionApp/go-logger"

// gocql addapter logger to go kit loger
type goCqlLogger struct {
	Logger hlog.Logger
}

func (g *goCqlLogger) Print(v ...interface{}) {
	g.Logger.Warn(v...)
}

func (g *goCqlLogger) Printf(format string, v ...interface{}) {
	g.Logger.Warnf(format, v...)
}

func (g *goCqlLogger) Println(v ...interface{}) {
	g.Logger.Warnln(v...)
}
