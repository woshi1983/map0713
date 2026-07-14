package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	graphql1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"go.uber.org/multierr"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// UpdateNode is the resolver for the UpdateNode field.
func (r *mutationResolver) UpdateNode(ctx context.Context, input model.UpdateNode) (*model.Node, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadWrite, model1.PolicyObjectClusterResource)
	}

	c := &model1.Cluster{ID: input.Cluster}
	if err := db.Take(c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "input.cluster", c.ID.String())
		}
		panic(&ErrDBConnection{err})
	}

	if c.AuthToken == nil || len(c.AuthToken) == 0 {
		return nil, errInvalidRequest(ctx, "input.cluster", "cluster is not initialized", input.Cluster.String())
	}

	block, err := cryptoutil.NewAESGCMCipher(c.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}
	wc, err := k8s.NewKubernetesFromToken(c.Endpoint, block.Decrypt(c.AuthToken), rest.TLSClientConfig{
		Insecure: c.InsecureTLS,
		CAData:   c.CACert,
	})
	if err != nil {
		panic(err)
	}

	obj, err := wc.GetNode(input.Name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			return nil, errObjectNotFound(ctx, "input.name", input.Name)
		}
		return nil, errGatewayError(ctx, err)
	}

	val := nodePatcher{
		Metadata: nodeMetadataPatcher{
			Labels: make(map[string]*string),
		},
		Spec: nodeSpecPatcher{
			Unschedulable: obj.Spec.Unschedulable,
			Taints:        obj.Spec.Taints,
		},
	}

	if input.Labels != nil {
		for k, v := range input.Labels {
			val.Metadata.Labels[k] = util.StringPtr(v)
		}
	}
	for k := range obj.Labels {
		ss := strings.SplitN(k, "/", 2)
		prefix := ss[0]
		if prefix == "kubernetes.io" || strings.HasSuffix(prefix, ".kubernetes.io") {
			continue
		}

		if _, ok := val.Metadata.Labels[k]; !ok {
			val.Metadata.Labels[k] = nil
		}
	}

	if input.Spec != nil {
		if input.Spec.Unschedulable != nil {
			val.Spec.Unschedulable = *input.Spec.Unschedulable
		}
		if len(input.Spec.RemoveTaints)+len(input.Spec.AddTaints) > 0 {
			removeT := make(map[string]bool)
			for _, each := range input.Spec.RemoveTaints {
				removeT[each.Key] = true
			}

			taintM := make(map[string]corev1.Taint)
			for _, each := range val.Spec.Taints {
				if !removeT[each.Key] {
					taintM[each.Key] = each
				}
			}

			for _, each := range input.Spec.AddTaints {
				taint := corev1.Taint{
					Key:    each.Key,
					Effect: corev1.TaintEffect(each.Effect),
				}
				if each.Value != nil {
					taint.Value = *each.Value
				}
				taintM[each.Key] = taint
			}

			var taints []corev1.Taint
			for i := range taintM {
				taints = append(taints, taintM[i])
			}
			val.Spec.Taints = taints
		}
	}

	obj, err = patchNodeLabelsAndSpec(wc, input.Name, val)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			return nil, errObjectNotFound(ctx, "input.name", input.Name)
		}
		return nil, errGatewayError(ctx, err)
	}
	node := &model.Node{Cluster: &model.Cluster{ID: input.Cluster}}
	convertNode(node, obj)
	node.Status.Name = node.Name
	node.Status.Cluster = input.Cluster
	return node, nil
}

// Cluster is the resolver for the cluster field.
func (r *nodeResolver) Cluster(ctx context.Context, obj *model.Node) (*model.Cluster, error) {
	if obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Cluster(WithGuardSilenced(ctx), obj.Cluster.ID)
}

// Usage is the resolver for the usage field.
func (r *nodeStatusResolver) Usage(ctx context.Context, obj *model.NodeStatus) (map[string]string, error) {
	c := &model1.Cluster{ID: obj.Cluster}
	if err := mapd.Database().Take(c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}

	if c.AuthToken == nil || len(c.AuthToken) == 0 {
		return nil, nil
	}

	block, err := cryptoutil.NewAESGCMCipher(c.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}
	wc, err := k8s.NewKubernetesFromToken(c.Endpoint, block.Decrypt(c.AuthToken), rest.TLSClientConfig{
		Insecure: c.InsecureTLS,
		CAData:   c.CACert,
	})
	if err != nil {
		panic(err)
	}

	metrics, err := getNodeMetrics(k8s.NewMetrics(wc), obj.Name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			return nil, nil
		}
		return nil, errGatewayError(ctx, err)
	}

	usage := make(map[string]string)
	for k, v := range metrics.Usage {
		usage[k.String()] = v.String()

		p := make([]byte, 0, 18)
		number, suffix := v.CanonicalizeBytes(p)
		usage[fmt.Sprintf("%s_number", k)] = string(number)
		usage[fmt.Sprintf("%s_suffix", k)] = string(suffix)
	}
	return usage, nil
}

// Allocated is the resolver for the allocated field.
func (r *nodeStatusResolver) Allocated(ctx context.Context, obj *model.NodeStatus) (*model.NodeAllocatedResources, error) {
	c := &model1.Cluster{ID: obj.Cluster}
	if err := mapd.Database().Take(c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}

	if c.AuthToken == nil || len(c.AuthToken) == 0 {
		return nil, nil
	}

	block, err := cryptoutil.NewAESGCMCipher(c.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}
	wc, err := k8s.NewKubernetesFromToken(c.Endpoint, block.Decrypt(c.AuthToken), rest.TLSClientConfig{
		Insecure: c.InsecureTLS,
		CAData:   c.CACert,
	})
	if err != nil {
		panic(err)
	}

	pods, err := getPods(wc, "", k8s.FieldEqual("spec.nodeName", obj.Name), k8s.FieldNotEqual("status.phase", "Failed"), k8s.FieldNotEqual("status.phase", "Succeeded"))
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}

	resources := &model.NodeAllocatedResources{
		Limits:   make(map[string]string),
		Requests: make(map[string]string),
	}
	reqs, limits := getPodsTotalRequestsAndLimits(pods)
	for k, v := range limits {
		resources.Limits[k.String()] = v.String()

		p := make([]byte, 0, 18)
		number, suffix := v.CanonicalizeBytes(p)
		resources.Limits[fmt.Sprintf("%s_number", k)] = string(number)
		resources.Limits[fmt.Sprintf("%s_suffix", k)] = string(suffix)
	}
	for k, v := range reqs {
		resources.Requests[k.String()] = v.String()

		p := make([]byte, 0, 18)
		number, suffix := v.CanonicalizeBytes(p)
		resources.Requests[fmt.Sprintf("%s_number", k)] = string(number)
		resources.Requests[fmt.Sprintf("%s_suffix", k)] = string(suffix)
	}
	return resources, nil
}

// Nodes is the resolver for the Nodes field.
func (r *queryResolver) Nodes(ctx context.Context, filters []*model.QueryFilter) ([]*model.Node, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	var (
		clusters  []model1.Cluster
		nodes     []*model.Node
		selectors []k8s.Selector
	)

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectClusterResource)
	}
	for _, each := range filters {
		switch each.Key {
		case "name":
			if len(each.Value) > 0 {
				selectors = append(selectors, k8s.FieldEqual("metadata.name", each.Value[0]))
			}
		case "cluster":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`id` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "organization":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`organization_id` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		}
	}
	if err := db.Where("`provisioned` >= ?", model1.ClusterProvisionFinish).Find(&clusters).Error; err != nil {
		panic(&ErrDBConnection{err})
	}

	handle := func(outChan chan<- []*model.Node, gqlErrChan chan<- *gqlerror.Error, errChan chan<- error, inChan <-chan *model1.Cluster) {
		defer func() {
			if r := recover(); r != nil {
				if e, ok := r.(error); ok {
					errChan <- e
				}
			}
		}()

		for c := range inChan {
			if c == nil {
				continue
			}

			block, err := cryptoutil.NewAESGCMCipher(c.ID[:]) // AES-128
			if err != nil {
				panic(err)
			}
			wc, err := k8s.NewKubernetesFromToken(c.Endpoint, block.Decrypt(c.AuthToken), rest.TLSClientConfig{
				Insecure: c.InsecureTLS,
				CAData:   c.CACert,
			})
			if err != nil {
				panic(err)
			}

			objs, err := getNodes(wc, selectors...)
			if err != nil {
				gqlErrChan <- errGatewayError(ctx, err)
				continue
			}

			var nodes []*model.Node
			for i := range objs {
				node := &model.Node{Cluster: &model.Cluster{ID: c.ID}}
				convertNode(node, &objs[i])
				node.Status.Name = node.Name
				node.Status.Cluster = c.ID
				nodes = append(nodes, node)
			}
			outChan <- nodes
		}
	}

	inChan := make(chan *model1.Cluster, len(clusters))
	for i := range clusters {
		inChan <- &clusters[i]
	}
	close(inChan)
	outChan := make(chan []*model.Node, 1)
	defer close(outChan)
	gqlErrChan := make(chan *gqlerror.Error, 1)
	defer close(gqlErrChan)
	errChan := make(chan error, 1)
	defer close(errChan)

	for i := 0; i < runtime.NumCPU(); i++ {
		go handle(outChan, gqlErrChan, errChan, inChan)
	}

	var err error
	for i := 0; i < len(clusters); i++ {
		select {
		case out := <-outChan:
			nodes = append(nodes, out...)
		case e := <-gqlErrChan:
			graphql.AddError(ctx, e)
		case e := <-errChan:
			if e != nil {
				err = multierr.Append(err, e)
			}
		}
	}
	if err != nil {
		panic(err)
	}
	return nodes, nil
}

// DrainNode is the resolver for the DrainNode field.
func (r *subscriptionResolver) DrainNode(ctx context.Context, input model.DrainNode) (<-chan *model.DrainNodeEvent, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadWrite, model1.PolicyObjectClusterResource)
	}

	c := &model1.Cluster{ID: input.Cluster}
	if err := db.Take(c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "input.cluster", c.ID.String())
		}
		panic(&ErrDBConnection{err})
	}

	if c.AuthToken == nil || len(c.AuthToken) == 0 {
		return nil, errInvalidRequest(ctx, "input.cluster", "cluster is not initialized", input.Cluster.String())
	}

	block, err := cryptoutil.NewAESGCMCipher(c.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}
	wc, err := k8s.NewKubernetesFromToken(c.Endpoint, block.Decrypt(c.AuthToken), rest.TLSClientConfig{
		Insecure: c.InsecureTLS,
		CAData:   c.CACert,
	})
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(wc.Config)
	if err != nil {
		panic(err)
	}

	var timeout time.Duration
	if input.Timeout != nil && *input.Timeout > time.Minute {
		timeout = *input.Timeout
	}
	if timeout == 0 {
		timeout = time.Minute * 15
	}
	outChan := make(chan *model.DrainNodeEvent, 16)
	go cordonAndDrainNode(ctx, clientset, input.Name, timeout, outChan)
	return outChan, nil
}

// Node returns graphql1.NodeResolver implementation.
func (r *Resolver) Node() graphql1.NodeResolver { return &nodeResolver{r} }

// NodeStatus returns graphql1.NodeStatusResolver implementation.
func (r *Resolver) NodeStatus() graphql1.NodeStatusResolver { return &nodeStatusResolver{r} }

type nodeResolver struct{ *Resolver }
type nodeStatusResolver struct{ *Resolver }
