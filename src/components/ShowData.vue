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
// 引入 Echarts
import * as echarts from "echarts";

axios.defaults.baseURL = "/api";

export default {
  name: "ShowData",
  data() {
    return {
      chart: null, // Echarts 实例
      chartData: [], // 存储图表的数据
      dataTimer: null, // 存储 setInterval 的ID
      dataIdToFetch: '0103502202', // X轴实际速度
      publishConfig: {
        topic: "Query/Request/STRESS_TEST_00000",
        payload: '{"ids":[{"id":"0103502202"}]}'
      },
    };
  },
  mounted() {
    // 页面加载时：初始化图表并开始抓取数据
    this.publishConfig.payload = `{"ids":[{"id":"${this.dataIdToFetch}"}]}`;
    this.initChart();
    this.startFetching();
  },
  beforeDestroy() {
    // 页面销毁时：停止定时器
    if (this.dataTimer) {
      clearInterval(this.dataTimer);
    }
  },
  methods: {
    // 6. 初始化图表
    initChart() {
      const chartDom = document.getElementById("data-chart");
      this.chart = echarts.init(chartDom);
      
      const option = {
        xAxis: {
          type: "category",
          data: [], 
        },
        yAxis: {
          type: "value",
          scale: true 
        },
        series: [
          {
            data: [], 
            type: "line",    // <--- ！！！(修正) 在初始化时就指定好
            smooth: true,  // <--- ！！！(修正) 在初始化时就指定好
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
      this.requestAndFetch(); // 立即执行一次
      this.dataTimer = setInterval(this.requestAndFetch, 3000); // 3秒
    },

    // 8. 组合方法：先请求，再获取
    requestAndFetch() {
      axios.post("/publish/", this.publishConfig)
        .catch(err => {
          console.error("发布数据请求失败:", err);
        });

      setTimeout(() => {
        this.fetchData();
      }, 500); // (给MQTT一个来回的时间)
    },

    // 9. 从后端获取数据并更新图表
    fetchData() {
      axios.post("/get_data/", { id: this.dataIdToFetch })
        .then(res => {
          if (res.data && res.data.status === 'ok' && res.data.value !== null) {
            const newValue = parseFloat(res.data.value); 
            
            if (this.chartData.length > 50) {
              this.chartData.shift(); 
            }
            this.chartData.push(newValue);
            
            const xAxisData = this.chartData.map((_, index) => index);

            // 9.3. (已修正) 更新 Echarts 图表
            this.chart.setOption({
              xAxis: {
                data: xAxisData,
              },
              series: [
                {
                  data: this.chartData,
                  type: 'line',    // <--- ！！！(修正) 每次更新都必须再带上！
                  smooth: true   // <--- ！！！(修正) 每次更新都必须再带上！
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