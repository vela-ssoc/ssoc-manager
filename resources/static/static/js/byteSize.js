// 格式转换
function company(bytes) {
  if (bytes === 0) return "0 B"
  let k = 1024
  let sizes = ["B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"]
  let i = Math.floor(Math.log(bytes) / Math.log(k))
  if (i > 0) {
    return (bytes / Math.pow(k, i)).toPrecision(3) + " " + sizes[i]
  } else {
    return bytes.toFixed(2) + " B"
  }
}

export default {company}
