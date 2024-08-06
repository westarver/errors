package errors

import (
	"fmt"
	"log"
	"strings"

	pkgerr "github.com/pkg/errors"
)

const (
	LogErr = "Log and return "
	Panic  = "Log and panic "
	Fail   = "Log and exit "
)

type LogFunc func(elem ...any)
type PrintFunc func(elem ...any)

var ErrLogger = errlog
var FailLogger = faillog
var ErrPrinter = print

// the errlog function is the default function used to log errors
// NOTE:
// the first arg passed will be used for the log prefix
var errlog LogFunc = func(elem ...any) {
	if len(elem) == 0 {
		return
	}

	log.SetPrefix(elem[0].(string))
	log.SetFlags(log.Ldate | log.Ltime)
	if len(elem) > 1 {
		log.Println(elem[1:]...)
		return
	}
	log.Println()
}

// default function to log fatal errors
// NOTE:
// the first arg passed will be used for the log prefix
var faillog LogFunc = func(elem ...any) {
	if len(elem) == 0 {
		return
	}
	log.SetPrefix(elem[0].(string))
	log.SetFlags(log.Ldate | log.Ltime)
	log.Fatalln(elem...)
}

// the print function is the default function used to print messages for the user
var print PrintFunc = func(msg ...any) {
	fmt.Println(msg...)
}

// ExtendedError is a type that allows for handling the error
// and continuing the function execution, returning immediately,
// or exiting the program. This error defaults to level 1
// logging, and 3 frames of stack trace.
// example: if err = errors.Handle(err, "not too bad")
type ExtendedError struct {
	werr        error     // error that is "wrapped" can be nil
	Id          string    // string constant, either LogErr, Panic or Fail
	usermsg     string    // error message shown to end users
	level       int       // logging level
	logfn       LogFunc   // function to use for logging
	printfn     PrintFunc // function to use to print user messages
	stackFrames int       // number of stack frame to log
	Handled     bool      // to prevent multiple logging episodes
}

// NewExtendedError returns a useable instance of the type
// if err is != nil it will be wrapped by werr
func New(err error, id string, usermsg string) *ExtendedError {
	e := &ExtendedError{Id: id, usermsg: usermsg, level: 1, stackFrames: 3, logfn: ErrLogger, printfn: ErrPrinter}
	if id == Fail || id == Panic {
		e.logfn = FailLogger
		e.level = 4
		e.stackFrames = -1
	} else {
		e.Id = LogErr
	}
	if err != nil {
		e.werr = err
	}

	return e
}

// Log changes the log level number.
// level 0 inhibits logging until changed or the error
// goes out of scope.
// optional to provide the function to use as the logger.
func (e *ExtendedError) Log(level int, logfn ...LogFunc) *ExtendedError {
	e.level = level
	if len(logfn) > 0 {
		e.logfn = logfn[0]
	}
	return e
}

// Print allows the user to pass a function to be used for user facing messages
func (e *ExtendedError) Print(printfn PrintFunc) *ExtendedError {
	e.printfn = printfn
	return e
}

// Stack allows the developer to set the number of stack frames
// to print in a stack trace.
// An argument of 0 turns off stack tracing for the call.
// An argument of < 0 sets no limit.
func (e *ExtendedError) Stack(frames int) *ExtendedError {
	e.stackFrames = frames
	return e
}

// (*ExtendedError).Handle() will log the error, print the message
// and return error back to caller or stop execution.
func (e *ExtendedError) Handle(logmsg ...any) error {
	handle(*e, concatMsg(logmsg...))
	e.Handled = true
	return *e
}

func (e ExtendedError) Error() string {
	em := e.usermsg
	if e.werr != nil {
		em += ": " + e.werr.Error()
		return em
	}
	return e.usermsg
}

func (e ExtendedError) Unwrap() error {
	return e.werr
}

func stackFrames(err error, f ...int) string {
	type stackTracer interface {
		StackTrace() pkgerr.StackTrace
	}

	er, ok := err.(stackTracer)
	if !ok {
		return ""
	}

	var fr int
	if len(f) > 0 {
		fr = f[0]
		if fr == 0 {
			return ""
		}
	}

	st := er.StackTrace()
	var ststr string
	if fr > 0 && fr < len(st) {
		ststr = fmt.Sprintf("%+v", st[0:fr]) // top f[0] frames
	} else {
		ststr = fmt.Sprintf("%+v", st) // all frames
	}

	return ststr
}

func concatMsg(msg ...any) string {
	var em string
	if len(msg) > 0 {
		for i, m := range msg {
			if i == 0 {
				em += fmt.Sprint(strings.TrimRight(m.(string), " "), ": ")
				continue
			}
			if i == len(msg)-1 {
				em += fmt.Sprint(m)
				continue
			}
			em += fmt.Sprint(m, ", ")
		}
		em += "\n"
	}
	return em
}

func handle(err ExtendedError, logmsg string) {
	e := pkgerr.WithStack(err)

	em := err.usermsg
	if err.werr != nil {
		em += ": " + err.Unwrap().Error()
	}

	err.printfn(em)

	st := stackFrames(e, err.stackFrames)
	if err.level > 0 {
		if err.stackFrames == 0 {
			err.logfn(err.Id, fmt.Sprintf("%s\nPrinted for user: %v\n", logmsg, err))
		} else {
			err.logfn(err.Id, fmt.Sprintf("%s\nPrinted for user: %+v\n%s ", logmsg, err, st))
		}
	}
}
