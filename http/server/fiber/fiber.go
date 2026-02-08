/*
 * Copyright (c) 2025 KAnggara75
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * See <https://www.gnu.org/licenses/gpl-3.0.html>.
 *
 * @author KAnggara75 on Sat 06/09/25 10.59
 * @project PakaiWA httpserver
 * https://github.com/PakaiWA/PakaiWA/tree/main/internal/pkg/httpserver
 */

package httpserver

import "github.com/gofiber/fiber/v3"

type Options struct {
	AppName      string
	ErrorHandler fiber.ErrorHandler

	TrustProxy         bool
	EnableIPValidation bool
	TrustedProxies     []string
}

func NewFiber(opts Options) *fiber.App {
	return fiber.New(fiber.Config{
		AppName:            opts.AppName,
		ErrorHandler:       opts.ErrorHandler,
		TrustProxy:         opts.TrustProxy,
		EnableIPValidation: opts.EnableIPValidation,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Proxies: opts.TrustedProxies,
		},
	})
}

func DefaultOptions() Options {
	return Options{
		TrustProxy:         true,
		EnableIPValidation: true,
	}
}
