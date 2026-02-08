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
 * @author KAnggara on Sunday 08/02/2026 09.06
 * @project pp
 * https://github.com/PakaiWA/pp/tree/main/security/password
 */

package password

import "golang.org/x/crypto/bcrypt"

// Hash menghasilkan password hash untuk disimpan ke DB
func Hash(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(plain),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Compare membandingkan hash di DB dengan password plaintext
func Compare(hashed string, plain string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(hashed),
		[]byte(plain),
	)
	return err == nil
}
