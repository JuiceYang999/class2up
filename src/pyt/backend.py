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