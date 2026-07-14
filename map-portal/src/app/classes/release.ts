import { CustomComponents } from './component';
import { Application } from './application';
import { Organization } from './organization';

export class Release {
  id: string;
  name: string;
  num: number;
  manifest: any[];
  component: CustomComponents;
  application: Application;
  organization: Organization;
  createdAt: string;
  lastModified: string;
}

export class Releases {
  totalCount: number;
  edge: {
    nodes: Release[];
  }
}
