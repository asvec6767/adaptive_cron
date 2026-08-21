package log

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}

type LogStorage interface {
	Write(entry *LogEntry) error
	BatchWrite(entries []*LogEntry) error
	Query(query *LogQuery) ([]*LogEntry, error)
	Close() error
}
