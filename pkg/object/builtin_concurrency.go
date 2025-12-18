package object

import "fmt"

func init() {
	// --- Concurrency Primitives (Scaffolding) ---

	// saluran_baru(buffer_size?) -> Channel
	RegisterBuiltin("saluran_baru", func(args ...Object) Object {
		buffer := 0
		if len(args) > 0 {
			if i, ok := args[0].(*Integer); ok {
				buffer = int(i.Value)
			} else {
				return &Error{Code: ErrCodeTypeMismatch, Message: "buffer size must be INTEGER"}
			}
		}
		ch := make(chan Object, buffer)
		return &Channel{Value: ch}
	})

	// kirim(saluran, nilai) -> Null
	RegisterBuiltin("kirim", func(args ...Object) Object {
		if len(args) != 2 {
			return newArgumentError(len(args), 2)
		}
		chObj, ok := args[0].(*Channel)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `kirim` must be CHANNEL, got %s", args[0].Type())}
		}

		// Blocking send
		chObj.Value <- args[1]
		return &Null{}
	})

	// terima(saluran) -> Object
	RegisterBuiltin("terima", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		chObj, ok := args[0].(*Channel)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `terima` must be CHANNEL, got %s", args[0].Type())}
		}

		// Blocking receive
		val := <-chObj.Value
		return val
	})

	// luncurkan(fungsi) -> Null
	// This is a placeholder. The actual implementation is intercepted by the VM.
	RegisterBuiltin("luncurkan", func(args ...Object) Object {
		return &Error{Code: ErrCodeRuntime, Message: "luncurkan() requires VM context"}
	})

	// gabung(utas) -> Object
	RegisterBuiltin("gabung", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		threadObj, ok := args[0].(*Thread)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `gabung` must be THREAD, got %s", args[0].Type())}
		}

		// Blocking wait
		val, ok := <-threadObj.Result
		if !ok {
			// Channel closed without value? Should not happen with our implementation unless panic
			return &Null{}
		}
		return val
	})

	// mutex_baru() -> Mutex
	RegisterBuiltin("mutex_baru", func(args ...Object) Object {
		return &Mutex{}
	})

	// kunci(mutex) -> Null
	RegisterBuiltin("kunci", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		mu, ok := args[0].(*Mutex)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `kunci` must be MUTEX, got %s", args[0].Type())}
		}
		mu.Mu.Lock()
		return &Null{}
	})

	// buka(mutex) -> Null
	RegisterBuiltin("buka", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		mu, ok := args[0].(*Mutex)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `buka` must be MUTEX, got %s", args[0].Type())}
		}
		mu.Mu.Unlock()
		return &Null{}
	})
}
