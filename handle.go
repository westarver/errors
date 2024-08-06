package errors

// Handle will respond according to the predetermined behavior expressed in call to New
func Handle(err error, msg ...any) error {
	if err == nil {
		return nil
	}

	if Err, ok := err.(ExtendedError); ok {
		return Err.Handle(msg...)
	}

	// not an ExtendedError type but will try to convert
	if len(msg) > 0 {
		if t, ok := msg[0].(string); ok { // first string will be interpreted as error id
			e := NewExtendedError(err, t, concatMsg(msg[1:]...))
			return e.Handle()
		}
	}
	// should never get here
	return err
}
