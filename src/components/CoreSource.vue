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
                <el-button @click="onDisconnect">断开</el-button>
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
              <span>订阅与发布</span>
            </div>
            <el-form label-width="120px">
              <el-form-item label="订阅主题">
                <el-input placeholder="例如: Probe/Query/Response/STRESS_TEST_00000"></el-input>
              </el-form-item>
              <el-form-item>
                <el-button type="success">订阅</el-button>
              </el-form-item>
              <el-form-item label="收到的消息">
                <el-input
                  type="textarea"
                  :rows="5"
                  placeholder="等待消息..."
                  readonly>
                </el-input>
              </el-form-item>
            </el-form>
          </el-card>
        </el-col>
      </el-row>
    </el-main>
  </el-container>
</template>

<script>
// 1. 引入 axios 用于前后端通信
import axios from "axios";

// 2. 设置 axios 的基础URL，所有请求都会自动加上 /api 前缀
// 这对应《实验步骤.pdf》P9的 "axios.defaults.baseURL = "/api";"
// 我们在 vue.config.js 中配置了 /api 代理
axios.defaults.baseURL = "/api";

export default {
  name: "CoreSource",
  data() {
    return {
      // 对应《实验步骤.pdf》P10 "前端定义输入变量: connection"
      connection: {
        host: "127.0.0.1",
        port: 1883,
        clientid: "vue_client_" + Math.random().toString(16).substr(2, 8),
      },
      connectStatus: false, // 用于显示连接状态
    };
  },
  methods: {
    // 3. onConnect 函数，对应《实验步骤.pdf》P10
    onConnect() {
      console.log("正在连接到MQTT服务器...");
      // 使用 axios.post 发送数据到后端
      // '/connect/' 对应后端 Flask 的路由
      // {data: this.connection} 对应后端 request.get_json()['data']
      axios
        .post("/connect/", { data: this.connection })
        .then((res) => {
          // 这里的 res.data 对应后端 `return {'rc_status': ...}`
          console.log("后端返回的数据:", res.data);
          // 我们用 rc_status 的内容来判断是否成功
          if (res.data && res.data.rc_status === "success") {
            this.connectStatus = true;
            console.log("连接成功！");
            this.$message({
              message: 'MQTT服务器连接成功！',
              type: 'success'
            });
          } else {
            this.connectStatus = false;
            this.$message.error('连接失败: ' + res.data.rc_status);
          }
        })
        .catch((err) => {
          // 如果后端服务没启动或出错，会在这里捕获
          console.error("连接请求失败:", err);
          this.connectStatus = false;
          this.$message.error('后端服务连接失败，请检查Python后端是否已运行。');
        });
    },
    onDisconnect() {
      // (功能待定，目前仅用于前端演示)
      console.log("断开连接");
      this.connectStatus = false;
      this.$message({
        message: '已断开连接（前端模拟）',
        type: 'info'
      });
      // 实际开发中，这里也应该向后端发送一个 /disconnect/ 请求
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