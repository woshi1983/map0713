package resolver

import (
	"context"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
)

const (
	defNamespace           = "vmware-system-map"
	defRegistrySecret      = "vmware-system-map-registry"
	defSSHPrivateKeySecret = "vmware-system-map-ssh"
	defInitJobName         = "cluster-initiator"
	defInitPackageName     = "cert-manager"

	defTkgServiceConfiguration = "tkg-service-configuration"

	defLabelApplication = "map.tanzu.vmware.com/application"
	defLabelComponent   = "map.tanzu.vmware.com/component"
	defLabelFederation  = "map.tanzu.vmware.com/federation"
	defLabelInstance    = "map.tanzu.vmware.com/instance"
	defLabelImage       = "map.tanzu.vmware.com/image"
	defLabelRevision    = "map.tanzu.vmware.com/revision"
	defLabelIssuer      = "map.tanzu.vmware.com/issuer"
	defLabelController  = "map.tanzu.vmware.com/controller"
	defLabelManaged     = "map.tanzu.vmware.com/managed"
	defLabelInternal    = "map.tanzu.vmware.com/internal"
	defLabelRestart     = "map.tanzu.vmware.com/restart"

	defAVIServiceEngineGroup = "map-segroup"
)

var (
	defExcludedNamespaces = map[string]bool{
		"avi-system":                   true,
		"build-service":                true,
		"cert-manager":                 true,
		"default":                      true,
		"kpack":                        true,
		"kube-node-lease":              true,
		"kube-public":                  true,
		"kube-system":                  true,
		"stacks-operator-system":       true,
		"tanzu-package-repo-global":    true,
		"tanzu-system":                 true,
		"tanzu-system-dashboards":      true,
		"tanzu-system-ingress":         true,
		"tanzu-system-logging":         true,
		"tanzu-system-monitoring":      true,
		"tkg-system":                   true,
		"tkg-system-public":            true,
		"tkg-system-telemetry":         true,
		"vmware-system-auth":           true,
		"vmware-system-cloud-provider": true,
		"vmware-system-csi":            true,
		"vmware-system-elastic":        true,
		"vmware-system-map":            true,
		"vmware-system-octant":         true,
	}
)

func finalizePageInfo(page *model.PageInfo, total int64) {
	if page.Number < 0 {
		page.Number = 0
		page.Limit = -1
	}
	if page.Limit == 0 {
		page.Limit = 10
	}
	if page.Limit < 0 {
		page.Number = 0
		page.Offset = 0
		page.Count = total
		return
	}
	if offset := page.Number * page.Limit; total > offset {
		page.Offset = offset
		if total > offset+page.Limit {
			page.HasNextPage = true
			page.Count = page.Limit
			return
		}
		page.Count = total - offset
		return
	}
	page.Offset = total
	page.Count = 0
}

func paginate(limit, offset int64, do func(lim, off int64) error) error {
	if limit < 1001 {
		return do(limit, offset)
	}
	n := limit/1000 + 1
	for i := int64(0); i < n; i++ {
		l := int64(1000)
		if i == n-1 {
			l = limit - i*1000
			if l < 1 {
				break
			}
		}
		o := offset + i*1000
		if err := do(l, o); err != nil {
			return err
		}
	}
	return nil
}

var (
	silenceGuardCtxKey = &contextKey{"silenceGuard"}
	omitNotFoundCtxKey = &contextKey{"omitNotFound"}
)

type contextKey struct {
	name string
}

type empty struct{}

func WithGuardSilenced(ctx context.Context) context.Context {
	return context.WithValue(ctx, silenceGuardCtxKey, &empty{})
}

func IsGuardSilenced(ctx context.Context) bool {
	_, ok := ctx.Value(silenceGuardCtxKey).(*empty)
	return ok
}

func WithNotFoundOmitted(ctx context.Context) context.Context {
	return context.WithValue(ctx, omitNotFoundCtxKey, &empty{})
}

func IsNotFoundOmitted(ctx context.Context) bool {
	_, ok := ctx.Value(omitNotFoundCtxKey).(*empty)
	return ok
}

const (
	manifestTemplateNSPlaceholder = "{{.Namespace}}"
)

type manifestTemplateRenderValue struct {
	Namespace string
}
