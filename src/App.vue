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