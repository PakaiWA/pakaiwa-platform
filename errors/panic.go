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

package errors

import "fmt"

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func PanicIfError(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
