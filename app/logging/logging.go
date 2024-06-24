package logging

import "golang.org/x/exp/slog"

// Hacks to stop amplitude from force debug logging as info
// https://github.com/amplitude/analytics-go/blob/48ab76effd990c9446a75200c70060e815a8ac47/amplitude/loggers/default_logger.go#L20-L22

type AmplitudeLogger slog.Logger

func (l *AmplitudeLogger) Debugf(message string, args ...interface{}) {
	slog.Debug(message, args...)
}

func (l *AmplitudeLogger) Infof(message string, args ...interface{}) {
	slog.Info(message, args...)
}

func (l *AmplitudeLogger) Errorf(message string, args ...interface{}) {
	slog.Error(message, args...)
}

func (l *AmplitudeLogger) Warnf(message string, args ...interface{}) {
	slog.Warn(message, args...)
}

func NewHanayoLogger(h slog.Handler) *AmplitudeLogger {
	return (*AmplitudeLogger)(slog.New(h))
}
