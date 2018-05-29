<template>
    <v-list class="pa-0">
          <v-list-tile avatar>
            <v-list-tile-avatar>
              <v-progress-circular v-if="isLoading" indeterminate color="primary"></v-progress-circular>
              <img v-else :src="user.avatar" >
            </v-list-tile-avatar>
          </v-list-tile>
        </v-list>
</template>

<script>
export default {
  data () {
    return {
      isLoading: false,
      user: {
        username: '',
        avatar: '/static/defaultDiscordAvatar.png'
      }
    }
  },
  mounted: function () {
    this.isLoading = true
    this.$http.get('/api/user', {headers: {'Authorization': 'Bearer ' + localStorage.getItem('jwt')}}).then(
      response => {
        this.isLoading = false
        this.user = response.data
      },
      response => {
        this.isLoading = false
      }
    )
  }
}
</script>

<style>

</style>
