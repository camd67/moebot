import Vue from 'vue'
import VueResource from 'vue-resource'
import App from './App'
import router from './router'
import Vuetify from 'vuetify'
import 'vuetify/dist/vuetify.min.css'
import colors from 'vuetify/es5/util/colors'

Vue.use(Vuetify, {
  theme: {
    primary: colors.purple.lighten1,
    secondary: colors.purple.lighten3,
    accent: colors.purple.accent2,
    error: colors.red.accent3
  }
})
Vue.use(VueResource)

new Vue({
  el: '#app',
  router,
  components: { App },
  template: '<App/>'
}).$mount('#app')
