import Vue from 'vue'
import App from './App.vue'

// 1. 引入 Element UI
import ElementUI from 'element-ui';
import 'element-ui/lib/theme-chalk/index.css';

// 2. 引入路由
import router from './router'

Vue.config.productionTip = false

// 3. 使用插件
Vue.use(ElementUI);

new Vue({
  // 4. 挂载 router
  router,
  render: h => h(App),
}).$mount('#app')