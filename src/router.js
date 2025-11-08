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