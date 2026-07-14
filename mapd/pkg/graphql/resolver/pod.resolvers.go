package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"bufio"
	"context"
	"errors"
	"io"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	graphql1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"go.uber.org/multierr"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

// Instance is the resolver for the instance field.
func (r *podResolver) Instance(ctx context.Context, obj *model.Pod) (*model.Instance, error) {
	if obj.Instance == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Instance(WithGuardSilenced(ctx), obj.Instance.ID)
}

// Component is the resolver for the component field.
func (r *podResolver) Component(ctx context.Context, obj *model.Pod) (*model.Component, error) {
	if obj.Component == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Component(WithGuardSilenced(ctx), obj.Component.ID)
}

// Application is the resolver for the application field.
func (r *podResolver) Application(ctx context.Context, obj *model.Pod) (*model.Application, error) {
	if obj.Application == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Application(WithGuardSilenced(ctx), obj.Application.ID)
}

// Cluster is the resolver for the cluster field.
func (r *podResolver) Cluster(ctx context.Context, obj *model.Pod) (*model.Cluster, error) {
	if obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Cluster(WithGuardSilenced(ctx), obj.Cluster.ID)
}

// Pod is the resolver for the Pod field.
func (r *queryResolver) Pod(ctx context.Context, cluster uuid.UUID, namespace string, name string) (*model.Pod, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if name == "" {
		return nil, errInvalidRequest(ctx, "name", "value cannot be empty")
	}
	if namespace == "" {
		return nil, errInvalidRequest(ctx, "namespace", "value cannot be empty")
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectClusterResource)
	}

	c := &model1.Cluster{ID: cluster}
	if err := db.Take(c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "cluster", cluster.String())
		}
		panic(&ErrDBConnection{err})
	}

	if c.Provisioned < model1.ClusterProvisionFinish {
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

	obj, err := wc.GetPod(namespace, name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			if IsNotFoundOmitted(ctx) {
				return nil, nil
			}
			return nil, errObjectNotFound(ctx, "name", name)
		}
		return nil, errGatewayError(ctx, err)
	}
	pod := &model.Pod{Cluster: &model.Cluster{ID: cluster}}
	convertPod(pod, obj)
	return pod, nil
}

// PodLog is the resolver for the PodLog field.
func (r *queryResolver) PodLog(ctx context.Context, cluster uuid.UUID, namespace string, name string, container *string) (*model.PodLog, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectClusterResource)
	}

	c := &model1.Cluster{ID: cluster}
	if err := db.Take(c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "cluster", cluster.String())
		}
		panic(&ErrDBConnection{err})
	}

	if c.Provisioned < model1.ClusterProvisionFinish {
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

	var containerName string
	if container != nil {
		containerName = *container
	}

	str, err := wc.GetPodLog(namespace, name, containerName)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			if IsNotFoundOmitted(ctx) {
				return nil, nil
			}
			return nil, errObjectNotFound(ctx, "name", name)
		}
		return nil, errGatewayError(ctx, err)
	}
	return &model.PodLog{
		Name:      name,
		Namespace: namespace,
		Container: containerName,
		Message:   string(str),
		Timestamp: time.Now(),
	}, nil
}

// SmartPodLog is the resolver for the SmartPodLog field.
func (r *queryResolver) SmartPodLog(ctx context.Context, cluster uuid.UUID, namespace string, controller string, issuer string) (*model.PodLog, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectClusterResource)
	}

	c := &model1.Cluster{ID: cluster}
	if err := db.Take(c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "cluster", cluster.String())
		}
		panic(&ErrDBConnection{err})
	}

	if c.Provisioned < model1.ClusterProvisionFinish {
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

	var pods []corev1.Pod
	switch controller {
	case "job", "binary_build", "dockerfile_build":
		job, err := wc.GetJob(namespace, issuer)
		if err != nil {
			if k8s.IsErrObjectNotExists(err) {
				return nil, errObjectNotFound(ctx, "issuer", issuer)
			}
			return nil, errGatewayError(ctx, err)
		}
		pods, err = getPods(wc, job.Namespace, k8s.LabelEqual("controller-uid", string(job.UID)))
		if err != nil {
			return nil, errGatewayError(ctx, err)
		}
	case "source_build":
		build, err := k8s.NewKpack(wc).GetBuild(namespace, issuer)
		if err != nil {
			if k8s.IsErrObjectNotExists(err) {
				return nil, errObjectNotFound(ctx, "issuer", issuer)
			}
			return nil, errGatewayError(ctx, err)
		}
		pods, err = getPods(wc, build.Namespace, k8s.LabelEqual("kpack.io/build", build.Name))
		if err != nil {
			return nil, errGatewayError(ctx, err)
		}
	case "deployment":
		deployment, err := wc.GetDeployment(namespace, issuer)
		if err != nil {
			if k8s.IsErrObjectNotExists(err) {
				return nil, errObjectNotFound(ctx, "issuer", issuer)
			}
			return nil, errGatewayError(ctx, err)
		}
		pods, err = getPods(wc, deployment.GetNamespace(), k8s.FromLabelSelector(deployment.Spec.Selector)...)
		if err != nil {
			return nil, errGatewayError(ctx, err)
		}
	}
	if len(pods) < 1 {
		return nil, nil
	}

	sort.Slice(pods, func(i, j int) bool { return pods[i].GetCreationTimestamp().After(pods[j].GetCreationTimestamp().Time) })
	pod := &pods[0] // latest pod

	var latestContainer string
	for _, each := range pod.Status.InitContainerStatuses {
		if !each.Ready {
			latestContainer = each.Name
			break
		}
	}
	if latestContainer == "" {
		for i, each := range pod.Status.ContainerStatuses {
			if !each.Ready || i == len(pod.Status.ContainerStatuses)-1 {
				latestContainer = each.Name
				break
			}
		}
	}

	str, err := wc.GetPodLog(namespace, pod.Name, latestContainer)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			if IsNotFoundOmitted(ctx) {
				return nil, nil
			}
			return nil, errObjectNotFound(ctx, "issuer", issuer)
		}
		return nil, errGatewayError(ctx, err)
	}
	return &model.PodLog{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Container: latestContainer,
		Message:   string(str),
		Timestamp: time.Now(),
	}, nil
}

// Pods is the resolver for the Pods field.
func (r *queryResolver) Pods(ctx context.Context, filters []*model.QueryFilter) ([]*model.Pod, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	var (
		clusters  []model1.Cluster
		pods      []*model.Pod
		selectors []k8s.Selector

		controller, issuer string
		ns                 string
	)

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectClusterResource)
	}
LOOP:
	for _, each := range filters {
		switch each.Key {
		case "name":
			if len(each.Value) > 0 {
				selectors = append(selectors, k8s.FieldEqual("metadata.name", each.Value[0]))
			}
		case "namespace":
			if len(each.Value) > 0 {
				selectors = append(selectors, k8s.FieldEqual("metadata.namespace", each.Value[0]))
				ns = each.Value[0]
			}
		case "instance":
			switch len(each.Value) {
			case 0:
				selectors = append(selectors, k8s.LabelExists(defLabelInstance))
			case 1:
				selectors = append(selectors, k8s.LabelEqual(defLabelInstance, each.Value[0]))
			default:
				selectors = append(selectors, k8s.LabelIn(defLabelInstance, each.Value...))
			}
		case "component":
			switch len(each.Value) {
			case 0:
				selectors = append(selectors, k8s.LabelExists(defLabelComponent))
			case 1:
				selectors = append(selectors, k8s.LabelEqual(defLabelComponent, each.Value[0]))
			default:
				selectors = append(selectors, k8s.LabelIn(defLabelComponent, each.Value...))
			}
		case "application":
			switch len(each.Value) {
			case 0:
				selectors = append(selectors, k8s.LabelExists(defLabelApplication))
			case 1:
				selectors = append(selectors, k8s.LabelEqual(defLabelApplication, each.Value[0]))
			default:
				selectors = append(selectors, k8s.LabelIn(defLabelApplication, each.Value...))
			}
		case "controller":
			if len(each.Value) > 0 {
				switch each.Value[0] {
				case "job", "binary_build", "dockerfile_build", "source_build", "deployment":
				default:
					continue LOOP
				}
				controller = each.Value[0]
			}
		case "issuer":
			if len(each.Value) > 0 {
				issuer = each.Value[0]
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

	handle := func(outChan chan<- []*model.Pod, gqlErrChan chan<- *gqlerror.Error, errChan chan<- error, inChan <-chan *model1.Cluster) {
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

			if controller != "" && issuer != "" {
				if ns == "" {
					ns = "default"
				}

				switch controller {
				case "job", "binary_build", "dockerfile_build":
					job, err := wc.GetJob(ns, issuer)
					if err != nil {
						if k8s.IsErrObjectNotExists(err) {
							goto DONE
						}
						gqlErrChan <- errGatewayError(ctx, err)
						continue
					}
					objs, err := getPods(wc, job.Namespace, k8s.LabelEqual("controller-uid", string(job.UID)))
					if err != nil {
						gqlErrChan <- errGatewayError(ctx, err)
						continue
					}
					for i := range objs {
						pod := &model.Pod{Cluster: &model.Cluster{ID: c.ID}}
						convertPod(pod, &objs[i])
						pods = append(pods, pod)
					}
				case "source_build":
					build, err := k8s.NewKpack(wc).GetBuild(ns, issuer)
					if err != nil {
						if k8s.IsErrObjectNotExists(err) {
							goto DONE
						}
						gqlErrChan <- errGatewayError(ctx, err)
						continue
					}
					objs, err := getPods(wc, build.Namespace, k8s.LabelEqual("kpack.io/build", build.Name))
					if err != nil {
						gqlErrChan <- errGatewayError(ctx, err)
						continue
					}
					for i := range objs {
						pod := &model.Pod{Cluster: &model.Cluster{ID: c.ID}}
						convertPod(pod, &objs[i])
						pods = append(pods, pod)
					}
				case "deployment":
					deployment, err := wc.GetDeployment(ns, issuer)
					if err != nil {
						if k8s.IsErrObjectNotExists(err) {
							goto DONE
						}
						gqlErrChan <- errGatewayError(ctx, err)
						continue
					}
					objs, err := getPods(wc, deployment.GetNamespace(), k8s.FromLabelSelector(deployment.Spec.Selector)...)
					if err != nil {
						gqlErrChan <- errGatewayError(ctx, err)
						continue
					}
					for i := range objs {
						pod := &model.Pod{Cluster: &model.Cluster{ID: c.ID}}
						convertPod(pod, &objs[i])
						pods = append(pods, pod)
					}
				}
			} else {
				objs, err := getPods(wc, "", selectors...)
				if err != nil {
					gqlErrChan <- errGatewayError(ctx, err)
					continue
				}
				for i := range objs {
					pod := &model.Pod{Cluster: &model.Cluster{ID: c.ID}}
					convertPod(pod, &objs[i])
					pods = append(pods, pod)
				}
			}
		DONE:
			outChan <- pods
		}
	}

	inChan := make(chan *model1.Cluster, len(clusters))
	for i := range clusters {
		inChan <- &clusters[i]
	}
	close(inChan)
	outChan := make(chan []*model.Pod, 1)
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
			pods = append(pods, out...)
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
	return pods, nil
}

// PodLog is the resolver for the PodLog field.
func (r *subscriptionResolver) PodLog(ctx context.Context, cluster uuid.UUID, namespace string, name string, container string) (<-chan *model.PodLog, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectClusterResource)
	}

	c := &model1.Cluster{ID: cluster}
	if err := db.Take(c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "cluster", cluster.String())
		}
		panic(&ErrDBConnection{err})
	}

	if c.Provisioned < model1.ClusterProvisionFinish {
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

	rc, err := wc.WatchPodLog(ctx, namespace, name, container)
	if err != nil {
		return nil, err
	}

	outChan := make(chan *model.PodLog, 16)
	go func() {
		defer rc.Close()

		buf := bufio.NewReader(rc)
		for {
			line, isPrefix, err := buf.ReadLine()
			if !isPrefix {
				line = append(line, '\n')
			}
			outChan <- &model.PodLog{
				Name:      name,
				Namespace: namespace,
				Container: container,
				Message:   string(line),
				Timestamp: time.Now(),
			}

			if err == io.EOF {
				break
			}
			if err != nil {
				// TODO: add details logger here.
				break
			}
		}

		close(outChan)
	}()
	return outChan, nil
}

// PodExec is the resolver for the PodExec field.
func (r *subscriptionResolver) PodExec(ctx context.Context, cluster uuid.UUID, namespace string, name string, container string, command []string) (<-chan *model.PodExec, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermModify, model1.PolicyObjectClusterResource)
	}

	c := &model1.Cluster{ID: cluster}
	if err := db.Take(c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "cluster", cluster.String())
		}
		panic(&ErrDBConnection{err})
	}

	if c.Provisioned < model1.ClusterProvisionFinish {
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

	stdin, out, done, err := wc.WatchPodExec(ctx, namespace, name, container, command...)
	if err != nil {
		return nil, err
	}
	session := r.Resolver.PodExecSubscription.New(stdin, out)

	outChan := make(chan *model.PodExec, 16)
	go func() {
		select {
		case <-done:
			r.Resolver.PodExecSubscription.Delete(session.ID)
			return
		default:
		}

		buf := bufio.NewReader(session.Out)
		for {
			line, isPrefix, err := buf.ReadLine()
			if !isPrefix {
				line = append(line, '\n')
			}
			outChan <- &model.PodExec{ID: session.ID, Message: string(line), Timestamp: time.Now()}

			if err == io.EOF {
				break
			}
			if err != nil {
				// TODO: add details logger here.
				break
			}
		}

		close(outChan)
	}()
	return outChan, nil
}

// Pod returns graphql1.PodResolver implementation.
func (r *Resolver) Pod() graphql1.PodResolver { return &podResolver{r} }

type podResolver struct{ *Resolver }
