package object

import "fmt"

func init() {
	// --- Concurrency Primitives (Scaffolding) ---

	RegisterBuiltin("saluran_baru", func(args ...Object) Object {
		buffer := 0
		if len(args) > 0 {
			if i, ok := args[0].(*Integer); ok {
				buffer = int(i.GetValue())
			} else {
				return &Error{Code: ErrCodeTypeMismatch, Message: "buffer size must be INTEGER"}
			}
		}
		ch := make(chan Object, buffer)
		return &Channel{Value: ch}
	})

	RegisterBuiltin("kirim", func(args ...Object) Object {
		if len(args) != 2 {
			return newArgumentError(len(args), 2)
		}
		chObj, ok := args[0].(*Channel)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `kirim` must be CHANNEL, got %s", args[0].Type())}
		}

		chObj.Value <- args[1]
		return NewNull()
	})

	RegisterBuiltin("terima", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		chObj, ok := args[0].(*Channel)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `terima` must be CHANNEL, got %s", args[0].Type())}
		}

		val := <-chObj.Value
		return val
	})

	RegisterBuiltin("luncurkan", func(args ...Object) Object {
		return &Error{Code: ErrCodeRuntime, Message: "luncurkan() requires VM context"}
	})

	RegisterBuiltin("gabung", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		threadObj, ok := args[0].(*Thread)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `gabung` must be THREAD, got %s", args[0].Type())}
		}

		val, ok := <-threadObj.Result
		if !ok {
			return NewNull()
		}
		return val
	})

	RegisterBuiltin("mutex_baru", func(args ...Object) Object {
		return &Mutex{}
	})

	RegisterBuiltin("gembok", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		mu, ok := args[0].(*Mutex)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `gembok` must be MUTEX, got %s", args[0].Type())}
		}
		mu.Mu.Lock()
		return NewNull()
	})

	RegisterBuiltin("buka_gembok", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		mu, ok := args[0].(*Mutex)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `buka_gembok` must be MUTEX, got %s", args[0].Type())}
		}
		mu.Mu.Unlock()
		return NewNull()
	})
}
