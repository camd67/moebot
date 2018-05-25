<template>
  <div>
    <v-navigation-drawer app mini-variant>
      <v-list class="pa-0">
        <v-list-tile avatar>
          <v-list-tile-avatar>
            <img :src="user.avatar" >
          </v-list-tile-avatar>
        </v-list-tile>
      </v-list>
      <ServerList/>
    </v-navigation-drawer>
    <v-toolbar dark color="primary" app>
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
export default {
  data () {
    return {
      user: {
        username: '',
        avatar: '/static/defaultDiscordAvatar.png'
      }
    }
  },
  created: function () {
    this.$http.get('/api/user', {headers: {'Authorization': 'Bearer ' + localStorage.getItem('jwt')}}).then(
      response => {
        this.user = response.data
      },
      response => {}
    )
  },
  methods: {
    logout: function (event) {
      localStorage.removeItem('jwt')
      this.$router.go(this.$router.currentRoute)
    }
  },
  components: {
    ServerList
  }
}
</script>

<style>

</style>
