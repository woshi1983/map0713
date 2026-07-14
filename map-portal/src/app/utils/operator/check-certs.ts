// cert expired in 30 days
export class certCheckResult {
  name: string;
  notAfter: string;
  isCA: boolean;
}

export function checkCerts(data: any[]): certCheckResult[] {
  const dateNow = Date.now();
  const expireDate = 30 * 24 * 60 * 60 * 1000; //30 days in ms
  let result: certCheckResult[] = [];
  if(!!data){
    for (let object of data){
      if(!!object.tlsCertChain){
        for (let cert of object.tlsCertChain){
          const notAfterDate = new Date(cert.notAfter).valueOf();
          if( (notAfterDate - dateNow) <= expireDate ){
            result.push({
              name: object.name,
              notAfter: cert.notAfter,
              isCA: cert.isCA
            })
          }
        }
      }
    }
  }
  return result;
}
