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
 * @author KAnggara on Saturday 07/02/2026 20.02
 * @project pp
 * https://github.com/PakaiWA/PakaiWA/tree/main/~/work/PakaiWA/pp/messaging/kafka
 */

package kafka

type ProducerConfig struct {
	Brokers  []string
	ClientID string
	Options  map[string]any
}
