package apicurito

import (
	"context"
	"fmt"
	"strings"

	"github.com/RHsyseng/operator-utils/pkg/logs"
	"github.com/RHsyseng/operator-utils/pkg/utils/kubernetes"
	"github.com/RHsyseng/operator-utils/pkg/utils/openshift"
	"github.com/apicurio/apicurio-operators/apicurito/config"
	api "github.com/apicurio/apicurio-operators/apicurito/pkg/apis/apicur/v1"
	"github.com/apicurio/apicurio-operators/apicurito/pkg/resources"
	"github.com/ghodss/yaml"
	consolev1 "github.com/openshift/api/console/v1"
	routev1 "github.com/openshift/api/route/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var logu = logs.GetLogger("openshift-webconsole")

func consoleLinkExists() error {
	gvk := schema.GroupVersionKind{Group: "console.openshift.io", Version: "v1", Kind: "ConsoleLink"}
	return kubernetes.CustomResourceDefinitionExists(gvk)
}

func removeConsoleLink(c client.Client, api *api.Apicurito) {
	doDeleteConsoleLink(getUIConsoleLinkName(api), c, api)
	doDeleteConsoleLink(getGeneratorConsoleLinkName(api), c, api)
}

func doDeleteConsoleLink(consoleLinkName string, c client.Client, api *api.Apicurito) {
	consoleLink := &consolev1.ConsoleLink{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: consoleLinkName}, consoleLink)
	if err == nil && consoleLink != nil {
		err = c.Delete(context.TODO(), consoleLink)
		if err != nil {
			logu.Error(err, "Failed to delete the consolelink:", consoleLinkName)
		} else {
			logu.Info("deleted the consolelink:", consoleLinkName)
		}
	}
}

func createConsoleLink(c client.Client, api *api.Apicurito) {
	doCreateConsoleLink(getUIConsoleLinkName(api), resources.DefineUIName(api), c, api)
	doCreateConsoleLink(getGeneratorConsoleLinkName(api), resources.DefineGeneratorName(api), c, api)
}

func doCreateConsoleLink(consoleLinkName string, routeName string, c client.Client, api *api.Apicurito) {
	route := &routev1.Route{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: routeName, Namespace: api.Namespace}, route)
	if err == nil && route != nil {
		checkConsoleLink(route, consoleLinkName, api, c)
	}
}

func checkConsoleLink(route *routev1.Route, consoleLinkName string, api *api.Apicurito, c client.Client) {
	consoleLink := &consolev1.ConsoleLink{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: consoleLinkName}, consoleLink)
	if err != nil && apierrors.IsNotFound(err) {
		consoleLink = createNamespaceDashboardLink(consoleLinkName, route, api)
		if err := c.Create(context.TODO(), consoleLink); err != nil {
			logu.Error(err, "Console link is not created.")
		} else {
			logu.Info("Console link has been created. ", consoleLinkName)
		}
	} else if err == nil && consoleLink != nil {
		reconcileConsoleLink(context.TODO(), route, consoleLink, c)
	}
}

func reconcileConsoleLink(ctx context.Context, route *routev1.Route, link *consolev1.ConsoleLink, client client.Client) {
	url := "https://" + route.Spec.Host
	linkTxt := consoleLinkText(route)
	if url != link.Spec.Href || linkTxt != link.Spec.Text {
		if err := client.Update(ctx, link); err != nil {
			logu.Error(err, "failed to reconcile Console Link", link)
		}
	}
}

func getUIConsoleLinkName(api *api.Apicurito) string {
	return fmt.Sprintf("%s-%s", resources.DefineUIName(api), api.Namespace)
}

func getGeneratorConsoleLinkName(api *api.Apicurito) string {
	return fmt.Sprintf("%s-%s", resources.DefineGeneratorName(api), api.Namespace)
}

func consoleLinkText(route *routev1.Route) string {
	name := route.Name
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "-", "")
	name = strings.TrimPrefix(name, "apicurito")
	name = strings.TrimSuffix(name, "apicurito")
	name = strings.Title(name)
	return "Apicurito - " + name
}

func createNamespaceDashboardLink(consoleLinkname string, route *routev1.Route, api *api.Apicurito) *consolev1.ConsoleLink {
	return &consolev1.ConsoleLink{
		ObjectMeta: metav1.ObjectMeta{
			Name: consoleLinkname,
			Labels: map[string]string{
				"apicurito.io/name": api.ObjectMeta.Name,
			},
		},
		Spec: consolev1.ConsoleLinkSpec{
			Link: consolev1.Link{
				Text: consoleLinkText(route),
				Href: "https://" + route.Spec.Host,
			},
			Location: consolev1.NamespaceDashboard,
			NamespaceDashboard: &consolev1.NamespaceDashboardSpec{
				Namespaces: []string{api.Namespace},
			},
		},
	}
}

func ConsoleYAMLSampleExists() error {
	gvk := schema.GroupVersionKind{Group: "console.openshift.io", Version: "v1", Kind: "ConsoleYAMLSample"}
	return kubernetes.CustomResourceDefinitionExists(gvk)
}

func createConsoleYAMLSamples(c client.Client) {
	logu.Info("Loading CR YAML samples.")
	apicuritoCR := "samples/apicur_v1_apicurito_cr.yaml"
	asset, err := config.Asset(apicuritoCR)
	logu.Info("Sample:", apicuritoCR, string(asset))

	if err != nil {
		logu.Info("yaml", " name: ", apicuritoCR, " not created:  ", err.Error())
		return
	}

	logu.Info("Unmarshalling samples.")
	apicurito := api.Apicurito{}
	err = yaml.Unmarshal(asset, &apicurito)
	if err != nil {
		logu.Info("yaml", " name: ", apicuritoCR, " not created:  ", err.Error())
		return
	}

	logu.Info("Loading sample into openshift console")
	yamlSample, err := openshift.GetConsoleYAMLSample(&apicurito)
	if err != nil {
		logu.Info("yaml", " name: ", apicuritoCR, " not created:  ", err.Error())
		return
	}

	logu.Info("Resource being created: ")
	logu.Info(yamlSample)
	err = c.Create(context.TODO(), yamlSample)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			logu.Info("yaml", " name: ", apicuritoCR, " not created:+", err.Error())
		} else {
			logu.Info("yaml", " name: ", apicuritoCR, " already created.")
		}
		return
	}

	logu.Info("yaml", " name: ", apicuritoCR, " Created.")
}
