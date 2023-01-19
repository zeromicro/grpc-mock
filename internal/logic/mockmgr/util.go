package mockmgr

import (
	"fmt"
	"runtime"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

func GO(f func(), funcName string) {
	go func() {
		if f == nil {
			return
		}

		defer func() {
			if e := recover(); e != nil {
				buf := make([]byte, 64<<10)
				buf = buf[:runtime.Stack(buf, false)]

				panicError := fmt.Errorf("%v", e)

				logx.Errorf("recover from panic", zap.String("func_name", funcName), logx.Field("panic_error", panicError.Error()), logx.Field("stack", string(buf)))
			}
		}()

		f()
	}()
}
