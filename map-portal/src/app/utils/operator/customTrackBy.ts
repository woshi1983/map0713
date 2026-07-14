// fix "ng-reflect-model shows correct value but not reflecting in input", see https://stackoverflow.com/questions/40663198/ng-reflect-model-shows-correct-value-but-not-reflecting-in-input
export function customTrackBy (index: number, obj: any): any {
  return index;
}
