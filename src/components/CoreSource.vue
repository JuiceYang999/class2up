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