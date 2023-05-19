import Vue from "vue"

Vue.filter("date", function (dateTime, fmt) {
  var dateTimes = new Date(dateTime)
  var o = {
    "M+": dateTimes.getMonth() + 1,
    "d+": dateTimes.getDate(),
    "h+": dateTimes.getHours(),
    "m+": dateTimes.getMinutes(),
    "s+": dateTimes.getSeconds(),
    "q+": Math.floor((dateTimes.getMonth() + 3) / 3),
    "S": dateTimes.getMilliseconds()
  }

  if (/(y+)/.test(fmt)) {
    fmt = fmt.replace(RegExp.$1, (dateTimes.getFullYear() + "").substr(4 - RegExp.$1.length))
  }
  for (var k in o) {
    if (new RegExp(`(${k})`).test(fmt)) {
      fmt = fmt.replace(RegExp.$1, (RegExp.$1.length === 1) ? (o[k]) : (("00" + o[k]).substr(("" + o[k]).length)))
    }
  }
  return fmt
})


