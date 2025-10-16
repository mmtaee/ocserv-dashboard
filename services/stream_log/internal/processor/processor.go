package processor

import (
	"context"
	"github.com/mmtaee/ocserv-users-management/common/ocserv/occtl"
	"github.com/mmtaee/ocserv-users-management/common/ocserv/user"
)

var (
	ocservOcctlRepo occtl.OcservOcctlInterface = nil
	ocservUserRepo  user.OcservUserInterface   = nil
)

func Init() {
	ocservOcctlRepo = occtl.NewOcservOcctl()
	ocservUserRepo = user.NewOcservUser()
}

func Processor(ctx context.Context, streamText <-chan string, broadcaster chan<- string) {
	for {
		select {
		case <-ctx.Done():
			return
		case line, ok := <-streamText:
			if !ok {
				return
			}
			// SSE clients
			select {
			case broadcaster <- line:
			case <-ctx.Done():
				return
			}
		}
	}
}
