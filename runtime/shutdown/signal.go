/*
 * Copyright (c) 2026 KAnggara
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * See <https://www.gnu.org/licenses/gpl-3.0.html>.
 *
 * @author KAnggara on Sunday 08/02/2026 10.50
 * @project pp
 * https://github.com/PakaiWA/pakaiwa-platform/tree/main/runtime/shutdown
 */

package shutdown

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func WaitForSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
}

func Wait(ctx context.Context, sigs ...os.Signal) os.Signal {
	if len(sigs) == 0 {
		sigs = []os.Signal{os.Interrupt, syscall.SIGTERM}
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, sigs...)
	defer signal.Stop(ch)

	select {
	case sig := <-ch:
		return sig
	case <-ctx.Done():
		return nil
	}
}
