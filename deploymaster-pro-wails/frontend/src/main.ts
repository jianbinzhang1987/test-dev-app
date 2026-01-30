import { createApp } from 'vue'
import App from './App.vue'
import Antd from 'ant-design-vue'
import AntdX from 'ant-design-x-vue'
import 'ant-design-vue/dist/reset.css'
import './style.css'
// import './assets/all.min.css'

const app = createApp(App)
app.use(Antd)
app.use(AntdX)
app.mount('#app')
