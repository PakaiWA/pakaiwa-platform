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
 * @author KAnggara on Saturday 07/02/2026 18.22
 * @project pp
 * https://github.com/PakaiWA/PakaiWA/tree/main/~/work/PakaiWA/pp/errors
 */

// Package errors provides utility functions for panic-based error handling.
package errors

// Must returns the value v if err is nil, otherwise it panics with err.
// This is useful for handling errors in initialization code where recovery is not possible.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
