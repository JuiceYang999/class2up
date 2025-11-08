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