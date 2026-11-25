package paniccalls

var initialized = func() bool {
	panic("package initializer") // want "use of builtin panic is forbidden"
}()

func init() {
	panic("init") // want "use of builtin panic is forbidden"
}

func helper() {
	panic("helper")       // want "use of builtin panic is forbidden"
	(panic)("paren")      // want "use of builtin panic is forbidden"
	defer panic("defer")  // want "use of builtin panic is forbidden"
	go panic("goroutine") // want "use of builtin panic is forbidden"
	func() {
		panic("closure") // want "use of builtin panic is forbidden"
	}()
}
