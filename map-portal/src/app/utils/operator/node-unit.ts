export function convertCPUUnit (num: number, unit: string): number {
  // change to core
  const CPUUnitMap = new Map();
  CPUUnitMap.set('', 1)
  CPUUnitMap.set('n', 1e-9)
  CPUUnitMap.set('u', 1e-6)
  CPUUnitMap.set('m', 1e-3)
  CPUUnitMap.set('k', 1e3)
  CPUUnitMap.set('M', 1e6)
  CPUUnitMap.set('G', 1e9)
  CPUUnitMap.set('T', 1e12)
  CPUUnitMap.set('P', 1e15)
  CPUUnitMap.set('E', 1e18)
  const changedNum = num * CPUUnitMap.get(unit);
  return Math.round((changedNum + Number.EPSILON) * 10000) / 10000;
}
// used in cluster compute node
export function convertMemUsageUnit (num: number, unit: string): number {
  // change to Gi
  const memUnitMap = new Map();
  memUnitMap.set('', 0)
  memUnitMap.set('Ki', 1)
  memUnitMap.set('Mi', 2)
  memUnitMap.set('Gi', 3)
  memUnitMap.set('Ti', 4)
  memUnitMap.set('Pi', 5)
  memUnitMap.set('Ei', 6)
  const changeNum = num / (Math.pow(1024, 3 - memUnitMap.get(unit)));
  // don't use toFixed()
  return Math.round((changeNum + Number.EPSILON) * 10000) / 10000;
}

export function convertMemCapUnit (num: number, unit: string): number {
  // change to Gi
  const memUnitMap = new Map();
  memUnitMap.set('', 0)
  memUnitMap.set('Ki', 1)
  memUnitMap.set('Mi', 2)
  memUnitMap.set('Gi', 3)
  memUnitMap.set('Ti', 4)
  memUnitMap.set('Pi', 5)
  memUnitMap.set('Ei', 6)
  const changeNum = num / (Math.pow(1024, 3 - memUnitMap.get(unit)));
  return Number(Math.round(changeNum))
}
