package resolver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/license"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/vks"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func vksProviderLicenseStatus(ctx context.Context, db *gorm.DB, cp *model1.ClusterProvider, probe bool) (*model.VKSProviderLicenseStatus, error) {
	if cp.Class != model1.ClusterClassVKS {
		return nil, errInvalidRequest(ctx, "providerID", "provider must be VKS")
	}
	spec := new(model1.ClusterProviderVKSSpec)
	if err := json.Unmarshal(cp.Spec, spec); err != nil {
		return nil, errGatewayError(ctx, err)
	}
	uid := spec.SupervisorUID
	connectionWarning := ""
	if probe {
		client, _, err := vks.NewClientFromProvider(cp.ID, spec)
		if err != nil {
			if uid == "" {
				return nil, errGatewayError(ctx, err)
			}
			connectionWarning = "; Supervisor connection unavailable"
		} else {
			liveUID, probeErr := client.SupervisorUID(ctx)
			if probeErr != nil {
				if uid == "" {
					return nil, errGatewayError(ctx, probeErr)
				}
				connectionWarning = "; Supervisor connection unavailable"
			} else if spec.SupervisorUID != "" && liveUID != spec.SupervisorUID {
				return &model.VKSProviderLicenseStatus{SupervisorUID: liveUID, Message: "Supervisor identity changed; create a new provider"}, nil
			} else {
				uid = liveUID
			}
		}
	}
	status := &model.VKSProviderLicenseStatus{SupervisorUID: uid, Message: "VKS license not installed"}
	if uid == "" {
		status.Message = "Supervisor identity unavailable"
		return status, nil
	}
	key, err := license.PublicKey()
	if err != nil {
		status.Message = "VKS license verifier is not configured"
		return status, nil
	}
	var stored model1.VKSSupervisorLicense
	if err := db.Where("supervisor_uid = ?", uid).Take(&stored).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status.Message += connectionWarning
			return status, nil
		}
		return nil, errGatewayError(ctx, err)
	}
	payload, err := license.VerifyVKS(stored.Token, key, uid, time.Now())
	if err != nil {
		status.Message = err.Error()
		status.Message += connectionWarning
		return status, nil
	}
	expiry, _ := time.Parse(time.RFC3339, payload.ExpiresAt)
	status.Licensed, status.ExpiresAt, status.Message = true, &expiry, "VKS license active"+connectionWarning
	return status, nil
}

func requireVKSProviderLicense(ctx context.Context, db *gorm.DB, cp *model1.ClusterProvider) error {
	spec := new(model1.ClusterProviderVKSSpec)
	if err := json.Unmarshal(cp.Spec, spec); err != nil {
		return errGatewayError(ctx, err)
	}
	client, _, err := vks.NewClientFromProvider(cp.ID, spec)
	if err != nil {
		return errGatewayError(ctx, err)
	}
	liveUID, err := client.SupervisorUID(ctx)
	if err != nil {
		return errGatewayError(ctx, err)
	}
	if spec.SupervisorUID != "" && liveUID != spec.SupervisorUID {
		return newGQLErrorf(ctx, CodeInsufficientPrivileges, "VKS Supervisor identity changed")
	}
	verified := *cp
	verifiedSpec := *spec
	verifiedSpec.SupervisorUID = liveUID
	verified.Spec, err = json.Marshal(&verifiedSpec)
	if err != nil {
		return errGatewayError(ctx, err)
	}
	status, err := vksProviderLicenseStatus(ctx, db, &verified, false)
	if err != nil {
		return err
	}
	if !status.Licensed {
		return newGQLErrorf(ctx, CodeInsufficientPrivileges, "VKS Supervisor license required: %s", status.Message)
	}
	return nil
}

func (r *queryResolver) VKSProviderLicense(ctx context.Context, providerID uuid.UUID) (*model.VKSProviderLicenseStatus, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}
	if providerID == uuid.Nil {
		return nil, errInvalidRequest(ctx, "providerID", "ID cannot be empty")
	}
	var cp model1.ClusterProvider
	if err := mapd.Database().Where("id = ?", providerID).Take(&cp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "providerID", providerID.String())
		}
		return nil, errGatewayError(ctx, err)
	}
	if !sess.Can(model1.PermReadOnly, model1.PolicyObjectClusterProvider, cp.OrganizationID.UUID) {
		return nil, errInsufficientPrivileges(ctx)
	}
	return vksProviderLicenseStatus(ctx, mapd.Database(), &cp, true)
}

func (r *mutationResolver) UploadVKSProviderLicense(ctx context.Context, providerID uuid.UUID, token string) (*model.VKSProviderLicenseStatus, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}
	if !sess.Admin {
		return nil, errInsufficientPrivileges(ctx)
	}
	if providerID == uuid.Nil {
		return nil, errInvalidRequest(ctx, "providerID", "ID cannot be empty")
	}
	if len(token) == 0 || len(token) > license.MaxTokenBytes {
		return nil, errInvalidRequest(ctx, "license", "invalid license length")
	}
	var cp model1.ClusterProvider
	if err := mapd.Database().Where("id = ?", providerID).Take(&cp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "providerID", providerID.String())
		}
		return nil, errGatewayError(ctx, err)
	}
	if cp.Class != model1.ClusterClassVKS {
		return nil, errInvalidRequest(ctx, "providerID", "provider must be VKS")
	}
	spec := new(model1.ClusterProviderVKSSpec)
	if err := json.Unmarshal(cp.Spec, spec); err != nil {
		return nil, errGatewayError(ctx, err)
	}
	client, _, err := vks.NewClientFromProvider(cp.ID, spec)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}
	uid, err := client.SupervisorUID(ctx)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}
	if spec.SupervisorUID != "" && spec.SupervisorUID != uid {
		return nil, errInvalidRequest(ctx, "providerID", "Supervisor identity changed; create a new provider")
	}
	key, err := license.PublicKey()
	if err != nil {
		return nil, errInvalidRequest(ctx, "license", err.Error())
	}
	if _, err := license.VerifyVKS(token, key, uid, time.Now()); err != nil {
		return nil, errInvalidRequest(ctx, "license", err.Error())
	}
	spec.SupervisorUID = uid
	b, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("encode provider: %w", err)
	}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		var current model1.ClusterProvider
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", cp.ID).Take(&current).Error; err != nil {
			return err
		}
		if !bytes.Equal(current.Spec, cp.Spec) {
			return fmt.Errorf("VKS provider changed during license upload; retry")
		}
		if err := tx.Model(&current).Where("id = ?", cp.ID).Update("spec", b).Error; err != nil {
			return err
		}
		if err := tx.Save(&model1.VKSSupervisorLicense{SupervisorUID: uid, Token: token}).Error; err != nil {
			return err
		}
		return appendAuditLog(tx, "update", "ClusterProvider", cp.ID, cp.Name, sess.UserID, sess.UserName, cp.OrganizationID.UUID)
	}); err != nil {
		return nil, errGatewayError(ctx, err)
	}
	return vksProviderLicenseStatus(ctx, mapd.Database(), &cp, true)
}
