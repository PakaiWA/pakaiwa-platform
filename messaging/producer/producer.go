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
 * @author KAnggara on Sunday 01/02/2026 17.39
 * @project PakaiWA
 * https://github.com/PakaiWA/pakaiwa-platform/tree/main/messaging/producer
 */

package producer

import (
	"context"
)

type MessageProducer interface {
	Send(ctx context.Context, topic string, key []byte, clientJID []byte, value []byte) error
	Flush(timeoutMs int) int
	Close() error
}
