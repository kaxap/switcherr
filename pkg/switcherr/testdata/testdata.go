package p

import (
	"errors"
	"fmt"
)

var err1 = errors.New("")
var err2 = errors.New("")
var err3 = errors.New("")
var err4 = errors.New("")
var v = -1

func funcReturnsError() (int, error) {
	return v, fmt.Errorf("")
}

// error properly handled
func _() error {
	_, err := funcReturnsError()
	switch { //
	case errors.Is(err, err1):
		//
	case errors.Is(err, err2) || errors.Is(err, err3):
		//
	case errors.Is(err, err2) || errors.Is(err, err3) || errors.Is(err, err4):
		//
	default:
		// default action should count as the last case
		return err
	}
	return nil
}

// error properly handled
func _() error {
	_, err := funcReturnsError()
	someFunc := func(err error) error {
		return err
	}
	switch { //
	case errors.Is(err, err1):
		//
	case errors.Is(err, err2) || errors.Is(err, err3):
		//
	case errors.Is(err, err2) || errors.Is(err, err3) || errors.Is(err, err4):
		//
	default:
		// default action should count as the last case
		if err != nil {
			return someFunc(err)
		}
	}
	return nil
}

// error properly handled
func _() {
	type S struct {
		err error
	}
	c := make(chan S, 1)
	s := S{err1}
	c <- s
	s2 := <-c

	switch {
	case errors.Is(s2.err, err1):
		break
	case s2.err != nil:
		//
	default:

	}
}

// improper error handling
func _() {
	r, err := funcReturnsError()
	switch { // want "error type check comes after err != nil"
	case err != nil:
		//
	case errors.Is(err, err1):
		//
	case errors.Is(err, err2) || errors.Is(err, err3):
		//
	case errors.Is(err, err2) || errors.Is(err, err3) || errors.Is(err, err4):
		//
	case r == 0:
		//
	}
}

// improper error handling
func _() {
	r, err := funcReturnsError()
	switch { // want "case err != nil comes after non-error case"
	case r != 0:
		//
	case errors.Is(err, err1):
		//
	case err != nil:
		//
	}
}

// improper error handling
func _() {
	r, err := funcReturnsError()
	switch { // want "case err != nil is missing"
	case errors.Is(err, err1):
		//
	case r != 0:
	}
}

func _() {
	r, err := funcReturnsError()
	switch {
	case errors.Is(err, err1):
		//
	case err != nil:
		//
	case r != 0:
	}
}

// improper error handling
func _() {
	r, err := funcReturnsError()
	switch {
	case errors.Is(err, err1):
		switch { // want "error type check comes after err != nil"
		case err != nil:
			//
		case errors.Is(err, err1):
			//
		case errors.Is(err, err2) || errors.Is(err, err3):
			//
		case errors.Is(err, err2) || errors.Is(err, err3) || errors.Is(err, err4):
			//
		case r == 0:
			//
		}
	case err != nil:
		r, err := funcReturnsError()
		switch { // want "case err != nil comes after non-error case"
		case r != 0:
			//
		case errors.Is(err, err1):
			//
		case err != nil:
			//
		}
	case r != 0:
		r, err := funcReturnsError()
		switch { // want "case err != nil is missing"
		case errors.Is(err, err1):
			//
		case r != 0:
		}
	}
}

// proper error handling, this currently fails
func _() {
	_, err := funcReturnsError()
	switch {
	case errors.Is(err, err1):
		//
	case errors.Is(err, err2) || errors.Is(err, err3):
		//
	case errors.Is(err, err2) || errors.Is(err, err3) || errors.Is(err, err4):
		//
	}

	// late check
	if err != nil {
		//
	}
}

// proper error handling, this currently fails
func _() error {
	_, err := funcReturnsError()
	switch { //
	case errors.Is(err, err1):
		//
	case errors.Is(err, err2) || errors.Is(err, err3):
		//
	case errors.Is(err, err2) || errors.Is(err, err3) || errors.Is(err, err4):
		//
	default:
		// default action should count as the last case
		return err
	}
	return nil
}

// proper error handling, this currently fails
func _() error {
	_, err := funcReturnsError()
	switch {
	case errors.Is(err, err1):
		//
	case errors.Is(err, err2) || errors.Is(err, err3):
		//
	case errors.Is(err, err2) || errors.Is(err, err3) || errors.Is(err, err4):
		//
	}
	return err
}

// this currently fails
func _() error {
	r, err := funcReturnsError()
	switch { // want "case err != nil comes after non-error case"
	case errors.Is(err, err1):
		//
	case errors.Is(err, err2) || errors.Is(err, err3):
		//
	case errors.Is(err, err2) || errors.Is(err, err3) || errors.Is(err, err4):
		//
	case r == 0:

	}
	return err
}

// error properly handled
func _() error {
	r, err := funcReturnsError()
	switch {
	case err == nil:
		//
	case errors.Is(err, err2) || errors.Is(err, err3):
		//
	case errors.Is(err, err2) || errors.Is(err, err3) || errors.Is(err, err4):
		//
	case err != nil:
		//
	case r > 0:
		//
	}
	return err
}
