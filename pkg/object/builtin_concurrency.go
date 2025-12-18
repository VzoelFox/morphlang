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
				return NewError("buffer size must be INTEGER", ErrCodeTypeMismatch, 0, 0)
			}
		}
		ch := make(chan Object, buffer)
		return NewChannel(ch)
	})

	RegisterBuiltin("kirim", func(args ...Object) Object {
		if len(args) != 2 {
			return newArgumentError(len(args), 2)
		}
		chObj, ok := args[0].(*Channel)
		if !ok {
			return NewError(fmt.Sprintf("argument to `kirim` must be CHANNEL, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
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
			return NewError(fmt.Sprintf("argument to `terima` must be CHANNEL, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
		}

		val := <-chObj.Value
		return val
	})

	RegisterBuiltin("luncurkan", func(args ...Object) Object {
		return NewError("luncurkan() requires VM context", ErrCodeRuntime, 0, 0)
	})

	RegisterBuiltin("tunggu", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		threadObj, ok := args[0].(*Thread)
		if !ok {
			return NewError(fmt.Sprintf("argument to `tunggu` must be THREAD, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
		}

		val, ok := <-threadObj.Result
		if !ok {
			return NewNull()
		}
		return val
	})

	RegisterBuiltin("mutex_baru", func(args ...Object) Object {
		return NewMutex()
	})

	RegisterBuiltin("gembok", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		mu, ok := args[0].(*Mutex)
		if !ok {
			return NewError(fmt.Sprintf("argument to `gembok` must be MUTEX, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
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
			return NewError(fmt.Sprintf("argument to `buka_gembok` must be MUTEX, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
		}
		mu.Mu.Unlock()
		return NewNull()
	})
}
