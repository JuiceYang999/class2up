# 搭建系统框架 (VUE)

将其从一个基础模板修改为符合我们实验要求的“大数据采集与分析系统”框架。

您当前的文件是使用`vue-cli`创建的一个标准Vue 2项目。

### 1\. “解读”——分析现有代码

  * `package.json`: 定义了项目依赖，核心是`vue: "^2.6.14"`。
  * `src/main.js`: 项目的入口文件。它目前只加载了`App.vue`组件。
  * `src/App.vue`: 项目的根组件。它目前显示一个Logo和`HelloWorld`组件。
  * `src/components/HelloWorld.vue`: 默认的欢迎页面。
  * `vue.config.js`: Vue CLI的配置文件，目前是空的。

### 2\. “调试”——修改代码以搭建框架

“调试”在这里意味着我们需要修改和添加代码，以满足《实验步骤.pdf》和《大数据部分项目指导书.pdf》中的要求。

#### 2.1: 安装核心依赖

根据实验指导书，我们需要**页面切换** (`vue-router`)、**UI界面** (`element-ui`) 和**前后端通信** (`axios`)。

请在您的项目根目录（`juiceyang999/class2up/class2up-9730b2fcdf474be54786a1f787cc6b005a26d2fc/`）打开终端，运行以下命令：

```bash
yarn add vue-router@3 element-ui axios
```

*(注意：因为您使用的是Vue 2，所以我们安装 `vue-router@3`)*

#### 2.2: 创建页面组件

在 `src/components/` 文件夹中，除了 `HelloWorld.vue`，请**新建**以下三个文件：

  * `CoreSource.vue` (数据采集界面)
  * `ShowData.vue` (数据显示界面)
  * `DataProcess.vue` (数据分析界面)

为了让它们能显示内容，您可以暂时在每个文件中填入以下基础代码（以`CoreSource.vue`为例）：

```vue
<template>
  <div class="page">
    <h1>数据采集界面</h1>
    </div>
</template>

<script>
export default {
  name: 'CoreSource'
}
</script>
```

#### 2.3: 配置路由 (src/router.js)

《实验步骤.pdf》中提到了 `vue router`。我们需要一个文件来管理页面之间的跳转。

在 `src/` 目录下**新建**一个文件 `router.js`，并填入以下内容：

```javascript
import Vue from 'vue'
import VueRouter from 'vue-router'

// 引入我们的组件
import HelloWorld from './components/HelloWorld.vue'
import CoreSource from './components/CoreSource.vue'
import ShowData from './components/ShowData.vue'
import DataProcess from './components/DataProcess.vue'

Vue.use(VueRouter)

// 定义路由规则
const routes = [
  {
    path: '/',
    name: 'Home',
    component: HelloWorld
  },
  {
    path: '/core',
    name: 'CoreSource',
    component: CoreSource
  },
  {
    path: '/show',
    name: 'ShowData',
    component: ShowData
  },
  {
    path: '/process',
    name: 'DataProcess',
    component: DataProcess
  }
]

// 创建 router 实例
const router = new VueRouter({
  routes // (缩写) 相当于 routes: routes
})

export default router
```

#### 2.4: 挂载路由和UI库 (修改 src/main.js)

[cite\_start]现在我们需要告诉Vue“使用”我们刚刚安装的 `ElementUI` [cite: 1091] 和 `router.js`。

请**修改** `src/main.js` 文件，内容如下：

```javascript
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
```

#### 2.5: 修改主布局 (修改 src/App.vue)

最后，我们修改根组件 `App.vue`，将它从显示 "HelloWorld" 改为显示**导航菜单**和**路由页面**。

请**替换** `src/App.vue` 的全部内容为：

```vue
<template>
  <div id="app">
    <el-menu :router="true" mode="horizontal" default-active="/">
      <el-menu-item index="/">欢迎界面</el-menu-item>
      <el-menu-item index="/core">数据采集</el-menu-item>
      <el-menu-item index="/show">数据显示</el-menu-item>
      <el-menu-item index="/process">数据分析</el-menu-item>
    </el-menu>

    <router-view></router-view>
  </div>
</template>

<script>
export default {
  name: 'App'
}
</script>

<style>
/* 保持原有样式，并移除顶部的60px边距 */
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  /* margin-top: 60px; */ /* 移除或注释掉此行 */
}
</style>
```

#### 2.6: 配置跨域代理 (修改 vue.config.js)

[cite\_start]为了准备后续的“前后端通信”任务，我们需要配置一个代理（proxy）来解决跨域问题，这在《实验步骤.pdf》中也有提及 [cite: 1149, 1151]。

请**修改** `vue.config.js` 文件，内容如下：

```javascript
const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
  transpileDependencies: true,

  // 添加 devServer 配置
  devServer: {
    proxy: {
      // '/api' 是一个自定义的标识，所有以 /api 开头的请求都会被代理
      '/api': {
        target: 'http://127.0.0.1:5000', // 这是我们后端 Python Flask 服务的地址
        changOrigin: true, // 允许跨域
        pathRewrite: {
          '^/api': '' // 在请求时，将 /api 替换为空字符串
        }
      }
    }
  }
})
```

-----

### 3\. 验证框架

完成以上所有修改后，请在终端运行：

```bash
yarn serve
```

[cite\_start]现在访问浏览器（通常是 `http://localhost:8080` 或 `http://localhost:8081` [cite: 901]），您应该能看到一个包含“欢迎界面”、“数据采集”、“数据显示”、“数据分析”的导航栏。点击它们，下方的页面内容会随之切换。

这标志着**任务1：搭建系统框架**已成功完成。

接下来，我们将开始**任务2：运行虚拟数据源 (MQTT)**。请确认您是否已准备好 `emqx` 和 `StressTest-INC-Cloud` 软件。