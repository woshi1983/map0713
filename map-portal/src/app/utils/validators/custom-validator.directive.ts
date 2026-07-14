import { Directive } from '@angular/core';
import { AbstractControl, ValidatorFn } from '@angular/forms';
import subnetOverlap from 'subnet-overlap';
import { AddPodSpecContainer } from '../../classes';
import YAML from 'yaml';
import { isValidCron } from 'cron-validator';
import {Namespacediffset} from "../operator";

@Directive({
  selector: '[appCustomValidator]'
})
export class CustomValidatorDirective {
  // common object name pattern, like a01-b02.exampe-01.edf01.exsf. same with k8s resource pattern
  // used for k8s resource: ingress, backup, backuplocation, restore
  public static commonNamePattern = '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*';

  // restrictive object name Pattern, like test01-xsd-02-esd, same with k8s resource pattern
  // used for cluster-workloads, user, role, route, cluster, registry, namespace, pvc, cm, secret
  public static restricNamePattern = '[a-z]([-a-z0-9]*[a-z0-9])?';

  // endpoint pattern, like http://url:5432 / https://AWSdfgc.example-es.com
  public static endpointPattern = 'https?://[A-Za-z0-9.-]+(:\\d{1,5})?';

  // broad endpoint pattern, like http://url:5432/#/pipeline?xxxx
  public static broadEndpointPattern = 'https?:\\/\\/(www\\.)?[-a-zA-Z0-9@:%._\\+~#=]{1,256}\\.[a-zA-Z0-9()]{1,6}\\b([-a-zA-Z0-9()@:%_\\+.~#?&//=]*)';

  // domain pattern, like AWSdfgc.example-es.com
  public static hostsPattern = '^(?=^.{3,255}$)[a-zA-Z0-9*][-a-zA-Z0-9]{0,62}(\\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+$';

  // package pattern, like http://xxx/he.war
  // used for binary build package or dockerfile build
  public static packagePattern = '^http(s)?://.*';

  // git URL, used for build
  public static gitUrlPattern = '^http(s)?://([^\\/]+?\\/){2}.*?.git$';

  // CIDR pattern
  public static CIDRPattern = '\\d{0,3}\\.\\d{0,3}\\.\\d{0,3}\\.\\d{0,3}/\\d{1,2}';

  // IP address pattern
  public static IPPattern = '\\d{0,3}\\.\\d{0,3}\\.\\d{0,3}\\.\\d{0,3}';

  // check user password
  // 密码长度至少为8，且需包含至少一个大写字母，一个小写字母和一个数字。
  public static passwordPattern = '^.*(?=.{8,})(?=.*\\d)(?=.*[A-Z])(?=.*[a-z]).*$';

  // check label key
  // 不显示这些label
  public static disAllowLabels = ['map.tanzu.vmware.com/application', 'map.tanzu.vmware.com/component', 'map.tanzu.vmware.com/federation', 'map.tanzu.vmware.com/instance', 'map.tanzu.vmware.com/revision', 'map.tanzu.vmware.com/managed', 'map.tanzu.vmware.com/internal']

  // check node label key
  // 不可编辑节点这些label
  public static disAllowNodeLabels = ['beta.kubernetes.io/arch', 'beta.kubernetes.io/os', 'kubernetes.io/arch', 'kubernetes.io/hostname', 'kubernetes.io/os', 'node-role.kubernetes.io/control-plane', 'node-role.kubernetes.io/master', 'run.tanzu.vmware.com/kubernetesDistributionVersion']

  // check node taint key
  // 不可删除节点这些key
  public static disAllowNodeTaint = ['node.kubernetes.io/unschedulable'];
  // check annotation key
  // 不显示这些annotation
  public static disAllowAnnotations = ['ako.vmware.com/controller-cluster-uuid', 'ako.vmware.com/host-fqdn-vs-uuid-map']

  public static systemNamespace = Namespacediffset.excludedNamespaceNmaeList;
}

// check network mask is below /24 for pod network CIDR
export function podCIDRMinNetmaskValidator (minMask: number): ValidatorFn {
  return (control: AbstractControl): {[key: string]: boolean} | null => {
    if (control.value !== undefined && control.value) {
      const podCIDR = control.value.toString();
      const netmask = podCIDR.split('/', 2)[1];
      if (Number(netmask) > minMask) {
        return { podNetmask: true };
      }
    }
    return null;
  };
}

// check text match CA format
export function CAFormatValidator (): ValidatorFn {
  return (control: AbstractControl): {[key: string]: boolean} | null => {
    if (control.value !== undefined && control.value) {
      const caString = control.value.toString().replace(/\r\n|\r|\n/g);
      if (!caString.match('^-----BEGIN CERTIFICATE-----.*-----END CERTIFICATE-----$')) {
        return { CAFormat: true };
      }
    }
    return null;
  };
}

// check text match SSH/TLS private key format
export function KEYFormatValidator (): ValidatorFn {
  return (control: AbstractControl): {[key: string]: boolean} | null => {
    if (control.value !== undefined && control.value) {
      const keyString = control.value.toString().replace(/\r\n|\r|\n/g);
      if (!keyString.match('^-----BEGIN RSA PRIVATE KEY-----.*-----END RSA PRIVATE KEY-----$')) {
        return { KEYFormat: true };
      }
    }
    return null;
  };
}

// check network overlap for pod network CIDR and service network CIDR
export function networkOverlapValidator (checkdNetwork: string, supervisorNetwork: string): boolean {
  if (checkdNetwork && supervisorNetwork) {
    return subnetOverlap([checkdNetwork], supervisorNetwork);
  }
}

// check container port of deployment overlap
export function portOverlapValidator (allContainer: AddPodSpecContainer[], port: number, protocol: string): boolean {
  let portOverlap = false;
  if (allContainer.length > 0) {
    for (const container of allContainer) {
      if(container.ports){
        const portList = container.ports;
        for (const each of portList){
          if (each.protocol === protocol && each.containerPort === port){
            portOverlap = true;
            break;
          }
        }
        if(portOverlap){
          break;
        }
      }
    }
  }
  return portOverlap
}

export function nameValidator (invaildList: string[]): ValidatorFn {
  const staticInvalidName = ['null', 'Null', 'NUll', 'NULl', 'NULL'];
  return (control: AbstractControl): {[key: string]: boolean} | null => {
    if (control.value !== undefined && control.value) {
      const name = control.value.toString();
      if (invaildList.findIndex(x => x === name) !== -1 || staticInvalidName.findIndex(x => x === name) !== -1) {
        return { name: true };
      }
    }
    return null;
  };
}

// check JSON format
export function checkJsonFormat (context: string): [boolean, string] {
  try {
    JSON.parse(context);
  } catch (e) {
    return [false, e];
  }
  return [true, ''];
}

export function systemNamespaceValidator(): ValidatorFn {
  return (control: AbstractControl): {[key: string]: boolean} | null => {
    if (control.value !== undefined && control.value) {
      const name = control.value.toString();
      if (CustomValidatorDirective.systemNamespace.findIndex(x => x === name) !== -1) {
        return { 'namespace': true};
      }
    }
    return null;
  };
}

// check one more of YAML format
export function checkYamlFormat (context: string): [boolean, string] {
  try {
    const docs = YAML.parseAllDocuments(context);
    for (const doc of docs) {
      if (doc.errors) {
        // noinspection LoopStatementThatDoesntLoopJS
        for(const e of doc.errors){
          return [false, e.message]
        }
      }
    }
  } catch (e) {
    return [false, e];
  }
  return [true, ''];
}

// check one YAML format
export function checkOneYamlFormat (context: string): [boolean, string] {
  try {
    YAML.parse(context);
  } catch (e) {
    return [false, e];
  }
  return [true, ''];
}

// check cron vaild
export function cronValidator (): ValidatorFn {
  return (control: AbstractControl): {[key: string]: boolean} | null => {
    if (control.value !== undefined && control.value) {
      if (control.value.trim().split(/\s+/g).length < 5 || !isValidCron(control.value, { alias: true, allowBlankDay: true })) {
        return { cron: true };
      }
    }
    return null;
  };
}
