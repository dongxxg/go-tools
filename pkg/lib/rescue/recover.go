package rescue

import (
	"context"
	"runtime/debug"

	"unitechs.com/unios-dice/uni-base/core/log"
)

// Recover is used with defer to do cleanup on panics.
// Use it like:
//
//	defer Recover(func() {})
func Recover(cleanups ...func()) {
	for _, cleanup := range cleanups {
		cleanup()
	}

	if p := recover(); p != nil {
		log.Error(p)
	}
}

// RecoverCtx is used with defer to do cleanup on panics.
func RecoverCtx(ctx context.Context, cleanups ...func()) {
	for _, cleanup := range cleanups {
		cleanup()
	}

	if p := recover(); p != nil {
		log.Errorf("%+v\n%s", p, debug.Stack())
	}
}

// Safego runs funVar with panic recovery and optional resolve/reject/finally callbacks.
// name is used only for the panic log message.
func Safego(funVar func(), name string) {
	safego(funVar, name, nil, nil, nil)
}

// SafegoResolve runs res when funVar returns normally (no panic).
func SafegoResolve(funVar func(), name string, res func()) {
	safego(funVar, name, res, nil, nil)
}

// SafegoReject runs rej when funVar panics.
func SafegoReject(funVar func(), name string, rej func()) {
	safego(funVar, name, nil, rej, nil)
}

// SafegoFinally runs fin after funVar regardless of panic.
func SafegoFinally(funVar func(), name string, fin func()) {
	safego(funVar, name, nil, nil, fin)
}

func safego(funVar func(), name string, res func(), rej func(), fin func()) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("%s panic as: %s\n%s", name, err, debug.Stack())
			if rej != nil {
				rej()
			}
		} else if res != nil {
			res()
		}
		if fin != nil {
			fin()
		}
	}()
	funVar()
}
