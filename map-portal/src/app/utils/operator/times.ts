export function timeTransit (time: Date): string {
  const year = time.getFullYear() + ''; // 年
  const tempmouth = time.getMonth() + 1; // 月
  const mouth = tempmouth < 10 ? '0' + tempmouth : tempmouth + '';
  const tempdate = time.getDate(); // 日
  const date = tempdate < 10 ? '0' + tempdate : tempdate + '';
  const temphours = time.getHours(); // 时
  const hours = temphours < 10 ? '0' + temphours : temphours + '';
  const tempminutes = time.getMinutes(); // 分
  const minutes = tempminutes < 10 ? '0' + tempminutes : tempminutes + '';
  const tempseconds = time.getSeconds(); // 秒
  const seconds = tempseconds < 10 ? '0' + tempseconds : tempseconds + '';

  return year + mouth + date + hours + minutes + seconds;
}

// used for audit log filter
export function timeAgo (time: number, unit: string): string[] {
  const unitMap = {
    minutes: 60 * 1000,
    hours: 60 * 60 * 1000,
    days: 24 * 60 * 60 * 1000,
    weeks: 7 * 24 * 60 * 60 * 1000,
    months: 30 * 24 * 60 * 60 * 1000,
    years: 365 * 24 * 60 * 60 * 1000
  }
  const dateNow = Date.now();
  const untilTime = new Date(dateNow).toISOString();
  const startTime = new Date(dateNow - time * unitMap[unit]).toISOString();
  return [startTime, untilTime]
}
