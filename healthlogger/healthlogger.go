// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package healthlogger

import (
	"fmt"

	hlog "github.com/InVisionApp/go-logger"
	"github.com/go-kit/kit/log"
	"github.com/xmidt-org/webpa-common/v2/logging"
)

type HealthLogger struct {
	log.Logger
	keyValPairs []interface{}
}

type Option func(*HealthLogger)

func NewHealthLogger(logger log.Logger) hlog.Logger {
	h := HealthLogger{logger, []interface{}{}}

	return &h
}

func (h *HealthLogger) Debug(msg ...interface{}) {
	logging.Debug(h, h.keyValPairs...).Log(logging.MessageKey(), msg)
}

func (h *HealthLogger) Info(msg ...interface{}) {
	logging.Info(h, h.keyValPairs...).Log(logging.MessageKey(), msg)
}

func (h *HealthLogger) Warn(msg ...interface{}) {
	logging.Warn(h, h.keyValPairs...).Log(logging.MessageKey(), msg)
}

func (h *HealthLogger) Error(msg ...interface{}) {
	logging.Error(h, h.keyValPairs...).Log(logging.MessageKey(), msg)
}

func (h *HealthLogger) Debugln(msg ...interface{}) {
	logging.Debug(h, h.keyValPairs...).Log(logging.MessageKey(), msg)
}

func (h *HealthLogger) Infoln(msg ...interface{}) {
	logging.Info(h, h.keyValPairs...).Log(logging.MessageKey(), msg)
}

func (h *HealthLogger) Warnln(msg ...interface{}) {
	logging.Warn(h, h.keyValPairs...).Log(logging.MessageKey(), msg)
}

func (h *HealthLogger) Errorln(msg ...interface{}) {
	logging.Error(h, h.keyValPairs...).Log(logging.MessageKey(), msg)
}

func (h *HealthLogger) Debugf(format string, args ...interface{}) {
	logging.Debug(h, h.keyValPairs...).Log(logging.MessageKey(), fmt.Sprintf(format, args...))
}

func (h *HealthLogger) Infof(format string, args ...interface{}) {
	logging.Info(h, h.keyValPairs...).Log(logging.MessageKey(), fmt.Sprintf(format, args...))
}

func (h *HealthLogger) Warnf(format string, args ...interface{}) {
	logging.Warn(h, h.keyValPairs...).Log(logging.MessageKey(), fmt.Sprintf(format, args...))
}

func (h *HealthLogger) Errorf(format string, args ...interface{}) {
	logging.Error(h, h.keyValPairs...).Log(logging.MessageKey(), fmt.Sprintf(format, args...))
}

func (h *HealthLogger) WithFields(fields hlog.Fields) hlog.Logger {
	newKeyVals := h.keyValPairs
	for key, val := range fields {
		newKeyVals = append(newKeyVals, key, val)
	}
	return &HealthLogger{h, newKeyVals}
}
