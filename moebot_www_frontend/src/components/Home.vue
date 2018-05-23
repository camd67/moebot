<template>
  <div>
    <v-navigation-drawer app></v-navigation-drawer>
    <v-toolbar app>
      <v-spacer/>
      <v-menu>
        <v-btn slot="activator" icon>
          <v-avatar size="32px">
            <img :src="user.avatar"/>
          </v-avatar>
        </v-btn>
        <v-list>
          <v-list-tile @click="logout">
            <v-list-tile-title>Logout</v-list-tile-title>
          </v-list-tile>
        </v-list>
      </v-menu>
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
export default {
  data () {
    return {
      user: {
        username: '',
        avatar: '/static/baseline_person_black_18dp.png'
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
  }
}
</script>

<style>

</style>
