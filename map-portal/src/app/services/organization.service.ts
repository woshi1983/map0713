import { Injectable } from '@angular/core';
import { Apollo, gql } from 'apollo-angular';
import { Observable } from 'rxjs';
import { CreateOrganization, DeleteOrganization, Organization, UpdateOrganization } from '../classes';
import { AuthService } from './auth.service';
import { Router } from '@angular/router';
import { CookieService } from 'ngx-cookie';
import { CURRENT_ORG_KEY, cookieOption } from '../utils/static-entries/shared.const';

@Injectable({
  providedIn: 'root'
})
export class OrganizationService {
  private selectedOrg: Organization;

  setSelectOrg (orgId: string, orgName: string): void {
    this.selectedOrg = { id: orgId, name: orgName }
    this.cookie.put(CURRENT_ORG_KEY, JSON.stringify(this.selectedOrg), cookieOption);
  }

  get getSelectedOrgId (): string {
    return this.selectedOrg.id;
  }

  get getSelectedOrgName (): string {
    return this.selectedOrg ? this.selectedOrg.name : '';
  }

  constructor (
    private authService: AuthService,
    private apollo: Apollo,
    private router: Router,
    private cookie: CookieService
  ) {
    // get org when login
    if (this.authService.isLoggedIn) {
      const cookieOrg = this.cookie.get(CURRENT_ORG_KEY);
      // login already
      if (cookieOrg) {
        this.selectedOrg = JSON.parse(cookieOrg);
        return;
      }
      // first login
      const userInfo = this.authService.getUserInfo();
      if (!!userInfo.def_org?.id) {
        this.setSelectOrg(userInfo.def_org.id, userInfo.def_org.name)
      } else {
        // user not in any org error and not in logout page
        if (this.router.url !== '/portal/logout') {
          this.router.navigate(['/portal/no-org'], { replaceUrl: true });
        }
      }
    }
  }

  getAllOrgs (num: number, limit: number, filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Organizations(page: {number: ${num}, limit: ${limit}}, filters: $filters) {
            totalCount
            edge {
              nodes {
                id
                name
                description
                quota {
                  total {
                    cpus
                    memoryMi
                    storageGi
                  }
                }
                createdAt
                lastModified
              }
            }
          }
        }`,
      variables: {
        filters: filters
      },
      errorPolicy: 'all'
    });
  }

  getAllOrgNames (): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          Organizations(page: {limit: -1}) {
            edge {
              nodes {
                id
                name
              }
            }
          }
        }`,
      errorPolicy: 'all'
    });
  }

  getAllOrgCounts (): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          Organizations {
            totalCount
          }
        }`,
      errorPolicy: 'all'
    });
  }

  getOrgById (id: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
      query {
        Organization(id: "${id}") {
          name
          description
          quota {
            used {
              cpus
              memoryMi
              storageGi
            }
            total {
              cpus
              memoryMi
              storageGi
            }
          }
        }
      }`,
      errorPolicy: 'all'
    });
  }

  getOrgQuotaById (id: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
      query {
        Organization(id: "${id}") {
          quota {
            used {
              cpus
              memoryMi
              storageGi
            }
            total {
              cpus
              memoryMi
              storageGi
            }
          }
        }
      }`,
      errorPolicy: 'all'
    });
  }

  createOrg (input: CreateOrganization): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateOrganization!){
          CreateOrganization(input: $input){
            id
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  updateOrg (input: UpdateOrganization): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateOrganization!){
          UpdateOrganization(input: $input){
            id
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  deleteOrg (id: string): Observable<any> {
    const input: DeleteOrganization = {
      id: id
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteOrganization!){
          DeleteOrganization(input: $input){
            id
            name
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }
}
