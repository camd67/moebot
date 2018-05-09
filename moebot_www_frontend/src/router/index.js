import Vue from 'vue'
import Router from 'vue-router'
import Login from '../components/Login.vue'
import DiscordOAuth from '../components/DiscordOAuth.vue'
import Home from '../components/Home.vue'

Vue.use(Router)

function checkAuthenticated () {
  if (localStorage.getItem('jwt') !== null) return true
}

const routes = [
  { path: '/login', component: Login },
  { path: '/login/discord', component: DiscordOAuth },
  { path: '/', component: Home, meta: { requiresAuth: true } }
]

const router = new Router({
  mode: 'history',
  routes
})

router.beforeEach((to, from, next) => {
  if (to.meta.requiresAuth) { // check the meta field
    if (checkAuthenticated()) { // check if the user is authenticated
      next() // the next method allow the user to continue to the router
    } else {
      next('/login') // Redirect the user to the login page
    }
  } else {
    if (checkAuthenticated()) { // Prevents authenticated users to access the login page
      next('/')
    } else {
      next()
    }
  }
})

export default router
