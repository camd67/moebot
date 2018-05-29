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
  methods: {
    logout: function (event) {
      localStorage.removeItem('jwt')
      this.$router.go(this.$router.currentRoute)
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
