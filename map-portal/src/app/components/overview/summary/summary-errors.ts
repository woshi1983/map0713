export interface SummaryQueryError {
  message: string;
  extensions?: { cluster?: string; reason?: string };
}

export function splitSummaryErrors(errors: SummaryQueryError[]): {
  clusters: string[];
  otherErrors: SummaryQueryError[];
} {
  const clusters = new Set<string>();
  const otherErrors: SummaryQueryError[] = [];
  for (const error of errors) {
    if (error.extensions?.cluster) {
      clusters.add(error.extensions.cluster);
    } else {
      otherErrors.push(error);
    }
  }
  return { clusters: Array.from(clusters).sort(), otherErrors };
}
