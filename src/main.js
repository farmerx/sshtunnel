// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from 'vue'
import App from './App'
import router from './router'
import store from './store'

// 使用element ui
import 'element-ui/lib/theme-default/index.css'
import Element from 'element-ui'
Vue.use(Element)

// 使用vue-cookie
import VueCookie from 'vue-cookie'
Vue.use(VueCookie)

Vue.config.productionTip = false

/* eslint-disable no-new */
new Vue({
  el: '#app',
  router,
  store: store,
  template: '<App/>',
  components: { App }
})

