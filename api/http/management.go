// Copyright 2021 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mendersoftware/go-lib-micro/rest.utils"
	"github.com/mendersoftware/reporting/model"
	"github.com/pkg/errors"

	"github.com/mendersoftware/reporting/app/reporting"
)

type ManagementController struct {
	reporting reporting.App
}

func NewManagementController(r reporting.App) *ManagementController {
	return &ManagementController{
		reporting: r,
	}
}

func (mc *ManagementController) Search(c *gin.Context) {
	tenant := c.GetHeader("tenant")
	if tenant == "" {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.New("need `tenant` header"),
		)
		return
	}

	params, err := parseSearchParams(c)

	if err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "malformed request body"),
		)
		return
	}

	ctx := context.WithValue(c.Request.Context(), "tenant", tenant)

	output := c.Query("output")
	switch output {
	case "":
		res, _, err := mc.reporting.SearchDevices(ctx, params)
		if err != nil {
			rest.RenderError(c,
				http.StatusInternalServerError,
				err,
			)
		}

		c.JSON(http.StatusOK, res)

	case "raw_es":
		res, err := mc.reporting.DebugSearchDevicesRawES(ctx, params)
		if err != nil {
			rest.RenderError(c,
				http.StatusInternalServerError,
				err,
			)
		}
		c.JSON(http.StatusOK, res)
	}
}

func parseSearchParams(c *gin.Context) (*model.SearchParams, error) {
	var searchParams model.SearchParams

	err := c.ShouldBindJSON(&searchParams)
	if err != nil {
		return nil, err
	}

	if searchParams.Page < 1 {
		searchParams.Page = 1
	}
	if searchParams.PerPage < 1 {
		searchParams.PerPage = 20
	}

	if err := searchParams.Validate(); err != nil {
		return nil, err
	}

	return &searchParams, nil
}
