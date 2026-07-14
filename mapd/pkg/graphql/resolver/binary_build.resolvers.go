package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	graphql1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/abstract"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/randutil"
	"gorm.io/gorm"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// Cluster is the resolver for the cluster field.
func (r *binaryBuildResolver) Cluster(ctx context.Context, obj *model.BinaryBuild) (*model.Cluster, error) {
	if obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Cluster(WithGuardSilenced(ctx), obj.Cluster.ID)
}

// CreateBinaryBuild is the resolver for the CreateBinaryBuild field.
func (r *mutationResolver) CreateBinaryBuild(ctx context.Context, input model.CreateBinaryBuild) (*model.BinaryBuild, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: defNamespace,
			Labels: map[string]string{
				defLabelController: "binary_build",
				defLabelImage:      input.Image,
				defLabelManaged:    "true",
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:          util.Int32Ptr(3),
			Completions:           util.Int32Ptr(1),
			Parallelism:           util.Int32Ptr(1),
			ActiveDeadlineSeconds: util.Int64Ptr(21600), // 6h
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						defLabelController: "binary_build",
						defLabelImage:      input.Image,
						defLabelManaged:    "true",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: "Never",
					Volumes: []corev1.Volume{{
						Name: "docker-registry-secret",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								DefaultMode: util.Int32Ptr(420),
								Items: []corev1.KeyToPath{{
									Key:  ".dockerconfigjson",
									Path: "config.json",
								}},
								SecretName: defRegistrySecret,
							},
						},
					}},
				},
			},
		},
	}
	var obj *model.Job

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx
		if !IsGuardSilenced(ctx) {
			db = sess.FilterQueryByOrganization(tx, model1.PermReadWrite, model1.PolicyObjectClusterResource)
		}
		cluster := &model1.Cluster{ID: input.Cluster}
		if err := db.Take(cluster).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.cluster", cluster.ID.String())
			}
			panic(&ErrDBConnection{err})
		}
		if cluster.Provisioned < model1.ClusterProvisionFinish {
			return errInvalidRequest(ctx, "input.cluster", "cluster is not ready", input.Cluster.String())
		}

		registry := &model1.ImageRegistry{}
		// use the original session because it is tainted by last query for some reason.
		if err := tx.Where("`id` IN (SELECT `image_registry_id` FROM `cluster_image_registry` WHERE `cluster_id` = ? AND `default` = ?)", cluster.ID, true).Take(registry).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		block, err := cryptoutil.NewAESGCMCipher(cluster.ID[:]) // AES-128
		if err != nil {
			panic(err)
		}
		wc, err := k8s.NewKubernetesFromToken(cluster.Endpoint, block.Decrypt(cluster.AuthToken), rest.TLSClientConfig{
			Insecure: cluster.InsecureTLS,
			CAData:   cluster.CACert,
		})
		if err != nil {
			panic(err)
		}
		cli := k8s.NewMap(wc)

		image, err := cli.GetBinaryImage(defNamespace, input.Image)
		if err != nil {
			if k8s.IsErrObjectNotExists(err) {
				return errObjectNotFound(ctx, "input.image", input.Image)
			}
			return errGatewayError(ctx, err)
		}

		u := abstract.MustParseURL(registry.Endpoint)
		imageTag := "latest"
		if mapd.Version != "" {
			imageTag = mapd.Version
		}
		builderImage := filepath.Join(u.Host, mapd.GlobalConfig.ClusterImageProject, fmt.Sprintf("%s:%s", mapd.GlobalConfig.EmbeddedBuilderImage, imageTag))
		baseImage, _ := ParseImage(image.Spec.BaseImage)
		if baseImage.Registry == "" {
			baseImage.Registry = u.Host
		}
		distImage, _ := ParseImage(image.Spec.DistImage)
		if distImage.Registry == "" {
			distImage.Registry = u.Host
		}
		if distImage.Project == "" {
			distImage.Project = "library"
		}

		jobs, err := getJobs(wc, defNamespace, k8s.LabelEqual(defLabelController, "binary_build"), k8s.LabelEqual(defLabelIssuer, image.Name))
		if err != nil {
			return errGatewayError(ctx, err)
		}

		var objs []*model.BinaryBuild
		for _, job := range jobs {
			obj := &model.Job{Cluster: &model.Cluster{ID: cluster.ID}}
			convertJob(obj, &job)
			objs = append(objs, convertJobToBinaryBuild(obj))
		}
		sort.SliceStable(objs, func(i, j int) bool { return objs[i].Status.Num < objs[j].Status.Num })

		if len(objs) > int(mapd.GlobalConfig.EmbeddedBuildHistoryPreserved) {
			toDelete := objs[:len(objs)-int(mapd.GlobalConfig.EmbeddedBuildHistoryPreserved)]
			for _, each := range toDelete {
				if err := wc.DeleteJob(defNamespace, each.Name); err != nil && !k8s.IsErrObjectNotExists(err) {
					return errGatewayError(ctx, err)
				}
			}
		}

		var num int64
		if len(objs) > 0 {
			num = objs[len(objs)-1].Status.Num + 1
		} else {
			num = 1
		}
		job.Name = fmt.Sprintf("%s-%d-%s", image.Name, num, randutil.StringN(6, "0123456789abcdefghijklmnopqrstuvwxyz"))
		job.Annotations = image.Annotations
		job.Labels[defLabelIssuer] = image.Name
		job.Spec.Template.Labels[defLabelIssuer] = image.Name
		job.Labels[defLabelRevision] = strconv.FormatInt(num, 10)
		job.Spec.Template.Labels[defLabelRevision] = strconv.FormatInt(num, 10)

		container := corev1.Container{
			Name:            "main",
			Image:           builderImage,
			ImagePullPolicy: "Always",
			VolumeMounts: []corev1.VolumeMount{{
				MountPath: "/kaniko/.docker",
				Name:      "docker-registry-secret",
				ReadOnly:  true,
			}},
		}
		env := make(map[string]string)
		env["BUILD_TYPE"] = "Binary"
		env["BUILD_NUM"] = strconv.FormatInt(num, 10)
		env["BINARY_BUILD_BASE_IMAGE_NAME"] = baseImage.String()
		env["BINARY_DEPLOY_DIR"] = image.Spec.BinaryDir
		env["BINARY_FILE_URL"] = image.Spec.URL
		env["BINARY_PROCESS_RUN_UID"] = strconv.FormatInt(image.Spec.ProcessUID, 10)
		env["OUTPUT_IMAGE_REGISTRY"] = distImage.Registry
		env["OUTPUT_IMAGE_REPO"] = distImage.Project
		env["OUTPUT_IMAGE_NAME"] = distImage.Repository
		env["OUTPUT_IMAGE_TAG"] = time.Now().Format("20060102.150405")
		env["INSECURE_BASE_IMAGE"] = "true"
		env["INSECURE_OUTPUT_IMAGE"] = "true"
		for k, v := range env {
			container.Env = append(container.Env, corev1.EnvVar{
				Name:  k,
				Value: v,
			})
		}
		job.Spec.Template.Spec.Containers = append(job.Spec.Template.Spec.Containers, container)

		if err := appendAuditLog(tx, "create", "BinaryBuild", cluster.ID, job.Name, sess.UserID, sess.UserName, cluster.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		job, err = wc.CreateJob(job)
		if err != nil {
			return errGatewayError(ctx, err)
		}
		obj = &model.Job{Cluster: &model.Cluster{ID: cluster.ID}}
		convertJob(obj, job)
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertJobToBinaryBuild(obj), nil
}

// DeleteBinaryBuild is the resolver for the DeleteBinaryBuild field.
func (r *mutationResolver) DeleteBinaryBuild(ctx context.Context, input model.DeleteBinaryBuild) (*model.BinaryBuild, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermModify, model1.PolicyObjectClusterResource)
	}

	c := &model1.Cluster{ID: input.Cluster}
	if err := db.Take(c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "input.cluster", c.ID.String())
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

	obj, err := wc.GetJob(defNamespace, input.Name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			if IsNotFoundOmitted(ctx) {
				return nil, nil
			}
			return nil, errObjectNotFound(ctx, "input.name", input.Name)
		}
		return nil, errGatewayError(ctx, err)
	}
	job := &model.Job{Cluster: &model.Cluster{ID: input.Cluster}}
	convertJob(job, obj)
	if job.Labels == nil {
		return nil, errObjectNotFound(ctx, "input.name", input.Name)
	}
	controller, ok := job.Labels[defLabelController]
	if !ok || controller != "binary_build" {
		return nil, errObjectNotFound(ctx, "input.name", input.Name)
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := appendAuditLog(tx, "delete", "BinaryBuild", c.ID, obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := wc.DeleteJob(defNamespace, input.Name); err != nil && !k8s.IsErrObjectNotExists(err) {
			return errGatewayError(ctx, err)
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}

	return convertJobToBinaryBuild(job), nil
}

// BinaryBuild is the resolver for the BinaryBuild field.
func (r *queryResolver) BinaryBuild(ctx context.Context, cluster uuid.UUID, name string) (*model.BinaryBuild, error) {
	job, err := r.Resolver.Query().(*queryResolver).Job(ctx, cluster, defNamespace, name)
	if err != nil {
		return nil, err
	}
	if job.Labels == nil {
		return nil, errObjectNotFound(ctx, "name", name)
	}
	controller, ok := job.Labels[defLabelController]
	if !ok || controller != "binary_build" {
		return nil, errObjectNotFound(ctx, "name", name)
	}
	return convertJobToBinaryBuild(job), nil
}

// BinaryBuilds is the resolver for the BinaryBuilds field.
func (r *queryResolver) BinaryBuilds(ctx context.Context, filters []*model.QueryFilter) ([]*model.BinaryBuild, error) {
	filter := &model.QueryFilter{Key: "binaryimage"}
	for _, each := range filters {
		switch each.Key {
		case "image":
			filter.Value = each.Value
		}
	}
	jobs, err := r.Resolver.Query().(*queryResolver).Jobs(ctx, append(filters, filter, &model.QueryFilter{Key: "namespace", Value: []string{defNamespace}}))
	if err != nil {
		return nil, err
	}
	var objs []*model.BinaryBuild
	for _, job := range jobs {
		objs = append(objs, convertJobToBinaryBuild(job))
	}
	sort.SliceStable(objs, func(i, j int) bool { return objs[i].LastModified.After(objs[j].LastModified) })
	return objs, nil
}

// BinaryBuild returns graphql1.BinaryBuildResolver implementation.
func (r *Resolver) BinaryBuild() graphql1.BinaryBuildResolver { return &binaryBuildResolver{r} }

type binaryBuildResolver struct{ *Resolver }
