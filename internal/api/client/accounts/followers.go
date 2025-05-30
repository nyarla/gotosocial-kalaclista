// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package accounts

import (
	"errors"
	"net/http"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"github.com/gin-gonic/gin"
)

// AccountFollowersGETHandler swagger:operation GET /api/v1/accounts/{id}/followers accountFollowers
//
// See followers of account with given id.
//
// The next and previous queries can be parsed from the returned Link header.
// Example:
//
// ```
// <https://example.org/api/v1/accounts/0657WMDEC3KQDTD6NZ4XJZBK4M/followers?limit=80&max_id=01FC0SKA48HNSVR6YKZCQGS2V8>; rel="next", <https://example.org/api/v1/accounts/0657WMDEC3KQDTD6NZ4XJZBK4M/followers?limit=80&min_id=01FC0SKW5JK2Q4EVAV2B462YY0>; rel="prev"
// ````
//
// If account `hide_collections` is true, and requesting account != target account, no results will be returned.
//
//	---
//	tags:
//	- accounts
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: id
//		type: string
//		description: Account ID.
//		in: path
//		required: true
//	-
//		name: max_id
//		type: string
//		description: >-
//			Return only follower accounts *OLDER* than the given max ID.
//			The follower account with the specified ID will not be included in the response.
//			NOTE: the ID is of the internal follow, NOT any of the returned accounts.
//		in: query
//		required: false
//	-
//		name: since_id
//		type: string
//		description: >-
//			Return only follower accounts *NEWER* than the given since ID.
//			The follower account with the specified ID will not be included in the response.
//			NOTE: the ID is of the internal follow, NOT any of the returned accounts.
//		in: query
//		required: false
//	-
//		name: min_id
//		type: string
//		description: >-
//			Return only follower accounts *IMMEDIATELY NEWER* than the given min ID.
//			The follower account with the specified ID will not be included in the response.
//			NOTE: the ID is of the internal follow, NOT any of the returned accounts.
//		in: query
//		required: false
//	-
//		name: limit
//		type: integer
//		description: Number of follower accounts to return.
//		default: 40
//		minimum: 1
//		maximum: 80
//		in: query
//		required: false
//
//	security:
//	- OAuth2 Bearer:
//		- read:accounts
//
//	responses:
//		'200':
//			name: accounts
//			description: Array of accounts that follow this account.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/account"
//			headers:
//				Link:
//					type: string
//					description: Links to the next and previous queries.
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'404':
//			description: not found
//		'406':
//			description: not acceptable
//		'500':
//			description: internal server error
func (m *Module) AccountFollowersGETHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeReadAccounts,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	targetAcctID := c.Param(IDKey)
	if targetAcctID == "" {
		err := errors.New("no account id specified")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	page, errWithCode := paging.ParseIDPage(c,
		1,  // min limit
		80, // max limit
		40, // default limit
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.Account().FollowersGet(c.Request.Context(), authed.Account, targetAcctID, page)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if resp.LinkHeader != "" {
		c.Header("Link", resp.LinkHeader)
	}

	apiutil.JSON(c, http.StatusOK, resp.Items)
}
