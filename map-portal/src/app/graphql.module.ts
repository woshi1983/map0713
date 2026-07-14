import { NgModule } from '@angular/core';
import { ApolloModule, APOLLO_OPTIONS } from 'apollo-angular'
import { DefaultOptions, InMemoryCache, split, ApolloClientOptions } from '@apollo/client/core';
import { HttpLink } from 'apollo-angular/http';
import { WebSocketLink } from '@apollo/client/link/ws';
import { setContext } from '@apollo/client/link/context';
import { CookieService } from 'ngx-cookie';
import { Router } from '@angular/router';
import { getMainDefinition } from '@apollo/client/utilities';
import { TOKEN_COOKIE_KEY } from './utils/static-entries/shared.const';

const tokenCookieKey = TOKEN_COOKIE_KEY;
const host = window.location.host;
const protocol = window.location.protocol;
const defaultOptions: DefaultOptions = {
  query: {
    errorPolicy: 'all',
    fetchPolicy: 'no-cache'
  },
  watchQuery: {
    errorPolicy: 'all',
    fetchPolicy: 'no-cache'
  }
};

export function createApollo (httpLink: HttpLink, cookie: CookieService, router: Router): ApolloClientOptions<any> {
  const http = httpLink.create({
    uri: '/api/graphql'
  });

  const ws = new WebSocketLink({
    uri: `${protocol === 'https:' ? 'wss:' : 'ws:'}//${host}/stream/logs`,
    options: {
      reconnect: false,
      connectionParams: async () => {
        const rawToken = cookie.get(tokenCookieKey) ? cookie.get(tokenCookieKey).trim() : null;
        if (!rawToken) {
          // do noting
        }
        const tokenList = rawToken.split('.');
        if (tokenList.length < 2) {
          console.error('Token format is error.');
          router.navigate(['/portal/logout']);
        }
        const token = atob(tokenList[0]);
        return {
          Authorization: token ? `Bearer ${token}` : ''
        };
      }
    }
  });

  const authLink = setContext(() => {
    // get the authentication token from cookie if it exists
    const rawToken = cookie.get(tokenCookieKey) ? cookie.get(tokenCookieKey).trim() : null;
    if (!rawToken) {
      /*
          const href = window.location.protocol + '//' +  window.location.host + '/portal';
          window.location.href = window.location.origin + '/api/auth/.well-known/login?redirectURI=' + encodeURI(href);
           */
    }
    const tokenList = rawToken.split('.');
    if (tokenList.length < 2) {
      console.error('Token format is error.');
      router.navigate(['/portal/logout']);
    }
    const token = atob(tokenList[0]);
    // return the headers to the context so httpLink can read them
    return {
      headers: {
        authorization: token ? `Bearer ${token}` : ''
      }
    };
  });

  const link = split(
    ({ query }) => {
      const data = getMainDefinition(query);
      return (
        data.kind === 'OperationDefinition' && data.operation === 'subscription'
      );
    },
    ws,
    authLink.concat(http)
  );

  return {
    link: link,
    defaultOptions,
    cache: new InMemoryCache()
  };
}

@NgModule({
  imports: [
    ApolloModule
  ],
  providers: [
    {
      provide: APOLLO_OPTIONS,
      useFactory: createApollo,
      deps: [HttpLink, CookieService, Router]
    }
  ]
})

export class GraphQLModule {}
