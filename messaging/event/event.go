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
 * @author KAnggara on Thursday 05/02/2026 06.43
 * @project PakaiWA
 * https://github.com/PakaiWA/pakaiwa-platform/tree/main/messaging/event
 */

package event

type Event interface {
	EventID() string   // rename agar idiomatis
	EventName() string // opsional, untuk header Kafka
	EventKey() string  // opsional, key partition
}
