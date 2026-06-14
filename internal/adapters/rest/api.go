/*
 * Copyright 2026 Holger de Carne
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package rest

import (
	"net/http"
)

//	@title			PIM MCP Server REST API
//	@version		1.0
//	@description	MCP server providing Agent access to PIM services.

//	@contact.url	https://github.com/tdrn-org/pim-mcp

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:9123
//	@BasePath	/api/v1

type API interface {

	// GET @BasePath/ping
	//
	//	@Summary		Ping the server
	//	@Description	Ping the server to check general health
	//	@Produce		text/plain
	//	@Success		200	{string}	string	"ok"
	//	@Failure		500	{string}	string	"server error"
	//	@Router			/api/v1/ping [get]
	HandlePingGet(w http.ResponseWriter, r *http.Request)
}
