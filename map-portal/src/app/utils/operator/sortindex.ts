// used for summary org resource sorted
export function getSortIndexList (list: number[]): any[] {
  const listWithIndex = [];
  for (const i in list) {
    listWithIndex.push([list[i], i]);
  }
  listWithIndex.sort(function (left, right) {
    return left[0] < right[0] ? 1 : -1;
  });
  const indexes = [];
  const sorted = [];
  for (const j in listWithIndex) {
    sorted.push(listWithIndex[j][0]);
    indexes.push(Number(listWithIndex[j][1]));
  }
  return [indexes, sorted];
}
