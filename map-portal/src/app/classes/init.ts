import {Time} from "@angular/common";

export class X509Certificate {
  raw: string;
  sig: string;
  sigAlg: string;
  ver: number;
  sn: string;
  iss: string;
  sub: string;
  notBefore: Time;
  notAfter: Time;
  keyUsage: string;
  extensions: X509CertificateExtension[];
  basicConstraintsValid: boolean;
  isCA: boolean;
  dnsNames: string[];
}

export class X509CertificateExtension {
  id: string;
  value: string;
}

