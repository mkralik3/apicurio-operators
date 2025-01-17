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
	"context"
	"fmt"

	routev1 "github.com/openshift/api/route/v1"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api "github.com/apicurio/apicurio-operators/apicurito/pkg/apis/apicur/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func apicuritoConfig(client client.Client, a *api.Apicurito) (c client.Object, err error) {
	c = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: DefineUIName(a),
		},
	}

	// Fetch the route host from generator
	r := &routev1.Route{}
	if err = client.Get(context.TODO(), types.NamespacedName{
		Name:      DefineGeneratorName(a),
		Namespace: a.Namespace}, r); err != nil {
		return c, err
	}

	if r.Spec.Host == "" {
		return c, fmt.Errorf("unable to fetch Host from route %s", r.Name)
	}

	c = &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefineUIName(a),
			Namespace: a.Namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(a, schema.GroupVersionKind{
					Group:   api.SchemeGroupVersion.Group,
					Version: api.SchemeGroupVersion.Version,
					Kind:    a.Kind,
				}),
			},
		},
		Data: map[string]string{
			"config.js": fmt.Sprintf("var ApicuritoConfig = { \"generators\": [ { \"name\":\"Fuse Camel Project\", \"url\":\"https://%s/api/v1/generate/camel-project.zip\" } ] }", r.Spec.Host),
		},
	}
	return
}
