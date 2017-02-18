package storage

type logger struct{}

func (l *logger) Fatal(m string, d ...interface{}) {
	innerlogger.Panicf(`%s-%s`, m, d)
}

func (l *logger) Error(m string, d ...interface{}) {
	innerlogger.Errorf(`%s-%s`, m, d)
}

func (l *logger) Warn(m string, d ...interface{}) {
	innerlogger.Warnf(`%s-%s`, m, d)
}

func (l *logger) Info(m string, d ...interface{}) {
	innerlogger.Infof(`%s-%s`, m, d)
}

func (l *logger) Debug(m string, d ...interface{}) {
	innerlogger.Debugf(`%s-%s`, m, d)
}

func (l *logger) Trace(m string, d ...interface{}) {
	innerlogger.Debugf(`%s-%s`, m, d)
}
