package k8s

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func RenderChart(namespace, releaseName, fpath string, values map[string]interface{}) (crd, manifest []byte, err error) {
	chrt, err := loader.Load(fpath)
	if err != nil {
		return crd, manifest, err
	}

	// for library charts, just return an empty manifest.
	if err := checkIfInstallable(chrt); err != nil {
		return crd, manifest, nil
	}

	if req := chrt.Metadata.Dependencies; req != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/helm/helm/issues/2209
		if err := action.CheckDependencies(chrt, req); err != nil {
			return crd, manifest, err
		}
	}

	buf := bytes.NewBuffer(nil)
	if crds := chrt.CRDObjects(); len(crds) > 0 {
		// We do these one file at a time in the order they were read.
		for i, obj := range crds {
			if i > 0 {
				fmt.Fprintln(buf, "\n---")
			}
			buf.Write(bytes.TrimRight(obj.File.Data, "\n"))
		}
	}

	cfg := &action.Configuration{}
	cfg.RESTClientGetter = &genericclioptions.ConfigFlags{}
	cfg.Log = func(format string, v ...interface{}) {}

	install := action.NewInstall(cfg)
	install.ReleaseName = releaseName
	install.Namespace = namespace
	install.ClientOnly = true
	install.DryRun = true
	r, err := install.Run(chrt, values)
	if err != nil {
		return crd, manifest, err
	}
	return buf.Bytes(), []byte(r.Manifest), nil
}

// checkIfInstallable validates if a chart can be installed
//
// Application chart type is only installable
func checkIfInstallable(ch *chart.Chart) error {
	switch ch.Metadata.Type {
	case "", "application":
		return nil
	}
	return errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}
