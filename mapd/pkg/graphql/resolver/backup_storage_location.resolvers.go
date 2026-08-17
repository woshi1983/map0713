package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	graphql1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/httputil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"go.uber.org/multierr"
	ini "gopkg.in/ini.v1"
	"gorm.io/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// Backups is the resolver for the backups field.
func (r *backupStorageLocationResolver) Backups(ctx context.Context, obj *model.BackupStorageLocation, filters []*model.QueryFilter) ([]*model.Backup, error) {
	if obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Backups(
		WithGuardSilenced(ctx),
		append(filters,
			&model.QueryFilter{Key: "cluster", Value: []string{obj.Cluster.ID.String()}},
			&model.QueryFilter{Key: "location", Value: []string{obj.Name}},
		),
	)
}

// Restores is the resolver for the restores field.
func (r *backupStorageLocationResolver) Restores(ctx context.Context, obj *model.BackupStorageLocation, filters []*model.QueryFilter) ([]*model.Restore, error) {
	if obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Restores(
		WithGuardSilenced(ctx),
		append(filters,
			&model.QueryFilter{Key: "cluster", Value: []string{obj.Cluster.ID.String()}},
			&model.QueryFilter{Key: "location", Value: []string{obj.Name}},
		),
	)
}

// Cluster is the resolver for the cluster field.
func (r *backupStorageLocationResolver) Cluster(ctx context.Context, obj *model.BackupStorageLocation) (*model.Cluster, error) {
	if obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Cluster(WithGuardSilenced(ctx), obj.Cluster.ID)
}

// CreateBackupStorageLocation is the resolver for the CreateBackupStorageLocation field.
func (r *mutationResolver) CreateBackupStorageLocation(ctx context.Context, input model.CreateBackupStorageLocation) (*model.BackupStorageLocation, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if input.Validate != nil && *input.Validate {
		switch input.Spec.Provider {
		case "aws":
			if input.Spec.Config.Region == nil {
				return nil, errInvalidRequest(ctx, "input.spec.config.region", "value cannot be empty")
			}
			if input.Spec.Config.S3Url == nil {
				return nil, errInvalidRequest(ctx, "input.spec.config.s3Url", "value cannot be empty")
			}
			if input.Spec.ObjectStorage == nil {
				return nil, errInvalidRequest(ctx, "input.spec.objectStorage", "value cannot be empty")
			}
			if input.Spec.ObjectStorage.Bucket == nil {
				return nil, errInvalidRequest(ctx, "input.spec.objectStorage.bucket", "value cannot be empty")
			}
			if input.Spec.Profile.AwsAccessKeyID == nil {
				return nil, errInvalidRequest(ctx, "input.spec.profile.awsAccessKeyID", "value cannot be empty")
			}
			if input.Spec.Profile.AwsSecretAccessKey == nil {
				return nil, errInvalidRequest(ctx, "input.spec.profile.awsSecretAccessKey", "value cannot be empty")
			}

			cfg := &aws.Config{
				Credentials: credentials.NewStaticCredentials(*input.Spec.Profile.AwsAccessKeyID, *input.Spec.Profile.AwsSecretAccessKey, ""),
				Endpoint:    aws.String(input.Spec.Config.S3Url.String()),
				Region:      aws.String(*input.Spec.Config.Region),
				MaxRetries:  aws.Int(3),
				HTTPClient:  httputil.NewClient().Client,
			}
			if input.Spec.Config.S3ForcePathStyle != nil && *input.Spec.Config.S3ForcePathStyle {
				cfg.S3ForcePathStyle = aws.Bool(*input.Spec.Config.S3ForcePathStyle)
			}
			switch input.Spec.Config.S3Url.Scheme {
			case "http":
				cfg.DisableSSL = aws.Bool(true)
			case "https":
				cfg.DisableSSL = aws.Bool(false)
			}
			sess, err := session.NewSession(cfg)
			if err != nil {
				return nil, errGatewayError(ctx, err)
			}

			ctx, cancel := context.WithTimeout(ctx, mapd.GlobalConfig.HTTPClientTimeout)
			defer cancel()

			_, err = s3.New(sess).HeadBucketWithContext(ctx, &s3.HeadBucketInput{Bucket: aws.String(*input.Spec.ObjectStorage.Bucket)})
			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					switch aerr.Code() {
					case s3.ErrCodeNoSuchBucket:
						return nil, errObjectNotFound(ctx, "input.spec.objectStorage.bucket", *input.Spec.ObjectStorage.Bucket)
					}
				}
				return nil, errGatewayError(ctx, err)
			}
		}
	}

	if !validateName(input.Name) {
		return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), input.Name)
	}

	c, caCerts, err := getClusterWithTrustedCACertsForCreateBackupStorageLocation(ctx, input.Cluster)
	if err != nil {
		return nil, err
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
	velero := k8s.NewVelero(wc)

	objs, err := getBackupStorageLocations(velero, defNamespace)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}
	preserved := make(map[string]bool)
	for _, each := range objs {
		preserved[standardizeString(each.Name)] = true
	}
	var (
		deleted []string
		cfg     []byte
	)
	secret, err := wc.GetSecret(defNamespace, "cloud-credentials")
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	b, ok := secret.Data["cloud"]
	if ok && b != nil {
		cfg = b
	}
	iniCfg, err := ini.Load(cfg)
	if err != nil {
		panic(err)
	}
LOOP:
	for _, section := range iniCfg.Sections() {
		switch section.Name() {
		case "default", "DEFAULT":
			continue LOOP
		}
		if name := section.Name(); !preserved[name] {
			deleted = append(deleted, name)
		}
	}
	for _, each := range deleted {
		iniCfg.DeleteSection(each)
	}
	def := iniCfg.Section("default")
	def.Comment = "default section MUST be preserved."
	def.Key("aws_access_key_id").SetValue("dummy")
	def.Key("aws_secret_access_key").SetValue("dummy")
	if input.Spec.Profile != nil {
		section := iniCfg.Section(standardizeString(input.Name))
		if input.Spec.Profile.AwsAccessKeyID != nil {
			section.Key("aws_access_key_id").SetValue(*input.Spec.Profile.AwsAccessKeyID)
		}
		if input.Spec.Profile.AwsSecretAccessKey != nil {
			section.Key("aws_secret_access_key").SetValue(*input.Spec.Profile.AwsSecretAccessKey)
		}
	}
	var buf bytes.Buffer
	if _, err := iniCfg.WriteTo(&buf); err != nil {
		panic(err)
	}
	buf.WriteByte('\n')
	secret.Data["cloud"] = buf.Bytes()
	if input.Dryrun == nil || !*input.Dryrun {
		if _, err = wc.UpdateSecret(secret); err != nil {
			return nil, errGatewayError(ctx, err)
		}
	}

	obj := &velerov1.BackupStorageLocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      input.Name,
			Namespace: defNamespace,
			Labels: map[string]string{
				defLabelManaged: "true",
			},
		},
		Spec: velerov1.BackupStorageLocationSpec{
			Provider: input.Spec.Provider,
			Config: map[string]string{
				"profile": standardizeString(input.Name),
			},
		},
	}
	if input.Spec.Config.Region != nil {
		obj.Spec.Config["region"] = *input.Spec.Config.Region
	}
	if input.Spec.Config.S3ForcePathStyle != nil {
		obj.Spec.Config["s3ForcePathStyle"] = strconv.FormatBool(*input.Spec.Config.S3ForcePathStyle)
	}
	if input.Spec.Config.S3Url != nil {
		obj.Spec.Config["s3Url"] = input.Spec.Config.S3Url.String()
	}

	if input.Spec.ObjectStorage != nil {
		var b bytes.Buffer
		b.Write(caCerts)
		b.WriteRune('\n')

		location := &velerov1.ObjectStorageLocation{}
		if input.Spec.ObjectStorage.Bucket != nil {
			location.Bucket = *input.Spec.ObjectStorage.Bucket
		}
		if input.Spec.ObjectStorage.Prefix != nil {
			location.Prefix = *input.Spec.ObjectStorage.Prefix
		}
		if input.Spec.ObjectStorage.CaCert != nil {
			b.WriteString(*input.Spec.ObjectStorage.CaCert)
		}
		if p := b.Bytes(); len(bytes.TrimSpace(p)) > 0 {
			location.CACert = p
		}
		obj.Spec.ObjectStorage = location
	}
	switch input.Spec.AccessMode {
	case model.BackupStorageLocationAccessModeReadWrite:
		obj.Spec.AccessMode = velerov1.BackupStorageLocationAccessModeReadWrite
	case model.BackupStorageLocationAccessModeReadOnly:
		obj.Spec.AccessMode = velerov1.BackupStorageLocationAccessModeReadOnly
	}
	if input.Dryrun == nil || !*input.Dryrun {
		if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
			if err := appendAuditLog(tx, "create", "BackupStorageLocation", c.ID, obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
				panic(&ErrDBConnection{err})
			}

			obj, err = velero.CreateBackupStorageLocation(obj)
			if err != nil {
				if k8s.IsErrObjectAlreadyExists(err) {
					return errObjectDuplicated(ctx, "input.name", input.Name)
				}
				return errGatewayError(ctx, err)
			}
			return nil
		}); err != nil {
			if !IsGQLError(err) {
				panic(err)
			}
			return nil, err
		}
	} else {
		_, err := velero.GetBackupStorageLocation(defNamespace, input.Name)
		if err != nil {
			if !k8s.IsErrObjectNotExists(err) {
				return nil, errGatewayError(ctx, err)
			}
		} else {
			return nil, errObjectDuplicated(ctx, "input.name", input.Name)
		}
	}
	location := &model.BackupStorageLocation{Cluster: &model.Cluster{ID: input.Cluster}}
	convertBackupStorageLocation(location, obj)
	return location, nil
}

// DeleteBackupStorageLocation is the resolver for the DeleteBackupStorageLocation field.
func (r *mutationResolver) DeleteBackupStorageLocation(ctx context.Context, input model.DeleteBackupStorageLocation) (*model.BackupStorageLocation, error) {
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
	velero := k8s.NewVelero(wc)

	obj, err := velero.GetBackupStorageLocation(defNamespace, input.Name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			return nil, errObjectNotFound(ctx, "input.name", input.Name)
		}
		return nil, errGatewayError(ctx, err)
	}
	location := &model.BackupStorageLocation{Cluster: &model.Cluster{ID: input.Cluster}}
	convertBackupStorageLocation(location, obj)

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := appendAuditLog(tx, "delete", "BackupStorageLocation", c.ID, obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := velero.DeleteBackupStorageLocation(defNamespace, input.Name); err != nil && !k8s.IsErrObjectNotExists(err) {
			return errGatewayError(ctx, err)
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}

	objs, err := getBackupStorageLocations(velero, defNamespace)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}
	preserved := make(map[string]bool)
	for _, each := range objs {
		preserved[standardizeString(each.Name)] = true
	}
	var (
		deleted []string
		cfg     []byte
	)
	secret, err := wc.GetSecret(defNamespace, "cloud-credentials")
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	b, ok := secret.Data["cloud"]
	if ok && b != nil {
		cfg = b
	}
	iniCfg, err := ini.Load(cfg)
	if err != nil {
		panic(err)
	}
LOOP:
	for _, section := range iniCfg.Sections() {
		switch section.Name() {
		case "default", "DEFAULT":
			continue LOOP
		}
		if name := section.Name(); !preserved[name] {
			deleted = append(deleted, name)
		}
	}
	for _, each := range deleted {
		iniCfg.DeleteSection(each)
	}
	var buf bytes.Buffer
	if _, err := iniCfg.WriteTo(&buf); err != nil {
		panic(err)
	}
	buf.WriteByte('\n')
	secret.Data["cloud"] = buf.Bytes()
	if _, err = wc.UpdateSecret(secret); err != nil {
		return location, errGatewayError(ctx, err)
	}
	return location, nil
}

// BackupStorageLocation is the resolver for the BackupStorageLocation field.
func (r *queryResolver) BackupStorageLocation(ctx context.Context, cluster uuid.UUID, name string) (*model.BackupStorageLocation, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if name == "" {
		return nil, errInvalidRequest(ctx, "name", "value cannot be empty")
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
	velero := k8s.NewVelero(wc)

	obj, err := velero.GetBackupStorageLocation(defNamespace, name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			if IsNotFoundOmitted(ctx) {
				return nil, nil
			}
			return nil, errObjectNotFound(ctx, "name", name)
		}
		return nil, errGatewayError(ctx, err)
	}
	location := &model.BackupStorageLocation{Cluster: &model.Cluster{ID: cluster}}
	convertBackupStorageLocation(location, obj)
	return location, nil
}

// BackupStorageLocations is the resolver for the BackupStorageLocations field.
func (r *queryResolver) BackupStorageLocations(ctx context.Context, filters []*model.QueryFilter) ([]*model.BackupStorageLocation, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	var (
		clusters  []model1.Cluster
		locations []*model.BackupStorageLocation
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

	handle := func(outChan chan<- []*model.BackupStorageLocation, gqlErrChan chan<- *gqlerror.Error, errChan chan<- error, inChan <-chan *model1.Cluster) {
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
			velero := k8s.NewVelero(wc)

			objs, err := getBackupStorageLocations(velero, defNamespace, selectors...)
			if err != nil {
				gqlErrChan <- errGatewayError(ctx, err)
				continue
			}

			var locations []*model.BackupStorageLocation
			for i := range objs {
				location := &model.BackupStorageLocation{Cluster: &model.Cluster{ID: c.ID}}
				convertBackupStorageLocation(location, &objs[i])
				locations = append(locations, location)
			}
			outChan <- locations
		}
	}

	inChan := make(chan *model1.Cluster, len(clusters))
	for i := range clusters {
		inChan <- &clusters[i]
	}
	close(inChan)
	outChan := make(chan []*model.BackupStorageLocation, 1)
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
			locations = append(locations, out...)
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
	return locations, nil
}

// BackupStorageLocation returns graphql1.BackupStorageLocationResolver implementation.
func (r *Resolver) BackupStorageLocation() graphql1.BackupStorageLocationResolver {
	return &backupStorageLocationResolver{r}
}

type backupStorageLocationResolver struct{ *Resolver }
