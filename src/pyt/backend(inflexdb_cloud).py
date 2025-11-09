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
# 1. MySQL 配置
MYSQL_HOST = 'localhost'
MYSQL_USER = 'root'
MYSQL_PASSWORD = 'cps@CPS123'  # 你的 MySQL 密码
MYSQL_DB = 'mqtt_data'
MYSQL_PORT = 3306

# 2. InfluxDB (本地) 配置
INFLUX_LOCAL_URL = "http://localhost:8086"
INFLUX_LOCAL_TOKEN = "7ucc4S8rrzwu85NA5nUYb_CNG7C-03Rbuyf2A85A5leATuxcPH_UlFvrCNXSGQtxvZQTuY_C6O7BUWNg4oIH-g==" # 你的本地 Token
INFLUX_LOCAL_ORG = "XJTU"
INFLUX_LOCAL_BUCKET = "class2down"

# 3. InfluxDB (云端) 配置
INFLUX_CLOUD_URL = "https://us-east-1-1.aws.cloud2.influxdata.com"
INFLUX_CLOUD_TOKEN = "zsOj7fEHWcI3DUOhFvhZ0iR39tRyWxko0oGAS5grC8PVyr-RhLe9X3WwHWPjdMJmiZRHwxqt2pE5XWi3TQ34Og==" # 你的云端 Token
INFLUX_CLOUD_ORG = "XJTU"
INFLUX_CLOUD_BUCKET = "class2down"

# --- HDFS 平台配置 (占位符) ---
HDFS_WEB_API_URL = "http://YOUR_HADOOP_HOST:9870/webhdfs/v1"
HDFS_TARGET_PATH = "/user/your_name/machine_data.csv"
HDFS_USER = "your_username"

# --- ML 模型配置 ---
MODEL_PATH = 'pyt/status_predictor.pkl'
DATA_WINDOW_SIZE = 50
model = None

# --- 初始化所有客户端 ---
try:
    influx_client_local = InfluxDBClient(url=INFLUX_LOCAL_URL, token=INFLUX_LOCAL_TOKEN, org=INFLUX_LOCAL_ORG)
    write_api_local = influx_client_local.write_api(write_options=SYNCHRONOUS)
    print("InfluxDB 本地客户端初始化成功。")
except Exception as e:
    print(f"InfluxDB 本地客户端初始化失败: {e}")

try:
    influx_client_cloud = InfluxDBClient(url=INFLUX_CLOUD_URL, token=INFLUX_CLOUD_TOKEN, org=INFLUX_CLOUD_ORG)
    write_api_cloud = influx_client_cloud.write_api(write_options=SYNCHRONOUS)
    print("InfluxDB 云端客户端初始化成功。")
except Exception as e:
    print(f"InfluxDB 云端客户端初始化失败: {e}")

# 加载 ML 模型
try:
    model = joblib.load(MODEL_PATH)
    print(f"机器学习模型 {MODEL_PATH} 加载成功。")
except Exception as e:
    print(f"模型加载失败: {e} (请先运行 _create_dummy_model.py)")

# HDFS 本地临时文件
local_temp_file_path = "hdfs_temp_data.csv"
local_file = None
try:
    # ！！！(修复 2 - 间接修复)！！！
    # 我们把临时文件放到了 'pyt' 文件夹内部，这样 debug 模式监视不到
    # （更好的修复是在 app.run 里设置 use_reloader=False）
    local_temp_file_path = "pyt/hdfs_temp_data.csv"
    local_file = open(local_temp_file_path, "a", encoding="utf-8")
    if local_file.tell() == 0:
        local_file.write("timestamp,data_id,value\n")
    print(f"打开本地临时文件 {local_temp_file_path} 成功。")
except Exception as e:
    print(f"打开本地临时文件失败: {e}")

@atexit.register
def upload_to_hdfs_on_exit():
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

# --- 数据库写入函数 (保持不变) ---
def save_to_mysql(data_id, value, time_str):
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

def save_to_influxdb_local(data_id, value, time_obj):
    try:
        value_float = float(value)
        p = Point("machine_data").tag("device_id", "STRESS_TEST_00000").tag("data_id", data_id).field("value", value_float).time(time_obj, WritePrecision.NS)
        write_api_local.write(bucket=INFLUX_LOCAL_BUCKET, org=INFLUX_LOCAL_ORG, record=p)
        print(f"InfluxDB 本地 写入成功: ID={data_id}, Value={value_float}")
    except Exception as e:
        print(f"InfluxDB 本地 写入失败: {e}")

def save_to_influxdb_cloud(data_id, value, time_obj):
    try:
        value_float = float(value)
        p = Point("machine_data_cloud") \
            .tag("device_id", "STRESS_TEST_00000") \
            .tag("data_id", data_id) \
            .field("value", value_float) \
            .time(time_obj, WritePrecision.NS)
        write_api_cloud.write(bucket=INFLUX_CLOUD_BUCKET, org=INFLUX_CLOUD_ORG, record=p)
        print(f"InfluxDB 云端 写入成功: ID={data_id}, Value={value_float}")
    except Exception as e:
        print(f"InfluxDB 云端 写入失败: {e}")

def save_to_distributed_platform(data_id, value, time_obj):
    global local_file
    if local_file:
        try:
            time_iso = time_obj.isoformat()
            line = f"{time_iso},{data_id},{value}\n"
            local_file.write(line)
            local_file.flush()
            print(f"本地 HDFS 缓存写入成功: ID={data_id}")
        except Exception as e:
            print(f"本地 HDFS 缓存写入失败: {e}")

# --- MQTT 回调 (保持不变) ---
@mqtt.on_connect()
def handle_connect(client, userdata, flags, rc):
    if rc == 0:
        print("MQTT 连接成功 (rc=0)")
        response_topic = "Query/Response/STRESS_TEST_00000"
        mqtt.subscribe(response_topic)
        print(f"已自动订阅主题: {response_topic}")
    else:
        print(f"MQTT 连接失败，返回码: {rc}")

@mqtt.on_message()
def handle_message(client, userdata, message):
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
                    
                    print(f"收到解析数据: ID={data_id}, Value={value}") # (新增日志)
                    
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
                    
                    # 5. 写入 InfluxDB (云端)
                    save_to_influxdb_cloud(data_id, value, time_now_utc)
                    
                    # 6. 写入 HDFS 缓存
                    save_to_distributed_platform(data_id, value, time_now_utc)
            
    except Exception as e:
        print(f"处理消息失败: {e}")

# --- Flask API 路由 (修复 /get_data/) ---
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
  
    global latest_data_store 
    
    try:
        data = request.get_json()
        data_id = data.get('id')
        value = latest_data_store.get(data_id) 
        if value is not None:
            print(f"前端获取数据成功: ID={data_id}, Value={value}") # (新增日志)
            return jsonify({'status': 'ok', 'id': data_id, 'value': value})
        else:
            print(f"前端获取数据失败: ID={data_id}, 数据尚未收到(None)") # (新增日志)
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
    app.run(debug=True, use_reloader=False, host='127.0.0.1', port=5000)