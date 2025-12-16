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
				return &Error{Message: "buffer size must be INTEGER"}
			}
		}
		ch := make(chan Object, buffer)
		return &Channel{Value: ch}
	})

	// kirim(saluran, nilai) -> Null
	RegisterBuiltin("kirim", func(args ...Object) Object {
		if len(args) != 2 {
			return &Error{Message: fmt.Sprintf("argument mismatch: expected 2, got %d", len(args))}
		}
		chObj, ok := args[0].(*Channel)
		if !ok {
			return &Error{Message: fmt.Sprintf("argument to `kirim` must be CHANNEL, got %s", args[0].Type())}
		}

		// Blocking send
		chObj.Value <- args[1]
		return &Null{}
	})

	// terima(saluran) -> Object
	RegisterBuiltin("terima", func(args ...Object) Object {
		if len(args) != 1 {
			return &Error{Message: fmt.Sprintf("argument mismatch: expected 1, got %d", len(args))}
		}
		chObj, ok := args[0].(*Channel)
		if !ok {
			return &Error{Message: fmt.Sprintf("argument to `terima` must be CHANNEL, got %s", args[0].Type())}
		}

		// Blocking receive
		val := <-chObj.Value
		return val
	})

	// luncurkan(fungsi) -> Null
	// This is a placeholder. The actual implementation is intercepted by the VM.
	RegisterBuiltin("luncurkan", func(args ...Object) Object {
		return &Error{Message: "luncurkan() requires VM context"}
	})
}
