package avoid

import (
	"fmt"
	"runtime"

	log "github.com/sirupsen/logrus"
)

// Error creates an error, and logs it righteously
func Error(message string) error {

	_, file, line, _ := runtime.Caller(1)

	log.WithFields(log.Fields{
		"caller": fmt.Sprintf("%s:%d", file, line),
	}).Error(message)

	return fmt.Errorf(message)

}

// ErrorE encapsulates err in a structured log and return an abstracted
// high-level error with message as the payload
func ErrorE(message string, err error) error {

	_, file, line, _ := runtime.Caller(1)

	log.WithFields(log.Fields{
		"error":  err,
		"caller": fmt.Sprintf("%s:%d", file, line),
	}).Error(message)

	return fmt.Errorf(message)

}

// ErrorF encapsulates fields in a structured log and return an abstracted
// high-level error with message as the payload
func ErrorF(message string, fields log.Fields) error {

	_, file, line, _ := runtime.Caller(1)

	log.WithFields(log.Fields{
		"caller": fmt.Sprintf("%s:%d", file, line),
	}).WithFields(fields).Error(message)

	return fmt.Errorf(message)

}

// ErrorEF encapsulates fields and err in a structured log and return an
// abstracted high-level error with message as the payload
func ErrorEF(message string, err error, fields log.Fields) error {

	_, file, line, _ := runtime.Caller(1)

	log.WithFields(log.Fields{
		"error":  err,
		"caller": fmt.Sprintf("%s:%d", file, line),
	}).WithFields(fields).Error(message)

	return fmt.Errorf(message)

}

// Warn acts like Error, but doesnt return anything
func Warn(message string) {

	_, file, line, _ := runtime.Caller(1)

	log.WithFields(log.Fields{
		"caller": fmt.Sprintf("%s:%d", file, line),
	}).Warn(message)
}

// WarnF includes an encapsulated field to the warning message
func WarnF(message string, fields log.Fields) {

	_, file, line, _ := runtime.Caller(1)

	log.WithFields(log.Fields{
		"caller": fmt.Sprintf("%s:%d", file, line),
	}).WithFields(fields).Warn(message)
}

// WarnF includes an encapsulated field to the warning message
func WarnEF(message string, err error, fields log.Fields) {

	_, file, line, _ := runtime.Caller(1)

	log.WithFields(log.Fields{
		"caller": fmt.Sprintf("%s:%d", file, line),
		"error":  err,
	}).WithFields(fields).Warn(message)
}

// WarnE includes an error message which is treated as not
func WarnE(message string, err error) {

	_, file, line, _ := runtime.Caller(1)

	log.WithFields(log.Fields{
		"error":  err,
		"caller": fmt.Sprintf("%s:%d", file, line),
	}).Warn(message)
}
