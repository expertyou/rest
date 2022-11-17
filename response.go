package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Response struct {
	code   int
	msg    string
	data   interface{}
	cookie *http.Cookie
}

func Ok(msg string) Response {
	return Response{
		code: http.StatusOK,
		msg:  msg,
	}
}

func NoContent() Response {
	return Response{
		code: http.StatusNoContent,
	}
}

func (r Response) WithData(d interface{}) Response {
	return Response{
		code: r.code,
		msg:  r.msg,
		data: d,
	}
}

func (r Response) WithCookie(cookie *http.Cookie) Response {
	return Response{
		code:   r.code,
		msg:    r.msg,
		data:   r.data,
		cookie: cookie,
	}
}

func (r Response) Code() int       { return r.code }
func (r Response) Message() string { return r.msg }

func (r Response) Write(w http.ResponseWriter) error {

	// setting the cookie must happen before we write anything
	// to the http.ResponseWriter otherwise the cookie is not set.
	// This also includes writing of the status code and response header
	if r.cookie != nil {
		http.SetCookie(w, r.cookie)
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(r.code)

	if r.data != nil {
		return json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  r.code,
			"message": r.msg,
			"data":    r.data,
			"ts":      time.Now().Unix(),
		})
	}

	if r.msg == "" {
		return nil
	}

	return json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  r.code,
		"message": r.msg,
		"ts":      time.Now().Unix(),
	})

}

type Error struct {
	StatusCode int    `json:"code"`
	Status     error  `json:"error"`
	Msg        string `json:"message"`
}

func NewError(code int, msg string, args ...interface{}) *Error {
	return &Error{
		StatusCode: code,
		Status:     fmt.Errorf(msg, args...),
	}
}

func (e Error) Err() error     { return e.Status }
func (e Error) Code() int      { return e.StatusCode }
func (e Error) String() string { return e.Status.Error() }

func (e *Error) WithMessage(msg string) *Error {
	return &Error{
		StatusCode: e.StatusCode,
		Status:     e.Status,
		Msg:        msg,
	}
}

func (e Error) Write(w http.ResponseWriter) error {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(e.StatusCode)
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"status": e.StatusCode,
		"error":  e.Msg,
		"ts":     time.Now().Unix(),
	})
}

func Internal(err string, args ...interface{}) *Error {
	return &Error{
		StatusCode: http.StatusInternalServerError,
		Status:     fmt.Errorf(err, args...),
	}
}

func BadRequest(err string, args ...interface{}) *Error {
	return &Error{
		StatusCode: http.StatusBadRequest,
		Status:     fmt.Errorf(err, args...),
	}
}

func NotFound(err string, args ...interface{}) *Error {
	return &Error{
		StatusCode: http.StatusNotFound,
		Status:     fmt.Errorf(err, args...),
	}
}

func Forbidden(err string, args ...interface{}) *Error {
	return &Error{
		StatusCode: http.StatusForbidden,
		Status:     fmt.Errorf(err, args...),
	}
}

func NotAuthorized(err string, args ...interface{}) *Error {
	return &Error{
		StatusCode: http.StatusUnauthorized,
		Status:     fmt.Errorf(err, args...),
	}
}

func WriteInternal(w http.ResponseWriter, msg string) error {
	return Internal("").WithMessage(msg).Write(w)
}

func WriteBadRequest(w http.ResponseWriter, msg string) error {
	return BadRequest("").WithMessage(msg).Write(w)
}

func WriteNotFound(w http.ResponseWriter, msg string) error {
	return NotFound("").WithMessage(msg).Write(w)
}

func WriteForbidden(w http.ResponseWriter, msg string) error {
	return Forbidden("").WithMessage(msg).Write(w)
}

func WriteNotAuthorized(w http.ResponseWriter, msg string) error {
	return NotAuthorized("").WithMessage(msg).Write(w)
}
