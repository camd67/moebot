<template>
  <div>
    <v-navigation-drawer app mini-variant>
      <v-toolbar flat class="transparent">
        <UserAvatar/>
      </v-toolbar>
      <ServerList/>
    </v-navigation-drawer>
    <v-toolbar dark color="primary" app>
      <v-toolbar-title>MoeBot Website</v-toolbar-title>
      <v-spacer/>
      <v-btn icon @click="logout">
        <v-icon>exit_to_app</v-icon>
      </v-btn>
    </v-toolbar>
    <v-content>
      <v-container fluid>
        <v-layout row wrap>
          <router-view></router-view>
        </v-layout>
      </v-container>
    </v-content>
  </div>
</template>

<script>
import ServerList from './ServerList.vue'
import UserAvatar from './UserAvatar.vue'
export default {
  created: function () {
    this.heartbeatId = setInterval(this.heartbeat, 60000 * 5)
    this.heartbeat()
  },
  data () {
    return {
      heartbeatId: 0
    }
  },
  methods: {
    logout: function (event) {
      localStorage.removeItem('jwt')
      this.$router.go(this.$router.currentRoute)
    },
    heartbeat: function () {
      this.$http.get('/api/heartbeat', {headers: {'Authorization': 'Bearer ' + localStorage.getItem('jwt')}}).then(
        response => {
          if (response.data && response.data.Jwt) {
            localStorage.setItem('jwt', response.data.Jwt)
          }
        },
        response => {
          clearInterval(this.heartbeatId)
          localStorage.removeItem('jwt')
          this.$router.go()
        }
      )
    }
  },
  components: {
    ServerList,
    UserAvatar
  }
}
</script>

<style>

</style>
