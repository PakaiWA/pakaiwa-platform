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
 * @author KAnggara on Saturday 07/02/2026 18.25
 * @project pp
 * https://github.com/PakaiWA/PakaiWA/tree/main/~/work/PakaiWA/pp/db/postgres
 */

package postgres

import "time"

type Config struct {
	DSN               string
	MinConns          int32
	MaxConns          int32
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	ConnectTimeout    time.Duration
}
