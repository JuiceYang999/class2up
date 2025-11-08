from flask import Flask, request, jsonify
from flask_cors import CORS
from flask_mqtt import Mqtt
import json
import pymysql       # (新增) 引入 MySQL 库
import datetime      # (新增) 引入时间库
from influxdb_client import InfluxDBClient, Point, WritePrecision # (新增) 引入 InfluxDB 库
from influxdb_client.client.write_api import SYNCHRONOUS

# --- (新增) 数据库配置 ---
# 1. MySQL 配置 (请根据您的设置修改)
MYSQL_HOST = 'localhost'
MYSQL_USER = 'root'
MYSQL_PASSWORD = 'cps@CPS123'  # ！！！请修改为您的MySQL密码！！！
MYSQL_DB = 'mqtt_data'
MYSQL_PORT = 3306

# 2. InfluxDB v2 配置 (请根据您在 5.1.2 中设置和复制的值修改)
INFLUX_URL = "http://localhost:8086"
INFLUX_TOKEN = "7ucc4S8rrzwu85NA5nUYb_CNG7C-03Rbuyf2A85A5leATuxcPH_UlFvrCNXSGQtxvZQTuY_C6O7BUWNg4oIH-g=="  # ！！！请修改为您复制的 Token！！！
INFLUX_ORG = "XJTU"                 # ！！！请修改为您的组织名称！！！
INFLUX_BUCKET = "class2down"         # ！！！请修改为您的存储桶名称！！！

# 3. (新增) 初始化 InfluxDB 客户端
try:
    influx_client = InfluxDBClient(url=INFLUX_URL, token=INFLUX_TOKEN, org=INFLUX_ORG)
    # 使用同步写入模式
    write_api = influx_client.write_api(write_options=SYNCHRONOUS)
    print("InfluxDB 客户端初始化成功。")
except Exception as e:
    print(f"InfluxDB 客户端初始化失败: {e}")
# -------------------------

# 初始化 Flask 应用
app = Flask(__name__)
CORS(app)

# 配置 Flask-MQTT
app.config['MQTT_BROKER_URL'] = '127.0.0.1'
app.config['MQTT_BROKER_PORT'] = 1883
app.config['MQTT_CLIENT_ID'] = 'flask_mqtt_client'
app.config['MQTT_KEEPALIVE'] = 60
app.config['MQTT_TLS_ENABLED'] = False

mqtt = Mqtt(app)

# 暂存最新数据
latest_data_store = {}

# --- (新增) 数据库写入函数 ---

def save_to_mysql(data_id, value):
    """
    将数据保存到 MySQL 数据库
    """
    conn = None
    cur = None
    try:
        # 获取当前时间
        time_str = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        
        # 建立连接
        conn = pymysql.connect(
            host=MYSQL_HOST,
            user=MYSQL_USER,
            password=MYSQL_PASSWORD,
            database=MYSQL_DB,
            port=MYSQL_PORT,
            charset='utf8'
        )
        # 创建游标
        cur = conn.cursor()
        
        # SQL 插入语句 (使用参数化查询防止SQL注入)
        sql = "INSERT INTO mac_data (data_id, payload, time) VALUES (%s, %s, %s)"
        
        # 执行 SQL
        cur.execute(sql, (data_id, str(value), time_str))
        
        # 提交事务
        conn.commit()
        print(f"MySQL 写入成功: ID={data_id}, Value={value}")
        
    except Exception as e:
        print(f"MySQL 写入失败: {e}")
        if conn:
            conn.rollback() # 出错时回滚
    finally:
        # 关闭游标和连接
        if cur:
            cur.close()
        if conn:
            conn.close()

def save_to_influxdb(data_id, value):
    """
    将数据保存到 InfluxDB
    """
    try:
        # 转换为 float，InfluxDB 推荐使用数值类型
        value_float = float(value)
        
        # 创建一个数据点 (Point)
        p = Point("machine_data") \
            .tag("device_id", "STRESS_TEST_00000") \
            .tag("data_id", data_id) \
            .field("value", value_float) \
            .time(datetime.datetime.utcnow(), WritePrecision.NS)
            
        # 写入数据
        write_api.write(bucket=INFLUX_BUCKET, org=INFLUX_ORG, record=p)
        print(f"InfluxDB 写入成功: ID={data_id}, Value={value_float}")
        
    except Exception as e:
        print(f"InfluxDB 写入失败: {e}")

# --- MQTT 事件回调 ---

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
    # 4. (更新) 当收到消息时，解析、存储、并写入数据库
    try:
        payload_str = message.payload.decode()
        print(f"收到消息 (Topic: {message.topic}): {payload_str}")
        
        data = json.loads(payload_str)
        
        if 'values' in data and isinstance(data['values'], list):
            for item in data['values']:
                if 'id' in item and 'values' in item:
                    data_id = item['id']
                    value = item['values'][0] # 假设我们只关心第一个值
                    
                    # 4.1. 暂存数据 (供前端Echarts使用)
                    latest_data_store[data_id] = value
                    print(f"已更新数据存储: {latest_data_store}")

                    # 4.2. (新增) 写入 MySQL
                    save_to_mysql(data_id, value)
                    
                    # 4.3. (新增) 写入 InfluxDB
                    save_to_influxdb(data_id, value)
            
    except Exception as e:
        print(f"处理消息失败: {e}")

# --- Flask API 路由 (保持不变) ---

@app.route('/connect/', methods=['POST', 'GET'])
def make_connect():
    try:
        data_connect = request.get_json()['data']
        print(f"收到来自前端的连接请求: {data_connect}")
        app.config['MQTT_BROKER_URL'] = data_connect['host']
        app.config['MQTT_BROKER_PORT'] = data_connect['port']
        app.config['MQTT_CLIENT_ID'] = data_connect['clientid']
        if mqtt.client.is_connected():
            mqtt.client.disconnect()
        mqtt.client.username_pw_set(None, None)
        mqtt.client._client_id = data_connect['clientid'].encode()
        mqtt.client.reinitialise()
        mqtt.client.connect(data_connect['host'], data_connect['port'])
        print(f"正在尝试连接到 {data_connect['host']}:{data_connect['port']}...")
        return jsonify({'rc_status': 'success'})
    except Exception as e:
        print(f"连接失败: {e}")
        return jsonify({'rc_status': str(e)})

@app.route('/publish/', methods=['POST'])
def do_publish():
    try:
        data = request.get_json()
        topic = data.get('topic')
        payload = data.get('payload', "{}")
        if not topic:
            return jsonify({'status': 'error', 'message': 'Topic is required'}), 400
        print(f"正在发布消息到 (Topic: {topic}): {payload}")
        mqtt.publish(topic, payload)
        return jsonify({'status': 'published', 'topic': topic})
    except Exception as e:
        print(f"发布失败: {e}")
        return jsonify({'status': 'error', 'message': str(e)}), 500

@app.route('/get_data/', methods=['POST'])
def get_data():
    # 
    # ！！！(修正) 加上这一行，告诉 Python 我们要用的是全局的那个 latest_data_store ！！！
    # 
    global latest_data_store 
    
    try:
        data = request.get_json()
        data_id = data.get('id')
        if not data_id:
            return jsonify({'status': 'error', 'message': 'ID is required'}), 400
        
        # 现在这里可以正确地从全局变量中取到数据了
        value = latest_data_store.get(data_id) 
        
        if value is not None:
            return jsonify({'status': 'ok', 'id': data_id, 'value': value})
        else:
            return jsonify({'status': 'not_found', 'id': data_id, 'value': None})
    except Exception as e:
        return jsonify({'status': 'error', 'message': str(e)}), 500

# 启动 Flask 服务
if __name__ == '__main__':
    app.run(debug=True, host='127.0.0.1', port=5000)