const elasticsearch = require("elasticsearch")
const client = new elasticsearch.Client({hosts: window.location.host + "/api/ribana"})


// 查询索引
function indices() {
  const bodys = client.cat.indices({
    headers: {"Authorization": localStorage.getItem("token"), "accept": "application/json;charset=utf-8"},
    h: ['index', 'health', 'status', 'uuid', 'pri', 'rep', 'docs.count', 'store.size'],
  })
  return bodys
}

// 删除索引
function indicesDelete(data) {
  const bodys = client.indices.delete({
    headers: {"Authorization": localStorage.getItem("token"), "accept": "application/json;charset=utf-8"},
    index: data
  })
  return bodys
}

// 查询搜索
function indicesSearch(data) {
  const bodys = client.search({
    headers: {"Authorization": localStorage.getItem("token"), "accept": "application/json;charset=utf-8"},
    h: ['index', 'health', 'status', 'uuid', 'pri', 'rep', 'docs.count', 'store.size'],
    body: {
      "query": {
        "bool": {
          "must": [{
            "match": {
              "_index": data
            }
          }]
        }
      }
    }
  })
  return bodys
}


// 查询nodes
function nodes() {
  const bodys = client.cat.nodes({
    headers: {"Authorization": localStorage.getItem("token"), "accept": "application/json;charset=utf-8"},
    h: ['ip', 'name', 'heap.percent', 'heap.current', 'heap.max', 'ram.percent', 'ram.current', 'ram.max', 'node.role', 'master', 'cpu', 'load_1m', 'load_5m', 'load_15m', 'disk.used_percent', 'disk.used', 'disk.total']
  })
  return bodys
}

// 根据索引查询列表
function searchIndices(mustNot, filterArray, searchData, current, page) {
  const body = client.search({
    headers: {"Authorization": localStorage.getItem("token"), "accept": "application/json;charset=utf-8"},
    body: {
      from: current,
      size: page,
      track_total_hits: true,
      "query": {
        "bool": {
          "must": [{
            "match": {
              "_index": searchData
            }
          }],
          "filter": filterArray,
          "must_not": mustNot
        }
      },
      sort: [
        {
          "@timestamp": {
            "order": "desc"
          }
        }
      ]
    }
  })
  return body
}


class Result {
  constructor(round) {
    this.round = round
    this.index = {}
    this.fields = [];
  }

  /**
   * 添加一条属性
   *
   * @param k 即 field 的名字
   * @param v 对应 field 的值
   */
  add(k, v) {
    let field = this.index[k]
    if (!field) {
      field = new Field(k, v)
      this.fields.push(field)
      this.index[k] = field
    } else {
      field.add(v)
    }
    this.sorted = false
  }

  sort() {
    if (this.sorted) return
    this.fields.forEach(field => field.sort())
    this.fields.sort((a, b) => a.name.localeCompare(b.name))
    this.sorted = true
  }

  topN(n = 5) {
    this.sort()
    let fields = []
    this.fields.forEach(f => {
      fields.push({name: f.name, type: f.type, values: f.topN(n)})
    })

    return {round: this.round, fields: fields}
  }
}

class Field {
  constructor(k, v) {
    let value = new Value(v)

    this.index = {}
    this.name = k
    this.type = typeof v
    this.values = [value]
    this.index[v] = value
  }

  add(v) {
    let value = this.index[v]
    if (this.type !== typeof v) this.type = "unknown"
    if (!value) {
      let value = new Value(v)
      this.values.push(value)
      this.index[v] = value
    } else {
      value.incr()
    }
    this.sorted = false
  }

  sort() {
    if (this.sorted) return
    this.values.sort((a, b) => a.compare(b))
    this.sorted = true
  }

  topN(n) {
    this.sort()
    return this.values.slice(0, n)
  }
}

class Value {
  constructor(v) {
    this.value = v
    this.count = 1
  }

  incr() {
    this.count++
  }

  compare(val) {
    // 先比较出现频次, 频次越高越靠前
    let tc = this.count
    let vc = val.count

    if (tc > vc) return -1
    if (tc < vc) return 1

    let tv = this.value
    let vv = val.value

    // 如果 count 的次数相同, 再根据 value 排序
    if (typeof tv === "string" && typeof vv === "string") return tv.localeCompare(vv)
    if (typeof tv === "number" && typeof vv === "number") {
      if (tv > vv) return -1
      if (tv < vv) return 1
    }

    return 0
  }
}

/**
 * 生成类似 kibana 左侧频次统计
 *
 * @param hits es 数据的 hits
 * @param limit 最大循环次数
 * @param top 结果排行前几名
 * @returns {{round, fields: *[]}} 结果, round-实际循环的次数, fields-内容
 */
function extractData(hits, limit = 500, top = 5) {
  let round = hits.length > limit ? limit : hits.length
  let ret = new Result(round)
  for (let i = 0; i < round; i++) {
    let hit = hits[i]
    unfold(hit, ret)
  }
  return ret.topN(top)
}

/**
 * 遍历对象的 field 和 value 并统计
 *
 * @param obj 对象
 * @param ret 结果集
 * @param prefix 前缀, 注层递归需要
 */
function unfold(obj, ret, prefix = "") {
  for (let field in obj) {
    let value = obj[field]
    let key = ""
    if (prefix === "") {
      key = field
    } else {
      key = prefix + "." + field
    }

    if (Array.isArray(value)) {
      for (let i = 0; i < value.length; i++) {
        let inner = value[i]
        if (typeof inner === "object") {
          unfold(inner, ret, key)
        } else {
          ret.add(key, inner)
        }
      }
    } else if (typeof value === "object") {
      // 如果是 _source 下的属性, 不用显示 _source 前缀, 目的是和 kibana 侧边栏效果一致
      if (key === "_source") key = ""
      unfold(value, ret, key)
    } else {
      ret.add(key, value)
    }
  }
}

// let data = JSON.parse('{"took":0,"timed_out":false,"_shards":{"total":1,"successful":1,"skipped":0,"failed":0},"hits":{"total":{"value":10,"relation":"eq"},"max_score":1,"hits":[{"_index":".tasks","_type":"task","_id":"OyTdlcmrSCWwR1AQPHeFgw:2277","_score":1,"_source":{"completed":true,"task":{"node":"OyTdlcmrSCWwR1AQPHeFgw","id":2277,"type":"transport","action":"indices:data/write/update/byquery","status":{"total":15,"updated":15,"created":0,"deleted":0,"batches":1,"version_conflicts":0,"noops":0,"retries":{"bulk":0,"search":0},"throttled_millis":0,"requests_per_second":-1,"throttled_until_millis":0},"description":"update-by-query [.kibana_task_manager_7.15.1_001]","start_time_in_millis":1636000363927,"running_time_in_nanos":55797467,"cancellable":true,"headers":{}},"response":{"took":55,"timed_out":false,"total":15,"updated":15,"created":0,"deleted":0,"batches":1,"version_conflicts":0,"noops":0,"retries":{"bulk":0,"search":0},"throttled":"0s","throttled_millis":0,"requests_per_second":-1,"throttled_until":"0s","throttled_until_millis":0,"failures":[]}}},{"_index":".tasks","_type":"task","_id":"OyTdlcmrSCWwR1AQPHeFgw:2263","_score":1,"_source":{"completed":true,"task":{"node":"OyTdlcmrSCWwR1AQPHeFgw","id":2263,"type":"transport","action":"indices:data/write/update/byquery","status":{"total":11,"updated":11,"created":0,"deleted":0,"batches":1,"version_conflicts":0,"noops":0,"retries":{"bulk":0,"search":0},"throttled_millis":0,"requests_per_second":-1,"throttled_until_millis":0},"description":"update-by-query [.kibana_7.15.1_001]","start_time_in_millis":1636000363771,"running_time_in_nanos":188089367,"cancellable":true,"headers":{}},"response":{"took":183,"timed_out":false,"total":11,"updated":11,"created":0,"deleted":0,"batches":1,"version_conflicts":0,"noops":0,"retries":{"bulk":0,"search":0},"throttled":"0s","throttled_millis":0,"requests_per_second":-1,"throttled_until":"0s","throttled_until_millis":0,"failures":[]}}},{"_index":".tasks","_type":"task","_id":"OyTdlcmrSCWwR1AQPHeFgw:31225","_score":1,"_source":{"completed":true,"task":{"node":"OyTdlcmrSCWwR1AQPHeFgw","id":31225,"type":"transport","action":"indices:data/write/update/byquery","status":{"total":79,"updated":79,"created":0,"deleted":0,"batches":1,"version_conflicts":0,"noops":0,"retries":{"bulk":0,"search":0},"throttled_millis":0,"requests_per_second":-1,"throttled_until_millis":0},"description":"update-by-query [.kibana_7.15.1_001]","start_time_in_millis":1636002631974,"running_time_in_nanos":203788177,"cancellable":true,"headers":{}},"response":{"took":203,"timed_out":false,"total":79,"updated":79,"created":0,"deleted":0,"batches":1,"version_conflicts":0,"noops":0,"retries":{"bulk":0,"search":0},"throttled":"0s","throttled_millis":0,"requests_per_second":-1,"throttled_until":"0s","throttled_until_millis":0,"failures":[]}}},{"_index":".tasks","_type":"task","_id":"OyTdlcmrSCWwR1AQPHeFgw:31247","_score":1,"_source":{"completed":true,"task":{"node":"OyTdlcmrSCWwR1AQPHeFgw","id":31247,"type":"transport","action":"indices:data/write/update/byquery","status":{"total":15,"updated":15,"created":0,"deleted":0,"batches":1,"version_conflicts":0,"noops":0,"retries":{"bulk":0,"search":0},"throttled_millis":0,"requests_per_second":-1,"throttled_until_millis":0},"description":"update-by-query [.kibana_task_manager_7.15.1_001]","start_time_in_millis":1636002632186,"running_time_in_nanos":47320081,"cancellable":true,"headers":{}},"response":{"took":47,"timed_out":false,"total":15,"updated":15,"created":0,"deleted":0,"batches":1,"version_conflicts":0,"noops":0,"retries":{"bulk":0,"search":0},"throttled":"0s","throttled_millis":0,"requests_per_second":-1,"throttled_until":"0s","throttled_until_millis":0,"failures":[]}}},{"_index":".tasks","_type":"task","_id":"OyTdlcmrSCWwR1AQPHeFgw:551","_score":1,"_source":{"completed":true,"task":{"node":"OyTdlcmrSCWwR1AQPHeFgw","id":551,"type":"transport","action":"indices:data/write/update/byquery","status":{"total":15,"updated":15,"created":0,"deleted":0,"batches":1,"version_conflicts":0,"noops":0,"retries":{"bulk":0,"search":0},"throttled_millis":0,"requests_per_second":-1,"throttled_until_millis":0},"description":"update-by-query [.kibana_task_manager_7.15.1_001]","start_time_in_millis":1636528036051,"running_time_in_nanos":210514575,"cancellable":true,"headers":{}},"response":{"took":203,"timed_out":false,"total":15,"updated":15,"created":0,"deleted":0,"batches":1,"version_conflicts":0,"noops":0,"retries":{"bulk":0,"search":0},"throttled":"0s","throttled_millis":0,"requests_per_second":-1,"throttled_until":"0s","throttled_until_millis":0,"failures":[]}}},{"_index":".tasks","_type":"task","_id":"OyTdlcmrSCWwR1AQPHeFgw:548","_score":1,"_source":{"completed":true,"task":{"node":"OyTdlcmrSCWwR1AQPHeFgw","id":548,"type":"transport","action":"indices:data/write/update/byquery","status":{"total":104,"updated":104,"created":0,"deleted":0,"batches":1,"version_conflicts":0,"noops":0,"retries":{"bulk":0,"search":0},"throttled_millis":0,"requests_per_second":-1,"throttled_until_millis":0},"description":"update-by-query [.kibana_7.15.1_001]","start_time_in_millis":1636528036032,"running_time_in_nanos":448571114,"cancellable":true,"headers":{}},"response":{"took":443,"timed_out":false,"total":104,"updated":104,"created":0,"deleted":0,"batches":1,"version_conflicts":0,"noops":0,"retries":{"bulk":0,"search":0},"throttled":"0s","throttled_millis":0,"requests_per_second":-1,"throttled_until":"0s","throttled_until_millis":0,"failures":[]}}},{"_index":".tasks","_type":"task","_id":"OyTdlcmrSCWwR1AQPHeFgw:35811","_score":1,"_source":{"completed":true,"task":{"node":"OyTdlcmrSCWwR1AQPHeFgw","id":35811,"type":"transport","action":"indices:data/write/update/byquery","status":{"total":16,"updated":16,"created":0,"deleted":0,"batches":1,"version_conflicts":0,"noops":0,"retries":{"bulk":0,"search":0},"throttled_millis":0,"requests_per_second":-1,"throttled_until_millis":0},"description":"update-by-query [.kibana_task_manager_7.15.1_001]","start_time_in_millis":1636974815201,"running_time_in_nanos":176859464,"cancellable":true,"headers":{}},"response":{"took":171,"timed_out":false,"total":16,"updated":16,"created":0,"deleted":0,"batches":1,"version_conflicts":0,"noops":0,"retries":{"bulk":0,"search":0},"throttled":"0s","throttled_millis":0,"requests_per_second":-1,"throttled_until":"0s","throttled_until_millis":0,"failures":[]}}},{"_index":".tasks","_type":"task","_id":"OyTdlcmrSCWwR1AQPHeFgw:35810","_score":1,"_source":{"completed":true,"task":{"node":"OyTdlcmrSCWwR1AQPHeFgw","id":35810,"type":"transport","action":"indices:data/write/update/byquery","status":{"total":134,"updated":134,"created":0,"deleted":0,"batches":1,"version_conflicts":0,"noops":0,"retries":{"bulk":0,"search":0},"throttled_millis":0,"requests_per_second":-1,"throttled_until_millis":0},"description":"update-by-query [.kibana_7.15.1_001]","start_time_in_millis":1636974815201,"running_time_in_nanos":383182869,"cancellable":true,"headers":{}},"response":{"took":380,"timed_out":false,"total":134,"updated":134,"created":0,"deleted":0,"batches":1,"version_conflicts":0,"noops":0,"retries":{"bulk":0,"search":0},"throttled":"0s","throttled_millis":0,"requests_per_second":-1,"throttled_until":"0s","throttled_until_millis":0,"failures":[]}}},{"_index":".tasks","_type":"task","_id":"OyTdlcmrSCWwR1AQPHeFgw:38136","_score":1,"_source":{"completed":true,"task":{"node":"OyTdlcmrSCWwR1AQPHeFgw","id":38136,"type":"transport","action":"indices:data/write/update/byquery","status":{"total":16,"updated":16,"created":0,"deleted":0,"batches":1,"version_conflicts":0,"noops":0,"retries":{"bulk":0,"search":0},"throttled_millis":0,"requests_per_second":-1,"throttled_until_millis":0},"description":"update-by-query [.kibana_task_manager_7.15.1_001]","start_time_in_millis":1636974945031,"running_time_in_nanos":41646865,"cancellable":true,"headers":{}},"response":{"took":41,"timed_out":false,"total":16,"updated":16,"created":0,"deleted":0,"batches":1,"version_conflicts":0,"noops":0,"retries":{"bulk":0,"search":0},"throttled":"0s","throttled_millis":0,"requests_per_second":-1,"throttled_until":"0s","throttled_until_millis":0,"failures":[]}}},{"_index":".tasks","_type":"task","_id":"OyTdlcmrSCWwR1AQPHeFgw:38128","_score":1,"_source":{"completed":true,"task":{"node":"OyTdlcmrSCWwR1AQPHeFgw","id":38128,"type":"transport","action":"indices:data/write/update/byquery","status":{"total":140,"updated":140,"created":0,"deleted":0,"batches":1,"version_conflicts":0,"noops":0,"retries":{"bulk":0,"search":0},"throttled_millis":0,"requests_per_second":-1,"throttled_until_millis":0},"description":"update-by-query [.kibana_7.15.1_001]","start_time_in_millis":1636974944861,"running_time_in_nanos":246988845,"cancellable":true,"headers":{}},"response":{"took":246,"timed_out":false,"total":140,"updated":140,"created":0,"deleted":0,"batches":1,"version_conflicts":0,"noops":0,"retries":{"bulk":0,"search":0},"throttled":"0s","throttled_millis":0,"requests_per_second":-1,"throttled_until":"0s","throttled_until_millis":0,"failures":[]}}}]}}')

function percent(num, total) {
  let cell = total / 10
  let size = num / cell

  let bar = "["
  for (let i = 0; i < 10; i++) {
    let fill = i > size ? " " : "="
    bar += fill
  }
  bar += "]"
  bar += total <= 0 ? "0%" : (Math.round(num / total * 10000) / 100.00) + "%"

  return bar
}

// console.log("---extract(data.hits.hits)", extract(data.hits.hits))
// let res = extract(data.hits.hits)
//
// let round = res.round
// let fields = res.fields;
//
// for (let i = 0; i < fields.length; i++) {
//   let field = fields[i]
//   console.groupCollapsed("%c[%s] %c%s", "color: #20B2AA;", field.type, "color: yellow;", field.name);
//   let values = field.values;
//   for (let j = 0; j < values.length; j++) {
//     let value = values[j]
//     let format = j === values.length - 1 ? "\t└─ %s\t%s" : "\t├─ %s\t%s"
//     console.log(format, percent(value.count, round), value.value)
//   }
//   console.groupEnd()
// }


export {indices, searchIndices, extractData, nodes, indicesDelete, indicesSearch}


