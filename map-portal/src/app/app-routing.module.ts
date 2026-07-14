import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';

import { handleObject } from './classes';
import { AuthGuard } from './utils/guard/auth.guard';
import { OrgRolePermissionGuard } from './utils/guard/org-role-permission.guard';
import { FeatureGuard } from './utils/guard/feature.guard';

import { HomeComponent } from './components/base/home/home.component';

import { ManagementComponent } from './components/cluster/management/management.component';
import { OperationComponent } from './components/cluster/operation/operation.component';
import { ClusterdetailComponent } from './components/cluster/clusterdetail/clusterdetail.component';

import { AppcomponentDetailsComponent } from './components/appmgmt/appcomponent-details/appcomponent-details.component';
import { CatalogsComponent } from './components/appmgmt/catalogs/catalogs.component';
import { CatalogDetailsComponent } from './components/appmgmt/catalog-details/catalog-details.component';
import { AppcomponentsComponent } from './components/appmgmt/appcomponents/appcomponents.component';
import { ApproutesComponent } from './components/appmgmt/approutes/approutes.component';
import { IngressComponent } from './components/appmgmt/approutes/ingress/ingress.component';
import { RouteComponent } from './components/appmgmt/approutes/route/route.component';

import { AppvolumesComponent } from './components/appmgmt/appvolumes/appvolumes.component';
import { ConfigmapComponent } from './components/appmgmt/appvolumes/configmap/configmap.component';
import { PvcComponent } from './components/appmgmt/appvolumes/pvc/pvc.component';
import { SecretsComponent } from './components/appmgmt/appvolumes/secrets/secrets.component';
import { StorageclassComponent } from './components/appmgmt/appvolumes/storageclass/storageclass.component';

import { BinaryImageComponent } from './components/appbuild/binary/binary-image/binary-image.component';
import { BinaryImageDetailsComponent } from './components/appbuild/binary/build-image-details/binary-image-details.component';
import { DockerfileImageComponent } from './components/appbuild/dockerfile/dockerfile-image/dockerfile-image.component';
import { DockerfileImageDetailsComponent } from './components/appbuild/dockerfile/dockerfile-image-details/dockerfile-image-details.component';
import { SourceImageComponent } from './components/appbuild/sources/source-image/source-image.component';
import { SourceImageDetailsComponent } from './components/appbuild/sources/source-image-details/source-image-details.component';

import { ProviderComponent } from './components/system/provider/provider.component';
import { ProviderdetailComponent } from './components/system/providerdetail/providerdetail.component';
import { LoadbalancerComponent } from './components/system/loadbalancer/loadbalancer.component';
import { RepositorysComponent } from './components/system/repositorys/repositorys.component';
import { StandardPackageRepositoryComponent } from './components/system/standard-package-repository/standard-package-repository.component';

import { LogoutComponent } from './components/auth/logout/logout.component';
import { UserNoOrgComponent } from './components/base/user-no-org/user-no-org.component';
import { NotFoundComponent } from './utils/shared/not-found/not-found.component';
import { NoPermissionComponent } from './utils/shared/no-permission/no-permission.component';

import { OrganizationComponent } from './components/usermgmt/organization/organization.component';
import { OrgDetailsComponent } from './components/usermgmt/org-details/org-details.component';
import { UsersComponent } from './components/usermgmt/users/users.component';
import { AuditlogComponent } from './components/usermgmt/auditlog/auditlog.component';

import { SummaryComponent } from './components/overview/summary/summary.component';

const routes: Routes = [
  { path: '', redirectTo: 'portal', pathMatch: 'full' },
  { path: 'portal/logout', component: LogoutComponent },
  { path: 'portal/no-org', component: UserNoOrgComponent },
  { path: 'portal/no-permission', component: NoPermissionComponent },
  {
    path: 'portal',
    component: HomeComponent,
    canActivateChild: [AuthGuard],
    children: [
      { path: '', redirectTo: 'cluster/management', pathMatch: 'full' },
      {
        path: 'overview',
        canActivateChild: [OrgRolePermissionGuard],
        data: {
          permissionParam: {
            resource: handleObject.OVERVIEW
          }
        },
        children: [
          { path: '', redirectTo: 'summary', pathMatch: 'full' },
          {
            path: 'summary',
            component: SummaryComponent
          }
        ]
      },
      {
        path: 'cluster',
        children: [
          { path: '', redirectTo: 'management', pathMatch: 'full' },
          {
            path: 'management',
            component: ManagementComponent
          },
          {
            path: 'management/:id/:name',
            component: ClusterdetailComponent
          },
          {
            path: 'operations',
            component: OperationComponent
          }
        ]
      },
      {
        path: 'appmgmt',
        canActivateChild: [FeatureGuard],
        data: {
          permissionParam: {
            resource: 'appMgmt'
          }
        },
        children: [
          { path: '', redirectTo: 'components', pathMatch: 'full' },
          {
            path: 'components',
            component: AppcomponentsComponent
          },
          {
            path: 'components/:id/:name',
            component: AppcomponentDetailsComponent
          },
          {
            path: 'catalogs',
            component: CatalogsComponent
          },
          {
            path: 'catalogs/details/:name',
            component: CatalogDetailsComponent
          },
          {
            path: 'routes',
            component: ApproutesComponent,
            children: [
              {
                path: 'ingress',
                component: IngressComponent,
                data: {
                  permissionParam: {
                    resource: 'appMgmt'
                  }
                },
              },
              {
                path: 'route',
                component: RouteComponent,
                data: {
                  permissionParam: {
                    resource: 'appMgmt'
                  }
                },
              }
            ]
          },
          {
            path: 'volumes',
            component: AppvolumesComponent,
            children: [
              {
                path: 'configmap',
                component: ConfigmapComponent,
                data: {
                  permissionParam: {
                    resource: 'appMgmt'
                  }
                },
              },
              {
                path: 'secret',
                component: SecretsComponent,
                data: {
                  permissionParam: {
                    resource: 'appMgmt'
                  }
                },
              },
              {
                path: 'pvc',
                component: PvcComponent,
                data: {
                  permissionParam: {
                    resource: 'appMgmt'
                  }
                },
              },
              {
                path: 'storageclass',
                component: StorageclassComponent,
                data: {
                  permissionParam: {
                    resource: 'appMgmt'
                  }
                },
              }
            ]
          }
        ]
      },
      {
        path: 'appbuild',
        canActivateChild: [FeatureGuard],
        data: {
          permissionParam: {
            resource: 'appBuild'
          }
        },
        children: [
          { path: '', redirectTo: 'binary-image', pathMatch: 'full' },
          {
            path: 'binary-image',
            component: BinaryImageComponent
          },
          {
            path: 'binary-image/:id/:name',
            component: BinaryImageDetailsComponent
          },
          {
            path: 'dockerfile-image',
            component: DockerfileImageComponent
          },
          {
            path: 'dockerfile-image/:id/:name',
            component: DockerfileImageDetailsComponent
          },
          {
            path: 'source-image',
            component: SourceImageComponent
          },
          {
            path: 'source-image/:id/:name',
            component: SourceImageDetailsComponent
          }
        ]
      },
      {
        path: 'usermgmt',
        children: [
          { path: '', redirectTo: 'org', pathMatch: 'full' },
          {
            path: 'org',
            component: OrganizationComponent
          },
          {
            path: 'org/:id',
            component: OrgDetailsComponent
          },
          {
            path: 'user',
            canActivate: [OrgRolePermissionGuard],
            data: {
              permissionParam: {
                resource: handleObject.USER
              }
            },
            component: UsersComponent
          },
          {
            path: 'auditlogs',
            canActivate: [OrgRolePermissionGuard],
            data: {
              permissionParam: {
                resource: handleObject.AUDITLOG
              }
            },
            component: AuditlogComponent
          }
        ]
      },
      {
        path: 'system',
        canActivateChild: [OrgRolePermissionGuard],
        data: {
          permissionParam: {
            resource: handleObject.CLUSTER_PROVIDER
          }
        },
        children: [
          { path: '', redirectTo: 'providers', pathMatch: 'full' },
          {
            path: 'providers',
            component: ProviderComponent
          },
          {
            path: 'providers/:id/:name',
            component: ProviderdetailComponent
          },
          {
            path: 'lbs',
            component: LoadbalancerComponent
          },
          {
            path: 'repos',
            component: RepositorysComponent
          },
          {
            path: 'standard-packages',
            component: StandardPackageRepositoryComponent
          }
        ]
      }
    ]
  },
  { path: '**', component: NotFoundComponent }
];

@NgModule({
  imports: [
    RouterModule.forRoot(routes, {
      anchorScrolling: 'enabled',
      enableTracing: false,
      scrollOffset: [0, 0],
      scrollPositionRestoration: 'enabled',
      relativeLinkResolution: 'legacy'
    })
  ],
  exports: [RouterModule],
  providers: [AuthGuard]
})
export class AppRoutingModule { }
