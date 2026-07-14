const fs = require('fs');
const path = require('path');

const files = [
  'node_modules/@clr/angular/esm2020/modal/modal-body.mjs',
  'node_modules/@clr/angular/fesm2020/clr-angular.mjs',
  'node_modules/@clr/angular/fesm2015/clr-angular.mjs',
];

const before = `    addOrRemoveTabIndex() {
        const modalBody = this.host.nativeElement.parentElement;
        if (modalBody.clientHeight < modalBody.scrollHeight) {`;

const after = `    addOrRemoveTabIndex() {
        const modalBody = this.host.nativeElement.parentElement;
        if (!modalBody) {
            return;
        }
        if (modalBody.clientHeight < modalBody.scrollHeight) {`;

let patched = 0;

for (const relativeFile of files) {
  const file = path.join(__dirname, '..', relativeFile);
  if (!fs.existsSync(file)) {
    continue;
  }

  const source = fs.readFileSync(file, 'utf8');
  if (source.includes(after)) {
    continue;
  }
  if (!source.includes(before)) {
    throw new Error(`Unable to patch ${relativeFile}: target block not found`);
  }

  fs.writeFileSync(file, source.replace(before, after));
  patched += 1;
}

if (patched > 0) {
  console.log(`Patched @clr/angular modal body null guard in ${patched} file(s).`);
}
