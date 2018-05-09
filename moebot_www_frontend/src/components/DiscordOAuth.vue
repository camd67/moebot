<template>
  <div>Please wait...</div>
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
        this.submitError = response.error
      }
    )
  }
}
</script>

<style>

</style>
