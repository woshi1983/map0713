/**
 * deepCopy dives into given `obj`, copy each of its properties, and returns a duplicated new instance.
 *
 * @export
 * @param {*} obj The given source to deep copy from.
 * @returns The duplicated instance.
 */
// eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
export function deepCopy (obj: any): any {
  let buf;

  // Handle the 3 simple types, and null or undefined
  if (obj == null || typeof obj !== 'object') {
    return obj;
  }

  // Handle Date
  if (obj instanceof Date) {
    buf = new Date();
    buf.setTime(obj.getTime());
    return buf;
  }

  // Handle Array
  if (obj instanceof Array) {
    buf = [];
    for (let i = 0, len = obj.length; i < len; i++) {
      buf[i] = deepCopy(obj[i]);
    }
    return buf;
  }

  // Handle Object
  if (obj instanceof Object) {
    buf = {};
    for (const attr in obj) {
      // eslint-disable-next-line no-prototype-builtins
      if (obj.hasOwnProperty(attr)) {
        buf[attr] = deepCopy(obj[attr]);
      }
    }
    return buf;
  }

  // Panic if unknown type was found.
  throw new Error(`Unable to copy obj due to unsupported type: ${typeof obj}`);
}
