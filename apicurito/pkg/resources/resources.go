/*
 * Copyright (C) 2020 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package resources

import (
	"github.com/go-logr/logr"

	api "github.com/apicurio/apicurio-operators/apicurito/pkg/apis/apicur/v1"
	"github.com/apicurio/apicurio-operators/apicurito/pkg/configuration"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Resource struct {
	Client    client.Client
	Apicurito *api.Apicurito
	Cfg       *configuration.Config
	Logger    logr.Logger
}

type Generator interface {
	Generate() (rs []client.Object, err error)
	Routes() (rs []client.Object)
}

func (r Resource) Routes() (rs []client.Object) {
	rs = []client.Object{}
	rs = append(rs, apicuritoRoute(r.Apicurito))
	rs = append(rs, generatorRoute(r.Apicurito))

	return
}

func (r Resource) Generate() (rs []client.Object, err error) {
	rs = []client.Object{}

	c, err := apicuritoConfig(r.Client, r.Apicurito)
	if err != nil {
		r.Logger.Error(err, "error creating resource, name[%s]", c.GetName())
		return rs, err
	}
	rs = append(rs, c)

	rs = append(rs, apicuritoRoute(r.Apicurito))
	rs = append(rs, generatorRoute(r.Apicurito))
	rs = append(rs, generatorService(r.Apicurito))
	rs = append(rs, generatorDeployment(r.Cfg, r.Apicurito))
	rs = append(rs, apicuritoService(r.Apicurito))
	rs = append(rs, apicuritoDeployment(r.Cfg, r.Apicurito))

	return
}
