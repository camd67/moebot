import Vue from 'vue'
import Router from 'vue-router'
import Login from '../components/Login.vue'
import DiscordOAuth from '../components/DiscordOAuth.vue'
import Home from '../components/Home.vue'
import Main from '../components/Main.vue'

Vue.use(Router)

function checkAuthenticated () {
  if (localStorage.getItem('jwt') !== null) return true
}

const routes = [
  { path: '/login', component: Login, meta: { requiresAuth: false } },
  { path: '/login/discord', component: DiscordOAuth, meta: { requiresAuth: false } },
  { path: '/login/reset' },
  {
    path: '/',
    component: Home,
    meta: { requiresAuth: true },
    children: [
      { path: '', component: Main, meta: { requiresAuth: true } }
    ]
  }
]

const router = new Router({
  mode: 'history',
  routes
})

router.beforeEach((to, from, next) => {
  if (to.meta == null || to.meta.requiresAuth == null) {
    next()
    return
  }
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
