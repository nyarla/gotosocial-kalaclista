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

package web

import (
	"context"
	"fmt"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/gin-gonic/gin"
)

const (
	domainBlockListPath = aboutPath + "/suspended"
)

func (m *Module) domainBlockListGETHandler(c *gin.Context) {
	instance, errWithCode := m.processor.InstanceGetV1(c.Request.Context())
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	// Return instance we already got from the db,
	// don't try to fetch it again when erroring.
	instanceGet := func(ctx context.Context) (*apimodel.InstanceV1, gtserror.WithCode) {
		return instance, nil
	}

	// We only serve text/html at this endpoint.
	if _, err := apiutil.NegotiateAccept(c, apiutil.TextHTML); err != nil {
		apiutil.WebErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), instanceGet)
		return
	}

	if !config.GetInstanceExposeSuspendedWeb() {
		err := fmt.Errorf("this instance does not publicy expose its blocklist")
		apiutil.WebErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), instanceGet)
		return
	}

	domainBlocks, errWithCode := m.processor.InstancePeersGet(c.Request.Context(), true, false, false)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, instanceGet)
		return
	}

	page := apiutil.WebPage{
		Template:    "domain-blocklist.tmpl",
		Instance:    instance,
		OGMeta:      apiutil.OGBase(instance),
		Stylesheets: []string{cssFA},
		Extra:       map[string]any{"blocklist": domainBlocks},
	}

	apiutil.TemplateWebPage(c, page)
}
