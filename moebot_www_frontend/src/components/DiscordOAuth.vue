<template>
  <v-content>
    <v-container fluid fill-height>
      <v-layout align-center justify-center>
        <v-flex  xs12 sm8 md4>
          <v-card elevation-12>
            <v-card-text>
              <div class="headline">
                <v-progress-circular indeterminate color="primary"></v-progress-circular>
                Please wait...
              </div>
            </v-card-text>
          </v-card>
        </v-flex>
      </v-layout>
    </v-container>
  </v-content>
</template>

<script>
export default {
  created: function () {
    this.$http.post('/auth/discordOAuth', {code: this.$route.query.code, state: this.$route.query.state}).then(
      response => {
        if (response.data && response.data.Jwt) {
          localStorage.setItem('jwt', response.data.Jwt)
          this.$router.push('/')
        }
      },
      response => {
        this.$router.push('/')
      }
    )
  }
}
</script>

<style>

</style>
