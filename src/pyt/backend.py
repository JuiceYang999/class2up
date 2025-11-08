from flask import Flask, request, jsonify
from flask_cors import CORS
from flask_mqtt import Mqtt
import time

# 1. 初始化 Flask 应用
app = Flask(__name__)
# 允许所有域名的跨域请求（在开发中很有用）
CORS(app)

# 2. 配置 Flask-MQTT
# 我们先设置默认值，稍后会用前端传来的数据覆盖它们
app.config['MQTT_BROKER_URL'] = '127.0.0.1'  # 默认服务器地址
app.config['MQTT_BROKER_PORT'] = 1883       # 默认端口
app.config['MQTT_CLIENT_ID'] = 'flask_mqtt_client'
app.config['MQTT_KEEPALIVE'] = 60
app.config['MQTT_TLS_ENABLED'] = False

# 3. 初始化 MQTT 客户端
mqtt = Mqtt(app)

# --- MQTT 事件回调 ---
# 当连接成功时
@mqtt.on_connect()
def handle_connect(client, userdata, flags, rc):
    if rc == 0:
        print("MQTT 连接成功 (rc=0)")
    else:
        print(f"MQTT 连接失败，返回码: {rc}")

# 当收到消息时 (我们将在下一步任务中用到)
@mqtt.on_message()
def handle_message(client, userdata, message):
    data = dict(
        topic=message.topic,
        payload=message.payload.decode()
    )
    print(f"收到消息: {data}")

# --- Flask API 路由 ---

# 4. 创建 /connect/ 接口，对应《实验步骤.pdf》P10
@app.route('/connect/', methods=['POST', 'GET'])
def make_connect():
    try:
        # 5. 接收前端发来的JSON数据
        # 对应 axios.post 中的 {data: this.connection}
        data_connect = request.get_json()['data']
        
        # 6. 在后端控制台打印收到的数据
        print(f"收到来自前端的连接请求: {data_connect}")

        # 7. 更新 MQTT 配置并尝试连接
        app.config['MQTT_BROKER_URL'] = data_connect['host']
        app.config['MQTT_BROKER_PORT'] = data_connect['port']
        app.config['MQTT_CLIENT_ID'] = data_connect['clientid']
        
        # 重新配置并连接
        mqtt.client.disconnect() # 先断开旧连接
        mqtt.client._client_id = data_connect['clientid'].encode()
        mqtt.client.reinitialise()
        mqtt.client.connect(data_connect['host'], data_connect['port'])

        print(f"正在尝试连接到 {data_connect['host']}:{data_connect['port']}...")
        
        # 8. 向前端返回成功响应
        # 对应 axios.then(res => ...)
        # 我们返回 'success' 字符串，前端会据此更新 "连接状态"
        return jsonify({'rc_status': 'success'})

    except Exception as e:
        print(f"连接失败: {e}")
        # 如果出错，返回错误信息
        return jsonify({'rc_status': str(e)})

# (我们将在任务4中添加 /subscribe/ 接口)

# 9. 启动 Flask 服务
if __name__ == '__main__':
    # 确保端口为 5000，以匹配 vue.config.js 中的代理设置
    app.run(debug=True, host='127.0.0.1', port=5000)