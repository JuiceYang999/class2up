## **任务四：Web 网页的设计与数据可视化**

此任务的目标是：

1.  **订阅数据：** 让后端订阅 MQTT 主题以接收数据。
2.  **请求数据：** 从前端按时发起数据请求（发布消息）。
3.  **获取数据：** 后端将收到的数据传递给前端。
4.  **可视化：** 前端 `ShowData.vue` 页面使用 `Echarts` 实时显示数据。

我们将采用“**前端定时轮询**”的方式：

  * 前端 `ShowData.vue` 页面每隔3秒向后端 **发布** 一个数据请求。
  * 同时，`ShowData.vue` 向后端 **获取** 上一次收到的数据。
  * 后端 `backend.py` 负责订阅和收发消息，并暂存数据。

-----

### 4.1 安装 Echarts 依赖

在您回到主机后，请在项目根目录（`.../class2up-9730b2fcdf474be54786a1f787cc6b005a26d2fc/`）打开终端，运行以下命令来安装 `echarts`：

```bash
yarn add echarts
```

*(这会更新您的 `package.json` 和 `yarn.lock` 文件)*

-----

### 4.2 后端 (Python) 代码更新

我们需要大幅更新 `backend.py`，使其具备 **订阅**、**发布** 和 **数据暂存/提供** 的能力。

请**替换** `pyt/backend.py` 文件的**全部内容**为以下代码：

**文件路径：** `juiceyang999/class2up/class2up-9730b2fcdf474be54786a1f787cc6b005a26d2fc/pyt/backend.py`

```python
from flask import Flask, request, jsonify
from flask_cors import CORS
from flask_mqtt import Mqtt
import json # 引入 json 库

# 1. 初始化 Flask 应用
app = Flask(__name__)
CORS(app)

# 2. 配置 Flask-MQTT
app.config['MQTT_BROKER_URL'] = '127.0.0.1'
app.config['MQTT_BROKER_PORT'] = 1883
app.config['MQTT_CLIENT_ID'] = 'flask_mqtt_client'
app.config['MQTT_KEEPALIVE'] = 60
app.config['MQTT_TLS_ENABLED'] = False

# 3. 初始化 MQTT 客户端
mqtt = Mqtt(app)

# 4. (新增) 全局变量，用于暂存从MQTT收到的最新数据
# 我们使用字典，以数据ID为键，方便前端按ID查询
latest_data_store = {}

# --- MQTT 事件回调 ---

@mqtt.on_connect()
def handle_connect(client, userdata, flags, rc):
    if rc == 0:
        print("MQTT 连接成功 (rc=0)")
        # 5. (新增) 一旦连接成功，立刻订阅“响应”主题
        # 这是虚拟数控系统返回数据的主题
        response_topic = "Query/Response/STRESS_TEST_00000"
        mqtt.subscribe(response_topic)
        print(f"已自动订阅主题: {response_topic}")
    else:
        print(f"MQTT 连接失败，返回码: {rc}")

@mqtt.on_message()
def handle_message(client, userdata, message):
    # 6. (更新) 当收到消息时，解析并存储它
    try:
        payload_str = message.payload.decode()
        print(f"收到消息 (Topic: {message.topic}): {payload_str}")
        
        # 解析JSON数据
        data = json.loads(payload_str)
        
        # 根据《项目课程2025.pdf》P30 的数据格式 {"values":[{"id":"...","values":[...]}]}
        if 'values' in data and isinstance(data['values'], list):
            for item in data['values']:
                if 'id' in item and 'values' in item:
                    # 我们只存储最新的值
                    latest_data_store[item['id']] = item['values'][0] # 假设我们只关心第一个值
            
            print(f"已更新数据存储: {latest_data_store}")

    except Exception as e:
        print(f"处理消息失败: {e}")

# --- Flask API 路由 ---

@app.route('/connect/', methods=['POST', 'GET'])
def make_connect():
    try:
        data_connect = request.get_json()['data']
        print(f"收到来自前端的连接请求: {data_connect}")

        app.config['MQTT_BROKER_URL'] = data_connect['host']
        app.config['MQTT_BROKER_PORT'] = data_connect['port']
        app.config['MQTT_CLIENT_ID'] = data_connect['clientid']
        
        # 重新配置并连接
        # 注意：在生产环境中，更健壮的连接管理是必要的
        if mqtt.client.is_connected():
            mqtt.client.disconnect()
            
        mqtt.client.username_pw_set(None, None) # 清除旧凭证
        mqtt.client._client_id = data_connect['clientid'].encode()
        mqtt.client.reinitialise()
        mqtt.client.connect(data_connect['host'], data_connect['port'])

        print(f"正在尝试连接到 {data_connect['host']}:{data_connect['port']}...")
        return jsonify({'rc_status': 'success'})

    except Exception as e:
        print(f"连接失败: {e}")
        return jsonify({'rc_status': str(e)})

# 7. (新增) /publish/ 接口
# 用于让前端“请求”虚拟数控系统发送数据
@app.route('/publish/', methods=['POST'])
def do_publish():
    try:
        data = request.get_json()
        topic = data.get('topic')
        payload = data.get('payload', "{}") # payload 应该是 JSON 字符串
        
        if not topic:
            return jsonify({'status': 'error', 'message': 'Topic is required'}), 400
            
        print(f"正在发布消息到 (Topic: {topic}): {payload}")
        mqtt.publish(topic, payload)
        
        return jsonify({'status': 'published', 'topic': topic})
    except Exception as e:
        print(f"发布失败: {e}")
        return jsonify({'status': 'error', 'message': str(e)}), 500

# 8. (新增) /get_data/ 接口
# 用于让前端获取已存储的最新数据
@app.route('/get_data/', methods=['POST'])
def get_data():
    try:
        data = request.get_json()
        data_id = data.get('id') # 前端将请求特定ID的数据
        
        if not data_id:
            return jsonify({'status': 'error', 'message': 'ID is required'}), 400
        
        # 从我们的存储中获取最新值
        value = latest_data_store.get(data_id)
        
        if value is not None:
            return jsonify({'status': 'ok', 'id': data_id, 'value': value})
        else:
            return jsonify({'status': 'not_found', 'id': data_id, 'value': None})
            
    except Exception as e:
        print(f"获取数据失败: {e}")
        return jsonify({'status': 'error', 'message': str(e)}), 500

# 9. 启动 Flask 服务
if __name__ == '__main__':
    app.run(debug=True, host='127.0.0.1', port=5000)
```

-----

### 4.3 前端 (VUE) 界面修改

#### 4.3.1 修改 `CoreSource.vue`（数据采集界面）

我们需要在 `CoreSource.vue` 页面添加“发布数据请求”的功能，以模拟《实验步骤.pdf》P2 中 MQTTX 的“发布”操作。

请**替换** `src/components/CoreSource.vue` 文件的**全部内容**为以下代码：

**文件路径：** `juiceyang999/class2up/class2up-9730b2fcdf474be54786a1f787cc6b005a26d2fc/src/components/CoreSource.vue`

```vue
<template>
  <el-container>
    <el-header>
      <h2>数据采集与服务器连接</h2>
    </el-header>
    <el-main>
      <el-row :gutter="20">
        <el-col :span="12">
          <el-card class="box-card">
            <div slot="header" class="clearfix">
              <span>MQTT 服务器连接配置</span>
            </div>
            <el-form ref="form" :model="connection" label-width="120px">
              <el-form-item label="服务器地址">
                <el-input v-model="connection.host"></el-input>
              </el-form-item>
              <el-form-item label="端口号">
                <el-input v-model.number="connection.port" type="number"></el-input>
              </el-form-item>
              <el-form-item label="Client ID">
                <el-input v-model="connection.clientid"></el-input>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="onConnect">连接</el-button>
                <el-button @click="onDisconnect" :disabled="!connectStatus">断开</el-button>
              </el-form-item>
              <el-form-item label="连接状态">
                <el-tag :type="connectStatus ? 'success' : 'danger'">
                  {{ connectStatus ? '已连接' : '未连接' }}
                </el-tag>
              </el-form-item>
            </el-form>
          </el-card>
        </el-col>
        
        <el-col :span="12">
          <el-card class="box-card">
            <div slot="header" class="clearfix">
              <span>(调试) 手动发布数据请求</span>
            </div>
            <el-form label-width="120px">
              <el-form-item label="请求主题">
                <el-input v-model="publish.topic"></el-input>
              </el-form-item>
              <el-form-item label="请求 Payload">
                <el-input type="textarea" :rows="3" v-model="publish.payload"></el-input>
              </el-form-item>
              <el-form-item>
                <el-button type="warning" @click="onPublish" :disabled="!connectStatus">
                  手动发布
                </el-button>
              </el-form-item>
            </el-form>
            <div style="font-size: 12px; color: #909399;">
              <p>说明：后端连接成功后会自动订阅响应主题。此卡片仅用于手动调试数据请求。</p>
              <p>例如，根据《实验步骤.pdf》P2，Payload 可填：</p>
              <p>{"ids":[{"id":"0103502202"}]}</p>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </el-main>
  </el-container>
</template>

<script>
import axios from "axios";
axios.defaults.baseURL = "/api";

export default {
  name: "CoreSource",
  data() {
    return {
      connection: {
        host: "127.0.0.1",
        port: 1883,
        clientid: "vue_client_" + Math.random().toString(16).substr(2, 8),
      },
      connectStatus: false,
      // (新增) 存储发布信息
      publish: {
        topic: "Query/Request/STRESS_TEST_00000",
        // '0103502202' 是X轴实际速度
        payload: '{"ids":[{"id":"0103502202"}]}' 
      }
    };
  },
  methods: {
    onConnect() {
      console.log("正在连接到MQTT服务器...");
      axios
        .post("/connect/", { data: this.connection })
        .then((res) => {
          if (res.data && res.data.rc_status === "success") {
            this.connectStatus = true;
            this.$message.success('MQTT服务器连接成功！');
          } else {
            this.connectStatus = false;
            this.$message.error('连接失败: ' + res.data.rc_status);
          }
        })
        .catch((err) => {
          console.error("连接请求失败:", err);
          this.connectStatus = false;
          this.$message.error('后端服务连接失败，请检查Python后端是否已运行。');
        });
    },
    onDisconnect() {
      // (功能待定，目前仅用于前端演示)
      console.log("断开连接");
      this.connectStatus = false;
      this.$message.info('已断开连接（前端模拟）');
      // 实际开发中，这里也应该向后端发送一个 /disconnect/ 请求
    },
    // (新增) 发布方法
    onPublish() {
      axios.post("/publish/", {
        topic: this.publish.topic,
        payload: this.publish.payload
      })
      .then(res => {
        if (res.data && res.data.status === 'published') {
          this.$message.success('请求已发布！请切换到“数据显示”页面查看。');
        } else {
          this.$message.error('发布失败: ' + res.data.message);
        }
      })
      .catch(err => {
        this.$message.error('发布请求失败: ' + err);
      });
    }
  },
};
</script>

<style scoped>
.el-header {
  background-color: #b3c0d1;
  color: #333;
  line-height: 60px;
}
.box-card {
  text-align: left;
}
</style>
```

#### 4.3.2 编写 `ShowData.vue`（数据显示界面）

这是本任务的核心文件。我们将在这里初始化 Echarts 并定时请求和刷新数据。

请**替换** `src/components/ShowData.vue` 文件的**全部内容**为以下代码：

**文件路径：** `juiceyang999/class2up/class2up-9730b2fcdf474be54786a1f787cc6b005a26d2fc/src/components/ShowData.vue`

```vue
<template>
  <el-container>
    <el-main>
      <el-card>
        <div slot="header">
          <span>实时数据显示 (X轴实际速度)</span>
          <span style="font-size: 12px; color: #909399; margin-left: 20px;">
            (数据ID: {{ dataIdToFetch }})
          </span>
        </div>
        <div id="data-chart" style="width: 100%; height: 400px;"></div>
      </el-card>
    </el-main>
  </el-container>
</template>

<script>
import axios from "axios";
// 2. 引入 Echarts
import * as echarts from "echarts";

axios.defaults.baseURL = "/api";

export default {
  name: "ShowData",
  data() {
    return {
      chart: null, // Echarts 实例
      chartData: [], // 存储图表的数据
      dataTimer: null, // 存储 setInterval 的ID
      // 3. 我们要抓取的数据ID：X轴实际速度
      // (来自《项目课程2025.pdf》P29, "0103502202" "SPEED" "实际速度")
      dataIdToFetch: '0103502202',
      // 请求发布的主题和payload
      publishConfig: {
        topic: "Query/Request/STRESS_TEST_00000",
        payload: '{"ids":[{"id":"0103502202"}]}' // 动态生成
      },
    };
  },
  mounted() {
    // 4. 页面加载时：初始化图表并开始抓取数据
    // 动态设置我们要请求的ID
    this.publishConfig.payload = `{"ids":[{"id":"${this.dataIdToFetch}"}]}`;
    
    this.initChart();
    this.startFetching();
  },
  beforeDestroy() {
    // 5. 页面销毁时：停止定时器，防止内存泄漏
    if (this.dataTimer) {
      clearInterval(this.dataTimer);
    }
  },
  methods: {
    // 6. 初始化图表
    initChart() {
      // 对应《实验步骤.pdf》P12/P13的初始化代码
      const chartDom = document.getElementById("data-chart");
      this.chart = echarts.init(chartDom);
      
      const option = {
        xAxis: {
          type: "category",
          data: [], // X轴数据（时间戳或索引）
        },
        yAxis: {
          type: "value",
          scale: true // 自动缩放
        },
        series: [
          {
            data: [], // Y轴数据（速度值）
            type: "line",
            smooth: true,
          },
        ],
        tooltip: {
          trigger: 'axis'
        }
      };
      
      this.chart.setOption(option);
    },
    
    // 7. 开始轮询
    startFetching() {
      // 每3秒钟执行一次 requestAndFetch
      this.requestAndFetch(); // 立即执行一次
      this.dataTimer = setInterval(this.requestAndFetch, 3000); // 3000毫秒 = 3秒
    },

    // 8. (新增) 组合方法：先请求，再获取
    requestAndFetch() {
      // 8.1. 第一步：向后端发请求，让后端去 "问" 虚拟数控机
      axios.post("/publish/", this.publishConfig)
        .catch(err => {
          console.error("发布数据请求失败:", err);
        });

      // 8.2. 第二步：稍等片刻（例如500ms），再去后端 "取" 数据
      // (给MQTT一个来回的时间)
      setTimeout(() => {
        this.fetchData();
      }, 500);
    },

    // 9. (新增) 从后端获取数据并更新图表
    fetchData() {
      axios.post("/get_data/", { id: this.dataIdToFetch })
        .then(res => {
          if (res.data && res.data.status === 'ok' && res.data.value !== null) {
            const newValue = parseFloat(res.data.value); // 确保是数字
            
            // 9.1. (新增) 管理数据数组，保持其长度不超过50
            if (this.chartData.length > 50) {
              this.chartData.shift(); // 移除第一个元素
            }
            this.chartData.push(newValue);
            
            // 9.2. 生成X轴标签（简单地使用索引）
            const xAxisData = this.chartData.map((_, index) => index);

            // 9.3. 更新 Echarts 图表
            this.chart.setOption({
              xAxis: {
                data: xAxisData,
              },
              series: [
                {
                  data: this.chartData,
                },
              ],
            });

          } else if (res.data.status === 'not_found') {
            console.log(`数据ID ${this.dataIdToFetch} 尚未收到.`);
          }
        })
        .catch(err => {
          console.error("获取数据失败:", err);
        });
    }
  },
};
</script>

<style scoped>
/* 可以在这里添加此页面专属的样式 */
</style>
```

-----

### 4.4 运行与调试（当您回到主机时）

1.  **启动环境（共4个）：** (同任务三)

      * **终端 1：** `emqx start`
      * **终端 2：** `StressTest-INC-Cloud -n 1 -b 0` (保持运行)
      * **终端 3：** `yarn serve` (VUE 前端)
      * **终端 4：** `python pyt/backend.py` (Python 后端)

2.  **测试：**

      * 打开浏览器访问 VUE 页面（如 `http://localhost:8080`）。
      * **步骤 1:** 点击导航栏 “**数据采集**”。
      * **步骤 2:** 点击 “**连接**” 按钮。确保提示 “MQTT服务器连接成功！”。
      * **步骤 3:** 点击导航栏 “**数据显示**”。
      * **预期结果：**
        1.  您会看到一个空的 Echarts 图表。
        2.  等待几秒钟...
        3.  图表上开始**实时**出现一条线，并不断向右滚动，显示来自虚拟数控机床的 "X轴实际速度" 数据。
        4.  在 **终端 4** (Python 后端) 中，您会看到 `收到消息...` 和 `已更新数据存储...` 的打印信息。

**任务四完成**。我们已经成功地从前端请求数据，通过后端与MQTT通信，并将获取到的实时数据展示在了Echarts图表上。

接下来，我们将进行**任务五：基于数据库的数据存储 (MySQL & InfluxDB)**。