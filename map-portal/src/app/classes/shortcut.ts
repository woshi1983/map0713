import { Organization } from './organization';

export class Shortcut {
  id: string;
  name: string;
  description: string;
  endpoint: string;
  organization: Organization;
  createdAt: string;
  lastModified: string;
}

export class ShortcutNodes {
  nodes: Shortcut[];
}

export class Shortcuts {
  // page: PageInfo!
  // filters: [FilterInfo!]
  totalCount: number;
  edge: ShortcutNodes;
}

export class CreateShortcut {
  name: string;
  description?: string;
  endpoint: string;
  organization: string;
}

export class UpdateShortcut {
  id: string;
  description?: string;
  endpoint?: string;
}

export class DeleteShortcut {
  id: string;
}
