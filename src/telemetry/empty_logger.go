package telemetry

type EmptyLogger struct{}

func (e *EmptyLogger) Debugf(format string, args ...interface{}) {}

func (e *EmptyLogger) Infof(format string, args ...interface{}) {}

func (e *EmptyLogger) Warnf(format string, args ...interface{}) {}

func (e *EmptyLogger) Errorf(format string, args ...interface{}) {}
