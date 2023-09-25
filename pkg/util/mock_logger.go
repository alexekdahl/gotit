package util

type mockLogger struct{}

func NewMockLogger() *mockLogger {
	return &mockLogger{}
}

func (l *mockLogger) Info(format string, v ...interface{})  {}
func (l *mockLogger) Warn(format string, v ...interface{})  {}
func (l *mockLogger) Error(err error)                       {}
func (l *mockLogger) Debug(format string, v ...interface{}) {}
func (l *mockLogger) Close() error {
	return nil
}
