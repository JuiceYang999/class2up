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

            // 9.3. (已修正) 更新 Echarts 图表
            this.chart.setOption({
              xAxis: {
                data: xAxisData,
              },
              series: [
                {
                  data: this.chartData,
                  type: 'line',    // <--- ！！！(修正) 加上这一行
                  smooth: true   // <--- ！！！(修正) 加上这一行
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