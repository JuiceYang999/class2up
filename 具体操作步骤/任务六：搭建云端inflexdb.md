我们将分四步走：

1.  **注册 InfluxDB Cloud 并创建 Bucket (在浏览器)**
2.  **获取云端 Token 和 Org (在浏览器)**
3.  **修改 `pyt/backend.py` 代码 (在 VSCode)**
4.  **运行和验证**

-----

### 步骤 1: 注册和创建 Bucket (在 InfluxDB Cloud 网站)

1.  **打开网页**：
    打开《实验步骤.pdf》中提供的云端服务网页：`https://us-east-1-1.aws.cloud2.influxdata.com`。

2.  **注册账号**：

      * 点击 "Sign Up" 或 "Get Started"。
      * 填写你的邮箱、密码、公司名（随便填）等信息完成注册。
      * 登录后，它可能会让你选择一个云服务商（AWS, GCP, Azure）和区域（Region），**你可以随便选一个免费的**（比如 AWS 的 us-east-1）。

3.  **创建 Bucket**：

      * 登录成功后，你会进入 InfluxDB Cloud 的主界面。
      * 在左侧菜单栏，找到一个像 “桶” 一样的图标，点击 “**Load Data**” -\> “**Buckets**”。
      * 点击右上角的 “**+ Create Bucket**”。
      * 给你的云端存储桶起一个新名字，例如：`my_cloud_bucket`。（不要和本地的 `mqtt_bucket` 搞混了）
      * 点击 "Create"。

-----

### 步骤 2: 获取云端 Token 和 Org (在 InfluxDB Cloud 网站)

现在，你需要 3 个关键信息才能让 Python 连接到云端。**请把这 3 个信息复制到一个记事本里**：

1.  **你的 Token (API 令牌)**：

      * 在左侧菜单栏，点击 “**Load Data**” -\> “**API Tokens**”。
      * 你会看到一个已经存在的、以你名字命名的 Token（例如 `Your_Name's Token`）。
      * **点击这个 Token 的名字**。
      * 点击 “**Copy to Clipboard**”（复制到剪贴板）按钮。
      * **这就是你的 `INFLUX_CLOUD_TOKEN`**。

2.  **你的 Org (组织名称)**：

      * 你的组织名称通常显示在**左上角**，在你的邮箱地址下面。
      * **这就是你的 `INFLUX_CLOUD_ORG`**。

3.  **你的 URL (云端地址)**：

      * 在左侧菜单栏，点击最下方的“个人”图标，然后选择 “**About**”。
      * 你会看到 “Cloud URL”，它就是 `https://us-east-1-1.aws.cloud2.influxdata.com`（或者根据你选择的区域而定）。
      * **这就是你的 `INFLUX_CLOUD_URL`**。

-----

### 步骤 3: 修改 `pyt/backend.py` 代码 (在 VSCode)

现在我们来修改 Python 后端，让它**同时**向**本地 InfluxDB** 和**云端 InfluxDB** 写入数据。

请**替换** `pyt/backend.py` 文件的**全部内容**为以下代码：

**文件路径：** `.../class2up/pyt/backend.py`

```python
from flask import Flask, request, jsonify
from flask_cors import CORS
from flask_mqtt import Mqtt
import json
import pymysql
import datetime
from influxdb_client import InfluxDBClient, Point, WritePrecision
from influxdb_client.client.write_api import SYNCHRONOUS
import requests
import atexit
import joblib
import numpy as np

# --- 数据库配置 ---
# 1. MySQL 配置 (保持不变)
MYSQL_HOST = 'localhost'
MYSQL_USER = 'root'
MYSQL_PASSWORD = 'xjtu2025'  # ！！！请修改为您的MySQL密码！！！
MYSQL_DB = 'mqtt_data'
MYSQL_PORT = 3306

# 2. InfluxDB (本地) 配置 (保持不变)
INFLUX_LOCAL_URL = "http://localhost:8086"
INFLUX_LOCAL_TOKEN = "YOUR_API_TOKEN_HERE"  # ！！！请修改为您的本地 Token！！！
INFLUX_LOCAL_ORG = "my-org"                 # ！！！请修改为您的本地组织！！！
INFLUX_LOCAL_BUCKET = "mqtt_bucket"         # ！！！请修改为您的本地存储桶！！！

# 3. (新增) InfluxDB (云端) 配置
# ！！！ 请用你在 步骤 2 中复制的信息替换这些占位符 ！！！
INFLUX_CLOUD_URL = "https://us-east-1-1.aws.cloud2.influxdata.com"
INFLUX_CLOUD_TOKEN = "YOUR_CLOUD_TOKEN_HERE" # ！！！请修改为您的云端 Token！！！
INFLUX_CLOUD_ORG = "YOUR_CLOUD_ORG_HERE"     # ！！！请修改为您的云端 Org！！！
INFLUX_CLOUD_BUCKET = "my_cloud_bucket"      # (你刚刚创建的云端 Bucket)

# --- (新增) HDFS 平台配置 (占位符) ---
HDFS_WEB_API_URL = "http://YOUR_HADOOP_HOST:9870/webhdfs/v1"
HDFS_TARGET_PATH = "/user/your_name/machine_data.csv"
HDFS_USER = "your_username"

# --- (新增) ML 模型配置 ---
MODEL_PATH = 'pyt/status_predictor.pkl'
DATA_WINDOW_SIZE = 50
model = None

# --- 初始化所有客户端 ---
try:
    # (修改) 初始化本地 InfluxDB 客户端
    influx_client_local = InfluxDBClient(url=INFLUX_LOCAL_URL, token=INFLUX_LOCAL_TOKEN, org=INFLUX_LOCAL_ORG)
    write_api_local = influx_client_local.write_api(write_options=SYNCHRONOUS)
    print("InfluxDB 本地客户端初始化成功。")
except Exception as e:
    print(f"InfluxDB 本地客户端初始化失败: {e}")

try:
    # (新增) 初始化云端 InfluxDB 客户端
    influx_client_cloud = InfluxDBClient(url=INFLUX_CLOUD_URL, token=INFLUX_CLOUD_TOKEN, org=INFLUX_CLOUD_ORG)
    write_api_cloud = influx_client_cloud.write_api(write_options=SYNCHRONOUS)
    print("InfluxDB 云端客户端初始化成功。")
except Exception as e:
    print(f"InfluxDB 云端客户端初始化失败: {e}")

# (新增) 加载 ML 模型
try:
    model = joblib.load(MODEL_PATH)
    print(f"机器学习模型 {MODEL_PATH} 加载成功。")
except Exception as e:
    print(f"模型加载失败: {e} (请先运行 _create_dummy_model.py)")

# HDFS 本地临时文件
local_temp_file_path = "hdfs_temp_data.csv"
local_file = None
try:
    local_file = open(local_temp_file_path, "a", encoding="utf-8")
    if local_file.tell() == 0:
        local_file.write("timestamp,data_id,value\n")
    print(f"打开本地临时文件 {local_temp_file_path} 成功。")
except Exception as e:
    print(f"打开本地临时文件失败: {e}")

@atexit.register
def upload_to_hdfs_on_exit():
    # ... (此函数保持不变) ...
    global local_file
    if local_file:
        local_file.close()
        print(f"程序退出，正在将 {local_temp_file_path} 上传到 HDFS...")
        try:
            create_url = f"{HDFS_WEB_API_URL}{HDFS_TARGET_PATH}?op=CREATE&overwrite=true&user.name={HDFS_USER}"
            r_create = requests.put(create_url, allow_redirects=False)
            if r_create.status_code == 307:
                data_node_url = r_create.headers['Location']
                with open(local_temp_file_path, "rb") as f:
                    r_upload = requests.put(data_node_url, data=f)
                    if r_upload.status_code == 201:
                        print(f"HDFS 文件上传成功: {HDFS_TARGET_PATH}")
                    else:
                        print(f"HDFS 文件上传失败: {r_upload.status_code} - {r_upload.text}")
            else:
                 print(f"HDFS 'CREATE' 请求失败: {r_create.status_code} - {r_create.text}")
        except Exception as e:
            print(f"HDFS 上传时发生严重错误: {e}")

# --- Flask 和 MQTT 初始化 (保持不变) ---
app = Flask(__name__)
CORS(app)
app.config['MQTT_BROKER_URL'] = '127.0.0.1'
app.config['MQTT_BROKER_PORT'] = 1883
app.config['MQTT_CLIENT_ID'] = 'flask_mqtt_client'
mqtt = Mqtt(app)

# --- 数据存储 (保持不变) ---
latest_data_store = {}
data_window_store = {}

# --- 数据库写入函数 ---
def save_to_mysql(data_id, value, time_str):
    # ... (此函数保持不变) ...
    conn = None
    cur = None
    try:
        conn = pymysql.connect(host=MYSQL_HOST, user=MYSQL_USER, password=MYSQL_PASSWORD, database=MYSQL_DB, port=MYSQL_PORT, charset='utf8')
        cur = conn.cursor()
        sql = "INSERT INTO mac_data (data_id, payload, time) VALUES (%s, %s, %s)"
        cur.execute(sql, (data_id, str(value), time_str))
        conn.commit()
        print(f"MySQL 写入成功: ID={data_id}, Value={value}")
    except Exception as e:
        print(f"MySQL 写入失败: {e}")
    finally:
        if cur: cur.close()
        if conn: conn.close()

def save_to_influxdb_local(data_id, value, time_obj): # (函数名修改)
    try:
        value_float = float(value)
        p = Point("machine_data").tag("device_id", "STRESS_TEST_00000").tag("data_id", data_id).field("value", value_float).time(time_obj, WritePrecision.NS)
        write_api_local.write(bucket=INFLUX_LOCAL_BUCKET, org=INFLUX_LOCAL_ORG, record=p) # (变量名修改)
        print(f"InfluxDB 本地 写入成功: ID={data_id}, Value={value_float}")
    except Exception as e:
        print(f"InfluxDB 本地 写入失败: {e}")

# (新增) 写入 InfluxDB 云端
def save_to_influxdb_cloud(data_id, value, time_obj):
    try:
        value_float = float(value)
        p = Point("machine_data_cloud") \
            .tag("device_id", "STRESS_TEST_00000") \
            .tag("data_id", data_id) \
            .field("value", value_float) \
            .time(time_obj, WritePrecision.NS)
        
        # 使用云端配置的 API 写入
        write_api_cloud.write(bucket=INFLUX_CLOUD_BUCKET, org=INFLUX_CLOUD_ORG, record=p)
        print(f"InfluxDB 云端 写入成功: ID={data_id}, Value={value_float}")
    except Exception as e:
        print(f"InfluxDB 云端 写入失败: {e}")

def save_to_distributed_platform(data_id, value, time_obj):
    # ... (此函数保持不变) ...
    global local_file
    if local_file:
        try:
            time_iso = time_obj.isoformat()
            line = f"{time_iso},{data_id},{value}\n"
            local_file.write(line)
            local_file.flush()
        except Exception as e:
            print(f"本地临时文件写入失败: {e}")

# --- MQTT 回调 (更新) ---
@mqtt.on_connect()
def handle_connect(client, userdata, flags, rc):
    # ... (此函数保持不变) ...
    if rc == 0:
        print("MQTT 连接成功 (rc=0)")
        response_topic = "Query/Response/STRESS_TEST_00000"
        mqtt.subscribe(response_topic)
        print(f"已自动订阅主题: {response_topic}")
    else:
        print(f"MQTT 连接失败，返回码: {rc}")

@mqtt.on_message()
def handle_message(client, userdata, message):
    # (更新)
    try:
        payload_str = message.payload.decode()
        data = json.loads(payload_str)
        
        time_now_utc = datetime.datetime.utcnow()
        time_now_local_str = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")

        if 'values' in data and isinstance(data['values'], list):
            for item in data['values']:
                if 'id' in item and 'values' in item:
                    data_id = item['id']
                    value = item['values'][0]
                    
                    # 1. 暂存 (供 Echarts)
                    latest_data_store[data_id] = value

                    # 2. 存储数据窗口 (供 ML)
                    if data_id not in data_window_store:
                        data_window_store[data_id] = []
                    data_window_store[data_id].append(float(value))
                    if len(data_window_store[data_id]) > DATA_WINDOW_SIZE:
                        data_window_store[data_id].pop(0)

                    # 3. 写入 MySQL
                    save_to_mysql(data_id, value, time_now_local_str)
                    
                    # 4. 写入 InfluxDB (本地)
                    save_to_influxdb_local(data_id, value, time_now_utc)
                    
                    # 5. (新增) 写入 InfluxDB (云端)
                    save_to_influxdb_cloud(data_id, value, time_now_utc)
                    
                    # 6. 写入 HDFS 缓存
                    save_to_distributed_platform(data_id, value, time_now_utc)
            
    except Exception as e:
        print(f"处理消息失败: {e}")

# --- Flask API 路由 (保持不变) ---
@app.route('/connect/', methods=['POST', 'GET'])
def make_connect():
    # ... (此函数保持不变) ...
    try:
        data_connect = request.get_json()['data']
        app.config['MQTT_BROKER_URL'] = data_connect['host']
        app.config['MQTT_BROKER_PORT'] = data_connect['port']
        app.config['MQTT_CLIENT_ID'] = data_connect['clientid']
        if mqtt.client.is_connected():
            mqtt.client.disconnect()
        mqtt.client.username_pw_set(None, None)
        mqtt.client._client_id = data_connect['clientid'].encode()
        mqtt.client.reinitialise()
        mqtt.client.connect(data_connect['host'], data_connect['port'])
        return jsonify({'rc_status': 'success'})
    except Exception as e:
        return jsonify({'rc_status': str(e)})

@app.route('/publish/', methods=['POST'])
def do_publish():
    # ... (此函数保持不变) ...
    try:
        data = request.get_json()
        topic = data.get('topic')
        payload = data.get('payload', "{}")
        mqtt.publish(topic, payload)
        return jsonify({'status': 'published', 'topic': topic})
    except Exception as e:
        return jsonify({'status': 'error', 'message': str(e)}), 500

@app.route('/get_data/', methods=['POST'])
def get_data():
    # ... (此函数保持不变) ...
    global latest_data_store 
    try:
        data = request.get_json()
        data_id = data.get('id')
        value = latest_data_store.get(data_id) 
        if value is not None:
            return jsonify({'status': 'ok', 'id': data_id, 'value': value})
        else:
            return jsonify({'status': 'not_found', 'id': data_id, 'value': None})
    except Exception as e:
        return jsonify({'status': 'error', 'message': str(e)}), 500

@app.route('/predict/', methods=['POST'])
def predict_status():
    # ... (此函数保持不变) ...
    global model
    if model is None:
        return jsonify({'status': 'error', 'message': '模型未加载'}), 500
    try:
        data = request.get_json()
        data_id = data.get('id')
        window_data = data_window_store.get(data_id)
        if not window_data or len(window_data) < DATA_WINDOW_SIZE:
            return jsonify({'status': 'wait', 'message': f'数据采集中，当前 {len(window_data) if window_data else 0}/{DATA_WINDOW_SIZE} 点，请稍候...'})
        features_vector = np.array(window_data).reshape(1, -1)
        prediction_code = model.predict(features_vector)
        status_map = {0: '设备状态正常', 1: '设备故障预警'}
        prediction_text = status_map.get(prediction_code[0], '未知状态')
        return jsonify({'status': 'ok', 'prediction': prediction_text})
    except Exception as e:
        return jsonify({'status': 'error', 'message': str(e)}), 500

# 启动 Flask 服务
if __name__ == '__main__':
    app.run(debug=True, host='127.0.0.1', port=5000)
```

-----

### 步骤 4: 运行和验证

1.  **保存代码**：

      * 确保你把从 InfluxDB Cloud 网站复制的 `URL`, `TOKEN`, 和 `ORG` **正确地**粘贴到了 `backend.py` 的 `INFLUX_CLOUD_...` 变量里。

2.  **重启后端**：

      * **停止**你的 Python 后端（终端 5）(按 `Ctrl+C`)。
      * **重新运行** `python pyt/backend.py`。
      * **（重要！）** 检查终端输出，你现在应该看到**两个**成功信息：
        ```
        InfluxDB 本地客户端初始化成功。
        InfluxDB 云端客户端初始化成功。
        ```
      * 如果云端初始化失败，请回去**仔细检查**你的 URL, Token 和 Org 是不是复制错了。

3.  **验证写入**：

      * 让所有服务（EMQX, StressTest, Influxd, VUE, Python）运行一两分钟。
      * 你的 Python 终端（终端 5）现在应该**同时**打印：
        ```
        InfluxDB 本地 写入成功: ...
        InfluxDB 云端 写入成功: ...
        ```

4.  **在云端查看数据**：

      * 回到你的 **InfluxDB Cloud 浏览器窗口**。
      * 点击左侧菜单的 “**Data Explorer**” (或 "Explore")。
      * 在查询构建器 (Query Builder) 中：
          * **FROM:** 选择你的云端存储桶 `my_cloud_bucket`。
          * **Filter (Measurement):** 选择 `machine_data_cloud`。
          * **Filter (Field):** 选择 `value`。
      * 点击 “**Submit**”。

**预期结果：**
你会在 **InfluxDB Cloud 网站**上看到一个**一模一样**的实时数据图表，证明你本地的数据已经成功同步到云端了！

这就完成了《实验步骤.pdf》中第 (7) 步的要求。